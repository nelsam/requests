package requests

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/stretchr/codecs/services"
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

// UnmarshalReplace performs the same process as Unmarshal, except
// that values not found in the request will be updated to their zero
// value.  For example, if foo.Bar == "baz" and foo.Bar has no
// corresponding data in a request, Unmarshal would leave it as "baz",
// but UnmarshalReplace will update it to "".
//
// Exceptions are made for unexported fields and fields which are
// found to have a name of "-".  Those are left alone.
func (request *Request) UnmarshalReplace(target interface{}) error {
	return request.unmarshal(target, true)
}

// Unmarshal unmarshals a request to a struct, using field tags to
// locate corresponding values in the request and check/parse them
// before assigning them to struct fields.  It acts similar to json's
// Unmarshal when used on a struct, but works with any codec
// registered with AddCodec().
//
// Field tags are used as follows:
//
// * All field tags are considered to be of the format
// name,option1,option2,...
//
// * Options will *only* be parsed from the "request" tag.
//
// * By default, name will only be checked in the "request" tag, but
// you can add fallback tag names using AddFallbackTag.
//
// * If no non-empty name is found using field tags, the lowercase
// field name will be used instead.
//
// * Once a name is found, if the name is "-", then the field will be
// treated as if it does not exist.
//
// For an explanation on how options work, see the documentation for
// RegisterOption.  For a list of tag options built in to this
// library, see the options package in this package.
//
// Fields which have no data in the request will be left as their
// current value.  They will still be passed through the option parser
// for the purposes of options like "required".
//
// Fields which implement Receiver will have their Receive method
// called using the value from the request after calling all
// OptionFuncs matching the field's tag options.
//
// An error will be returned if the target type is not a pointer to a
// struct, or if the target implements PreUnmarshaller, Unmarshaller,
// or PostUnmarshaller and the corresponding methods fail.  An
// UnusedFields error will be returned if fields in the request had no
// corresponding fields on the target struct.
//
// Panics from an Unmarshaller's Unmarshal method will be recovered and
// returned as error types.  The error will be the same error as
// returned by request.Params() if the error returned from
// request.Params() is of type *services.ContentTypeNotSupportedError,
// or a generic error with the recover output otherwise.
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
func (request *Request) Unmarshal(target interface{}) error {
	return request.unmarshal(target, false)
}

// unmarshal performes all of the logic for Unmarshal and
// UnmarshalReplace.
func (request *Request) unmarshal(target interface{}, replace bool) (unmarshalErr error) {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.Elem().Kind() != reflect.Struct {
		return errors.New("The value passed to Unmarshal must be a pointer to a struct")
	}
	targetValue = targetValue.Elem()

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
	// We allow Unmarshallers to handle types that aren't supported by this
	// library, so wait to return an error from Params() until after we've
	// checked if target is an Unmarshaler.
	params, err := request.Params()
	if unmarshaller, ok := target.(Unmarshaller); ok {
		var body interface{} = params
		if err != nil {
			// Only ignore ContentTypeNotSupportedError
			if _, ok := err.(*services.ContentTypeNotSupportedError); !ok {
				return err
			}
			body = request.httpRequest.Body
		}
		defer func() {
			if r := recover(); r != nil {
				if err != nil {
					unmarshalErr = err
				} else {
					unmarshalErr = fmt.Errorf("Unmarshal panicked: %v", r)
				}
			}
		}()
		return unmarshaller.Unmarshal(body)
	}
	if unmarshalErr != nil {
		return
	}

	matchedFields, inputErrs := unmarshalToValue(params, targetValue, replace)
	if len(inputErrs) > 0 {
		return inputErrs
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
func unmarshalToValue(params map[string]interface{}, targetValue reflect.Value, replace bool) (matchedFields set, parseErrs InputErrors) {
	matchedFields = make(set, 0, len(params))
	parseErrs = make(InputErrors)
	defer func() {
		// Clean up any nil errors from the error map.
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
				embeddedFields, newErrs := unmarshalToValue(params, fieldValue, replace)
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

		name := name(field)
		if name == "-" {
			continue
		}
		setter := fieldValue.Set
		if field.PkgPath == "" {
			// Unexported fields can only be supported if they have setters.
			// We detect these methods following the rules in Effective Go.
			setterName := "Set" + strings.Title(field.Name)
			setterMethod := targetValue.MethodByName(setterName)
			if !setterMethod.IsValid() {
				parseErrs.Set(name, fmt.Errorf("Unexported field %s needs a setter method %s, "+
					"or should be unused in the request (hint: field tag `request:\"-\"`).",
					field.Name, setterName))
			}
			setter = func(val reflect.Value) {
				setterMethod.Call([]reflect.Value{targetValue, val})
			}
		}
		valueInter, fromParams := params[name]
		var value reflect.Value
		if fromParams {
			value = reflect.ValueOf(valueInter)
			matchedFields = matchedFields.add(name)
		} else {
			// If we're not replacing the value, use the field's
			// current value.  If we are, use the field's zero
			// value.
			zero := reflect.Zero(fieldValue.Type())
			if replace {
				value = zero
			} else {
				value = fieldValue
			}
			if value == zero {
				// The value is empty, so see if its default can
				// be loaded.
				if defaulter, ok := fieldValue.Interface().(Defaulter); ok {
					value = reflect.ValueOf(defaulter.DefaultValue())
				}
			}
		}
		var optionValue interface{}
		if value.IsValid() {
			optionValue = value.Interface()
		}
		newVal, inputErr := ApplyOptions(field, fieldValue.Interface(), optionValue)
		if parseErrs.Set(name, inputErr) {
			continue
		}
		value = reflect.ValueOf(newVal)
		parseErrs.Set(name, setValue(fieldValue, value, setter, fromParams))
	}
	return
}

// isNil returns true if value.IsValid() returns false or if
// value.IsNil() returns true.  Returns false otherwise.  Recovers
// panics from value.IsNil().
func isNil(value reflect.Value) bool {
	defer func() {
		recover()
	}()
	if !value.IsValid() {
		return true
	}
	if value.IsNil() {
		return true
	}
	return false
}

// assignNil takes a target and value and handles nil assignment.  If
// value is nil or invalid, target will be assigned nil.  If value is
// non-nil and target is a nil pointer, it will be initialized.
// Returns whether or not value evaluates to nil, and any errors
// encountered while attempting assignment.
func assignNil(target, value reflect.Value, targetSetter func(reflect.Value)) (valueIsNil bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Nil value found, but type %s cannot be nil.", target.Type().Name())
		}
	}()

	if valueIsNil = isNil(value); valueIsNil {
		// target.IsNil() will panic if target's zero value is
		// non-nil.
		if !target.IsNil() {
			targetSetter(reflect.Zero(target.Type()))
		}
		return
	}
	return
}

