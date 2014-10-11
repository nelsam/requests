package requests

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

// set is a simple slice of unique strings.
type set []string

// add appends a variadic amount of strings to a set, returning the
// resulting set.  Duplicates will only exist once in the resulting
// set.
func (s set) add(values ...string) set {
	for _, newValue := range values {
		exists := false
		for _, value := range s {
			if newValue == value {
				exists = true
				break
			}
		}
		if !exists {
			s = append(s, newValue)
		}
	}
	return s
}

// UnmarshalParams unmarshals a request to a struct, using field tags
// to locate corresponding values in the request and check/parse them
// before assigning them to struct fields.  It acts similar to json's
// Unmarshal when used on a struct, but works with any codec
// registered with AddCodec().  Field tags are used as follows:
//
// * All field tags are considered to be of the format
// name,option1,option2,...
// * Options will *only* be parsed from the "request" tag.
// * By default, name will only be checked in the "request" tag, but
// you can add fallback tag names using AddFallbackTag.
// * If no non-empty name is found using field tags, the lowercase
// field name will be used instead.
// * Once a name is found, if the name is "-", then the field will be
// treated as if it does not exist.
//
// For an explanation on how options work, see the documentation for
// RegisterOption.  For a list of tag options built in to this
// library, see the options package in this package.
//
// Fields which implement Receiver will have their Receive method
// called using the value from the request after calling all
// OptionFuncs matching the field's tag options.
//
// An error will be returned if the target type is not a pointer to a
// struct, or if the target implements PreUnmarshaller, Unmarshaller,
// or PostUnmarshaller and the corresponding methods fail.
//
// Any errors encountered while attempting to apply input values to
// the target's fields will be stored in an error of type InputErrors.
// At the end of the Unmarshal process, the InputErrors error will be
// returned if any errors were encountered.
//
// A simple example:
//
//     type Example struct {
//         Foo string `request:",required"`
//         Bar string `response:"baz"`
//         Baz string `response:"-"`
//         Bacon string `response:"-" request:"bacon,required"`
//     }
//
//     func CreateExample(request *http.Request) (*Example, error) {
//         target := new(Example)
//         if err := requests.New(request).Unmarshal(target); err != nil {
//             if inputErrs, ok := err.(InputErrors); ok {
//                 // inputErrs is a map of input names to error
//                 // messages, so send them to a function to turn
//                 // them into a proper user-friendly error message.
//                 return nil, userErrors(inputErrs)
//             }
//             return nil, err
//         }
//         return target, nil
//     }
//
func (request *Request) Unmarshal(target interface{}) (unmarshalErr error) {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.Elem().Kind() != reflect.Struct {
		return errors.New("The value passed to Unmarshal must be a pointer to a struct")
	}
	params, err := request.Params()
	if err != nil {
		return err
	}

	if preUnmarshaller, ok := target.(PreUnmarshaller); ok {
		if unmarshalErr = preUnmarshaller.PreUnmarshal(); unmarshalErr != nil {
			return
		}
	}
	if postUnmarshaller, ok := target.(PostUnmarshaller); ok {
		defer func() {
			if unmarshalErr == nil {
				unmarshalErr = postUnmarshaller.PostUnmarshal()
			}
		}()
	}
	if unmarshaller, ok := target.(Unmarshaller); ok {
		return unmarshaller.Unmarshal(params)
	}

	matchedFields, err := unmarshalToValue(params, targetValue)
	if err != nil {
		return err
	}

	unused := &UnusedFields{
		params:  params,
		matched: matchedFields,
	}
	if unused.HasMissing() {
		return unused
	}
	return nil
}

