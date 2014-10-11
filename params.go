package requests

import (
	"errors"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
)

var multipartMem int64 = 2 << 20 * 10

// MultipartMem returns the current memory limit for multipart form
// data.
func MultipartMem() int64 {
	return multipartMem
}

// SetMultipartMem sets the memory limit for multipart form data.
func SetMultipartMem(mem int64) {
	multipartMem = mem
}

// Body returns the result of ParseBody for this request.  ParseBody
// will only be called the first time Body is called; subsequent calls
// will return the same value as the first call.
func (request *Request) Body() (interface{}, error) {
	if request.body == nil {
		body, err := ParseBody(request.httpRequest)
		if err != nil {
			return nil, err
		}
		request.body = body
	}
	return request.body, nil
}

// Params returns the result of ParseParams for this request.
// ParseParams will only be called the first time Params is called;
// subsequent calls will return the same value as the first call.
func (request *Request) Params() (map[string]interface{}, error) {
	if request.params == nil {
		body, err := request.Body()
		if err != nil {
			return nil, err
		}
		params, err := convertToParams(body)
		if err != nil {
			return nil, err
		}
		request.params = params
	}
	return request.params, nil
}

// ParseBody locates a codec matching the request's Content-Type
// header, then Unmarshals the request's body to an interface{} type.
// The resulting type is unpredictable and will be heavily based on
// the actual data in the request.
//
// There are two exceptions to the above, where no codec lookup is
// used:
//
// * application/x-www-form-urlencoded (or an empty Content-Type)
//
// ** The return value for this type will be the same as
//    "net/http".Request.PostForm after calling ParseForm.
//
// * multipart/form-data
//
// ** The return value for this type will be the same as
//    "net/http".Request.MultipartForm after calling
//    ParseMultipartForm.
func ParseBody(request *http.Request) (interface{}, error) {
	// Handle form data types
	contentType, _, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}
	switch contentType {
	case "application/x-www-form-urlencoded", "":
		if err := request.ParseForm(); err != nil {
			return nil, err
		}
		return request.PostForm, nil
	case "multipart/form-data":
		if err := request.ParseMultipartForm(MultipartMem()); err != nil {
			return nil, err
		}
		return request.MultipartForm, nil
	}

	// Now the general case
	codec, err := Codecs().GetCodec(contentType)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}
	var response interface{}
	if err = codec.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	return response, nil
}

// ParseParams returns a map[string]interface{} of values found in a
// request body.  In most cases, this is the equivalent of
// ParseBody(request).(map[string]interface{}).  However, there are
// two exceptions:
//
// * application/x-www-form-urlencoded (or an empty Content-Type)
//
// ** Each value in request.PostForm that has a len() of 1 will be
//    stored instead as the zeroeth index of the value.
//
// * multipart/form-data
//
// ** In addition to the above, files will be stored at the same
//    level as values.  Each value in the resulting map could contain
//    both string and *"mime/multipart".FileHeader values.
//
// The resulting code to parse a form may look like the following:
//
//     params, err := ParseParams(request)
//     // handle err
//     for key, value := range params {
//         switch v := value.(type) {
//         case string:
//             // Do stuff with single string value
//         case *multipart.FileHeader:
//             // Do stuff with single file
//         case []interface{}:
//             // There were multiple string and/or file values at
//             // this key, so deal with that.
//         }
//     }
//
func ParseParams(request *http.Request) (map[string]interface{}, error) {
	if request.Body == nil {
		return nil, nil
	}
	body, err := ParseBody(request)
	if err != nil {
		return nil, err
	}
	return convertToParams(body)
}

func convertToParams(body interface{}) (map[string]interface{}, error) {
	switch body.(type) {
	case url.Values, *multipart.Form:
		return convertForm(body), nil
	}
	m, ok := body.(map[string]interface{})
	if !ok {
		return nil, errors.New("The unmarshalled body is not of type map[string]interface{} " +
			"and cannot be converted to params")
	}
	return m, nil
}

func convertForm(body interface{}) map[string]interface{} {
	var (
		params = make(map[string]interface{})
		values map[string][]string
		files  map[string][]*multipart.FileHeader
	)
	switch form := body.(type) {
	case url.Values:
		values = map[string][]string(form)
	case *multipart.Form:
		values = form.Value
		files = form.File
	}
	for key, valueList := range values {
		if len(valueList) == 1 {
			params[key] = valueList[0]
		} else {
			values := make([]interface{}, len(valueList))
			for idx, value := range valueList {
				values[idx] = value
			}
			params[key] = values
		}
	}
	for key, fileList := range files {
		param, ok := params[key]
		if ok || len(fileList) > 1 {
			values := make([]interface{}, 0, len(params)+len(fileList))
			switch params := param.(type) {
			case []interface{}:
				values = append(values, params...)
			default:
				values = append(values, param)
			}
			for _, file := range fileList {
				values = append(values, file)
			}
			param = values
		} else {
			param = fileList[0]
		}
		params[key] = param
	}
	return params
}
