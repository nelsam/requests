package requests

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/nelsam/requests/options"
)

var (
	registeredOptions = map[string]OptionFunc{
		"required":  OptionFunc(options.Required),
		"default":   OptionFunc(options.Default),
		"immutable": OptionFunc(options.Immutable),
	}

	optionDefaults map[string]string
	fallbackTags   []string
)

// FallbackTags returns a list of tag names that are used as fallbacks
// to locate field names.  See AddFallbackTag for more details.
func FallbackTags() []string {
	if fallbackTags == nil {
		fallbackTags = make([]string, 0, 5)
	}
	return fallbackTags
}

// AddFallbackTag adds a tag name to the list of fallback tag names to
// use when locating a field name.  For example, if you use the "db"
// tag and find yourself duplicating the name in both the "db" and
// "request" tags as follows:
//
//     type Example struct {
//         User `db:"user_id" request:"user_id"`
//     }
//
// Then you could update your code to the following:
//
//     func init() {
//         AddFallbackTag("db")
//     }
//
//     type Example struct {
//         User `db:"user_id"`
//     }
//
// Fallback tags will only be added once - duplicates will be ignored.
// Fallbacks will be used in the order that they are added, so if you
// add "db" before "response" (as an example), the "db" tag will be
// preferred over the "response" tag as a fallback.
//
// In all cases, the "request" tag will be preferred over anything
// else.  However, an empty name (but non-empty options) will cause
// the fallbacks to be used, as in the following example:
//
//     func init() {
//         AddFallbackTag("db")
//     }
//
//     type Example struct {
//         // Use the name from the DB, but add the 'required' option
//         // for requests.
//         User `db:"user_id" request:",required"`
//     }
func AddFallbackTag(newTag string) {
	tags := FallbackTags()
	for _, tag := range tags {
		if tag == newTag {
			return
		}
	}
	fallbackTags = append(tags, newTag)
}

// OptionDefaults returns a list of default values for options.  See
// SetOptionDefault for more details.
func OptionDefaults() map[string]string {
	if optionDefaults == nil {
		optionDefaults = make(map[string]string)
	}
	return optionDefaults
}

// SetOptionDefault sets an option to default to a value.  By default,
// OptionFuncs that are excluded from a field's options are never
// called, but if you set a default for that option, it will always be
// called with the default value if the option is not provided on the
// struct field.  An example:
//
//     func init() {
//         SetOptionDefault("required", "true")
//     }
//
//     type Example struct {
//         // user_id will be required in the request
//         User `db:user_id`
//
//         // description will be optional in the request
//         Description `request:",required=false"`
//     }
func SetOptionDefault(option, value string) {
	OptionDefaults()[option] = value
}

// An OptionFunc is a function which takes a field's original value,
// new value (from a request), and a string of option values (anything
// to the right of the = sign in a field's tag option), and returns
// the final new value (parsed from the request value) and any errors encountered.
type OptionFunc func(originalValue, requestValue interface{}, optionValue string) (convertedValue interface{}, err error)

type tagOption struct {
	name  string
	value string
}

func (option tagOption) function() OptionFunc {
	function, _ := registeredOptions[option.name]
	return function
}

// RegisterOption can be used to register functions that should be
// called for struct fields with matching option strings.  For
// example:
//
//     RegisterOption("allow-empty", func(value interface{}, optionValue string) (interface{}, error) {
//         if optionValue == "false" {
//             if value == nil || value.(string) == "" {
//                 return nil, errors.New("Cannot be empty")
//             }
//         }
//         return value, nil
//     })
//
//     type Example struct {
//         Name `request:",allow-empty=false"`
//     }
//
// Any options without a value (e.g. request:",non-empty") will have
// their value set to "true".
//
// An error will be returned if an OptionFunc is already registered
// for the provided name.
func RegisterOption(name string, optionFunc OptionFunc) error {
	if _, ok := registeredOptions[name]; ok {
		return fmt.Errorf(`An OptionFunc is already registered for "%s"`, name)
	}
	registeredOptions[name] = optionFunc
	return nil
}

// ApplyOptions will attempt to apply any OptionFunc values registered
// with RegisterOption to a struct field.  If any of the OptionFuncs
// return an error, the process will immediately return a nil value
// and the returned error.
func ApplyOptions(field reflect.StructField, orig, input interface{}) (value interface{}, optionErr error) {
	value = input
	for _, option := range tagOptions(field) {
		optionFunc := option.function()
		if optionFunc == nil {
			return nil, fmt.Errorf(`Could not find a registered OptionFunc for option "%s"`, option.name)
		}
		if value, optionErr = optionFunc(orig, value, option.value); optionErr != nil {
			return nil, optionErr
		}
	}
	return
}

func tagOptions(field reflect.StructField) []*tagOption {
	remaining := field.Tag.Get("request")
	options := make([]*tagOption, 0, len(optionDefaults)+5)
	for startIdx := strings.IndexRune(remaining, ','); startIdx >= 0; startIdx = strings.IndexRune(remaining, ',') {
		// Skip over the ',' character
		startIdx++
		endIdx := strings.IndexRune(remaining[startIdx:], ',')
		if endIdx < 0 {
			endIdx = len(remaining)
		} else {
			endIdx += startIdx
		}
		currentOption := remaining[startIdx:endIdx]
		remaining = remaining[endIdx:]

		optionNameEnd := strings.IndexRune(currentOption, '=')
		optionNameStart := optionNameEnd + 1
		if optionNameEnd < 0 {
			optionNameEnd = len(currentOption)
			optionNameStart = len(currentOption)
		}

		option := new(tagOption)
		option.name = currentOption[:optionNameEnd]
		option.value = currentOption[optionNameStart:]
		if option.value == "" {
			option.value = "true"
		}
		options = append(options, option)
	}

	for name, defaultValue := range optionDefaults {
		useDefault := true
		for _, option := range options {
			if name == option.name {
				useDefault = false
				break
			}
		}
		if useDefault {
			options = append(options, &tagOption{
				name:  name,
				value: defaultValue,
			})
		}
	}
	return options
}

func name(field reflect.StructField) string {
	name := nameFromTag(field, "request")
	for fallback := 0; len(name) == 0 && fallback < len(fallbackTags); fallback++ {
		name = nameFromTag(field, fallbackTags[fallback])
	}
	if len(name) == 0 {
		name = strings.ToLower(field.Name)
	}
	return name
}

func nameFromTag(field reflect.StructField, tagName string) string {
	name := field.Tag.Get(tagName)
	nameEnd := strings.IndexRune(name, ',')
	if nameEnd == -1 {
		return name
	}
	return name[:nameEnd]
}
