package requests

import (
	"bytes"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBody_UrlEncoded(t *testing.T) {
	assert := assert.New(t)
	body := bytes.NewBufferString(`test=1&test=2&foo=bar`)
	httpRequest, err := http.NewRequest("POST", "/", body)
	httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	require.NoError(t, err)
	bodyInter, err := New(httpRequest).Body()
	require.NoError(t, err)
	bodyParams, ok := bodyInter.(url.Values)
	require.True(t, ok)
	assert.Equal(2, len(bodyParams["test"]))
	assert.Equal(1, len(bodyParams["foo"]))
}

func TestBody_JSON(t *testing.T) {
	assert := assert.New(t)
	body := bytes.NewBufferString(`{"test":["1", "2"],"foo":"bar"}`)
	httpRequest, err := http.NewRequest("POST", "/", body)
	httpRequest.Header.Set("Content-Type", "application/json")
	require.NoError(t, err)
	bodyInter, err := New(httpRequest).Body()
	require.NoError(t, err)
	bodyParams, ok := bodyInter.(map[string]interface{})
	require.True(t, ok)
	tests, ok := bodyParams["test"].([]interface{})
	if assert.True(ok) {
		assert.Equal(2, len(tests))
	}
	foo, ok := bodyParams["foo"].(string)
	if assert.True(ok) {
		assert.Equal("bar", foo)
	}
}

func TestParams_UrlEncoded(t *testing.T) {
	assert := assert.New(t)
	body := bytes.NewBufferString(`test=1&test=2&foo=bar`)
	httpRequest, err := http.NewRequest("POST", "/", body)
	httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	require.NoError(t, err)
	bodyParams, err := New(httpRequest).Params()
	require.NoError(t, err)

	expected := map[string]interface{}{
		"foo":  "bar",
		"test": []interface{}{"1", "2"}}
	assert.Equal(expected, bodyParams)
}

func TestParams_JSON(t *testing.T) {
	assert := assert.New(t)
	body := bytes.NewBufferString(`{"test":["1", "2"],"foo":"bar"}`)
	httpRequest, err := http.NewRequest("POST", "/", body)
	httpRequest.Header.Set("Content-Type", "application/json")
	require.NoError(t, err)
	bodyParams, err := New(httpRequest).Params()
	require.NoError(t, err)

	expected := map[string]interface{}{
		"foo":  "bar",
		"test": []interface{}{"1", "2"}}
	assert.Equal(expected, bodyParams)
}

func TestParams_BothForms(t *testing.T) {
	assert := assert.New(t)
	body := bytes.NewBufferString(`test=1&test=2&foo=bar`)
	httpRequest, err := http.NewRequest("POST", "/", body)
	httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	require.NoError(t, err)
	UrlEncodedParams, err := New(httpRequest).Params()
	require.NoError(t, err)

	body = bytes.NewBufferString(`{"test":["1", "2"],"foo":"bar"}`)
	httpRequest, err = http.NewRequest("POST", "/", body)
	httpRequest.Header.Set("Content-Type", "application/json")
	require.NoError(t, err)
	JSONParams, err := New(httpRequest).Params()
	require.NoError(t, err)

	assert.Equal(UrlEncodedParams, JSONParams)
}
