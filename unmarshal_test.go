package requests

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type RequiredTest struct {
	X string `request:",required"`
	Y string `request:"y"`
}

func Test_Required_Returns_Error_When_Missing(t *testing.T) {
	body := bytes.NewBufferString(`{"y":"bar"}`)
	httpRequest, err := http.NewRequest("POST", "/", body)
	require.NoError(t, err)
	httpRequest.Header.Set("Content-Type", "application/json")

	request := New(httpRequest)
	target := new(RequiredTest)

	assert.Error(t, request.Unmarshal(target))
}