func callReceivers(target reflect.Value, value interface{}) (receiverFound bool, err error) {
	receiveTyper, hasReceiveType := target.Interface().(ReceiveTyper)
	preReceiver, hasPreReceive := target.Interface().(PreReceiver)
	receiver, hasReceive := target.Interface().(Receiver)
	var (
		changeReceiver   ChangeReceiver
		hasChangeReceive bool
	)
	if !hasReceive {
		changeReceiver, hasChangeReceive = target.Interface().(ChangeReceiver)
	}
	postReceiver, hasPostReceive := target.Interface().(PostReceiver)
	if target.CanAddr() {
		// If interfaces weren't found, try again with the pointer
		targetPtr := target.Addr().Interface()
		if !hasReceiveType {
			receiveTyper, hasReceiveType = targetPtr.(ReceiveTyper)
		}
		if !hasPreReceive {
			preReceiver, hasPreReceive = targetPtr.(PreReceiver)
		}
		if !hasReceive {
			receiver, hasReceive = targetPtr.(Receiver)
		}
		if !hasReceive && !hasChangeReceive {
			changeReceiver, hasChangeReceive = targetPtr.(ChangeReceiver)
		}
		if !hasPostReceive {
			postReceiver, hasPostReceive = targetPtr.(PostReceiver)
		}
	}
	receiverFound = hasReceive || hasChangeReceive

	if hasPreReceive {
		if err = preReceiver.PreReceive(); err != nil {
			return
		}
	}
	if hasPostReceive {
		defer func() {
			if err == nil {
				err = postReceiver.PostReceive()
			}
		}()
	}
	if hasReceive || hasChangeReceive {
		if hasReceiveType {
			valueVal := reflect.ValueOf(value)
			// Make sure the target value is assignable
			targetVal := reflect.New(reflect.TypeOf(receiveTyper.ReceiveType())).Elem()
			assignVal := targetVal
			if !valueVal.Type().ConvertibleTo(targetVal.Type()) && targetVal.Kind() == reflect.Ptr {
				// If the source value isn't convertible to the target type,
				// but the target type is a pointer, try assigning to the
				// target's elem.
				if assignVal.IsNil() {
					assignVal.Set(reflect.New(assignVal.Type().Elem()))
				}
				assignVal = assignVal.Elem()
			}
			if !valueVal.Type().ConvertibleTo(assignVal.Type()) {
				return true, fmt.Errorf("Cannot convert kind %s to target kind %s", valueVal.Kind(), assignVal.Kind())
			}
			assignVal.Set(valueVal.Convert(assignVal.Type()))
			value = targetVal.Interface()
		}
		if hasReceive {
			err = receiver.Receive(value)
		} else {
			_, err = changeReceiver.Receive(value)
		}
	}
	return
}

// setValue takes a target and a value, and updates the target to
// match the value.  targetSetter should be target.Set for any settable
// values, but can perform other logic for situations such as unexported
// fields that have SetX methods on the parent struct.
func setValue(target, value reflect.Value, targetSetter func(reflect.Value), fromRequest bool) (parseErr error) {
	if isNil, err := assignNil(target, value, targetSetter); isNil || err != nil {
		return err
	}
	if target.Kind() == reflect.Ptr && target.IsNil() {
		target = reflect.New(target.Type().Elem())
		targetSetter(target)
	}

	// Only worry about the receive methods if the value is from a
	// request.
	if fromRequest {
		if receiverFound, err := callReceivers(target, value.Interface()); err != nil || receiverFound {
			return err
		}
	}

	if target.Interface() == value.Interface() {
		return nil
	}

	for target.Kind() == reflect.Ptr {
		target = target.Elem()
	}
	switch target.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parseErr = setInt(target, value.Interface())
	case reflect.Float32, reflect.Float64:
		parseErr = setFloat(target, value.Interface())
	default:
		for value.Kind() == reflect.Ptr {
			value = value.Elem()
		}
		inputType := value.Type()
		if !inputType.ConvertibleTo(target.Type()) {
			return fmt.Errorf("Cannot convert value of type %s to type %s",
				inputType.Name(), target.Type().Name())
		}
		targetSetter(value.Convert(target.Type()))
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