// unmarshalToValue is a helper for UnmarshalParams, which keeps track
// of the total number of fields matched in a request and which fields
// were missing from a request.
func unmarshalToValue(params map[string]interface{}, targetValue reflect.Value) (matchedFields set, parseErrs InputErrors) {
	matchedFields = make(set, 0, len(params))

	parseErrs = make(InputErrors)
	defer func() {
		// After parsing all input, get rid of nil errors in the input
		// error map, and set parseErrs to nil if there are no non-nil
		// errors.
		parseErrs = parseErrs.Errors()
	}()

	targetType := targetValue.Type()
	for i := 0; i < targetValue.NumField(); i++ {
		fieldValue := targetValue.Field(i)
		field := targetType.Field(i)
		if field.Anonymous {
			// Ignore non-struct anonymous fields, but treat fields in
			// struct or struct pointer anonymous fields as if they
			// were fields on the child struct.
			if fieldValue.Kind() == reflect.Ptr {
				fieldValue = fieldValue.Elem()
			}
			if fieldValue.Kind() == reflect.Struct {
				embeddedFields, newErrs := unmarshalToValue(params, fieldValue)
				if newErrs != nil {
					// Override input errors in the anonymous field
					// with input errors in the child.  Non-nil
					// errors from anonymous fields will be
					// overwritten with nil errors from overriding
					// child fields.
					parseErrs = newErrs.Merge(parseErrs)
				}
				matchedFields = matchedFields.add(embeddedFields...)
			}
			continue
		}

		// Skip unexported fields
		if field.PkgPath == "" {
			name := name(field)
			if name == "-" {
				continue
			}

			value, ok := params[name]
			if ok {
				matchedFields = matchedFields.add(name)
			} else if defaulter, ok := fieldValue.Interface().(Defaulter); ok {
				value = defaulter.DefaultValue()
			} else {
				// Here, use the current value of the field, so that
				// options like default and required won't act as if
				// the field is not set when the original value is
				// non-empty.
				value = fieldValue.Interface()
			}
			var inputErr error
			value, inputErr = ApplyOptions(field, fieldValue.Interface(), value)
			if parseErrs.Set(name, inputErr) {
				continue
			}
			parseErrs.Set(name, setValue(fieldValue, value))
		}
	}
	return
}

// setValue takes a target and a value, and updates the target to
// match the value.
func setValue(target reflect.Value, value interface{}) (parseErr error) {
	if value == nil {
		if target.Kind() != reflect.Ptr {
			return errors.New("Cannot set non-pointer value to null")
		}
		if !target.IsNil() {
			target.Set(reflect.Zero(target.Type()))
		}
		return nil
	}

	if target.Kind() == reflect.Ptr && target.IsNil() {
		target.Set(reflect.New(target.Type().Elem()))
	}

	preReceiver, hasPreReceive := target.Interface().(PreReceiver)
	receiver, hasReceive := target.Interface().(Receiver)
	postReceiver, hasPostReceive := target.Interface().(PostReceiver)
	if target.CanAddr() {
		// If interfaces weren't found, try again with the pointer
		targetPtr := target.Addr().Interface()
		if !hasPreReceive {
			preReceiver, hasPreReceive = targetPtr.(PreReceiver)
		}
		if !hasReceive {
			receiver, hasReceive = targetPtr.(Receiver)
		}
		if !hasPostReceive {
			postReceiver, hasPostReceive = targetPtr.(PostReceiver)
		}
	}

	if hasPreReceive {
		if parseErr = preReceiver.PreReceive(); parseErr != nil {
			return
		}
	}
	if hasPostReceive {
		defer func() {
			if parseErr == nil {
				parseErr = postReceiver.PostReceive()
			}
		}()
	}
	if hasReceive {
		return receiver.Receive(value)
	}

	for target.Kind() == reflect.Ptr {
		target = target.Elem()
	}
	switch target.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parseErr = setInt(target, value)
	case reflect.Float32, reflect.Float64:
		parseErr = setFloat(target, value)
	default:
		inputType := reflect.TypeOf(value)
		if !inputType.ConvertibleTo(target.Type()) {
			return fmt.Errorf("Cannot convert value of type %s to type %s",
				inputType.Name(), target.Type().Name())
		}
		target.Set(reflect.ValueOf(value).Convert(target.Type()))
	}
	return
}

func setInt(target reflect.Value, value interface{}) error {
	switch src := value.(type) {
	case string:
		intVal, err := strconv.ParseInt(src, 10, 64)
		if err != nil {
			return err
		}
		target.SetInt(intVal)
	case int:
		target.SetInt(int64(src))
	case int8:
		target.SetInt(int64(src))
	case int16:
		target.SetInt(int64(src))
	case int32:
		target.SetInt(int64(src))
	case int64:
		target.SetInt(src)
	case float32:
		target.SetInt(int64(src))
	case float64:
		target.SetInt(int64(src))
	}
	return nil
}

func setFloat(target reflect.Value, value interface{}) error {
	switch src := value.(type) {
	case string:
		floatVal, err := strconv.ParseFloat(src, 64)
		if err != nil {
			return err
		}
		target.SetFloat(floatVal)
	case int:
		target.SetFloat(float64(src))
	case int8:
		target.SetFloat(float64(src))
	case int16:
		target.SetFloat(float64(src))
	case int32:
		target.SetFloat(float64(src))
	case int64:
		target.SetFloat(float64(src))
	case float32:
		target.SetFloat(float64(src))
	case float64:
		target.SetFloat(src)
	}
	return nil
}
