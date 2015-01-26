package requests

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type emote string

func (e *emote) Receive(stoic interface{}) error {
	s := new(string)
	*s = stoic.(string) + "! :D"
	*e = emote(*s)
	return nil
}

type emoteTyper struct {
	emote
}

func (et *emoteTyper) ReceiveType() interface{} {
	return string("")
}

type structType struct{}

type testTarget struct {
	structType
	Foo                 int64
	Bar                 float64
	Baz                 string
	Qux                 *emote
	Quxtyper            *emoteTyper
	Interfacemember     interface{}
	Defaultermember     interface{} `request:"interfacemember,default"`
	Ignoredmember       interface{} `request:"-"`
	structmember        struct{ m int64 }
	unexported          int64
	unexportedinterface interface{}
}

func (tt *testTarget) SetUnexported(i int64) {
	tt.unexported = i
}
func (tt *testTarget) Unexported() int64 {
	return tt.unexported
}

func (tt *testTarget) SetUnexportedinterface(i interface{}) {
	tt.unexportedinterface = i
}
func (tt *testTarget) Unexportedinterface() interface{} {
	return tt.unexportedinterface
}

func (tt *testTarget) SetStructmember(s struct{ m int64 }) {
	tt.structmember = s
}
func (tt *testTarget) Structmember() struct{ m int64 } {
	return tt.structmember
}

type testTargetPtr struct {
	*structType
}

type testErrTarget struct {
	noset struct{}
	noget struct{}
}

func (tt *testErrTarget) Noset() struct{} {
	return tt.noset
}
func (tt *testErrTarget) SetNoget(s struct{}) {
	tt.noget = s
}

type testUnmarshaller struct{}

func (tu *testUnmarshaller) Unmarshal(b interface{}) error {
	return nil
}

func TestUnmarshal_All(t *testing.T) {
	body := bytes.NewBufferString(
		`foo=1&bar=2.7&baz=taz&quxtyper=welcome&interfacemember=2.7&unexported=1`)
	httpRequest, err := http.NewRequest("POST", "/", body)
	httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	require.NoError(t, err)

	target := new(testTarget)
	target.Qux = new(emote)
	target.Qux.Receive("default")

	err = New(httpRequest).Unmarshal(target)
	require.NoError(t, err)

	assert.Equal(t, int64(1), target.Foo)
	assert.Equal(t, float64(2.7), target.Bar)
	assert.Equal(t, "taz", target.Baz)
	assert.Equal(t, "default! :D", string(*target.Qux))
	assert.Equal(t, "welcome! :D", string((*target.Quxtyper).emote))
}

func TestUnmarshal_AnonPtr(t *testing.T) {
	httpRequest, err := http.NewRequest("POST", "/", nil)
	require.NoError(t, err)

	target := new(testTargetPtr)

	err = New(httpRequest).Unmarshal(target)
	require.NoError(t, err)
}

func TestUnmarshalReplace_All(t *testing.T) {
	body := bytes.NewBufferString(`quxtyper=welcome`)
	httpRequest, err := http.NewRequest("POST", "/", body)
	httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	require.NoError(t, err)

	target := new(testTarget)
	target.Foo = 1
	target.Bar = 2.7
	target.Baz = "default"
	target.Qux = new(emote)
	target.Qux.Receive("default")

	err = New(httpRequest).UnmarshalReplace(target)
	require.NoError(t, err)

	assert.Equal(t, int64(0), target.Foo)
	assert.Equal(t, float64(0), target.Bar)
	assert.Equal(t, "", target.Baz)
	assert.Equal(t, (*emote)(nil), target.Qux)
	assert.Equal(t, "welcome! :D", string((*target.Quxtyper).emote))
}

func TestUnmarshaller(t *testing.T) {
	body := bytes.NewBufferString(`foo=1`)
	httpRequest, _ := http.NewRequest("POST", "/", body)
	httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	require.Implements(t, (*Unmarshaller)(nil), new(testUnmarshaller))
	var target1 testUnmarshaller
	err := New(httpRequest).Unmarshal(&target1)
	assert.NoError(t, err)
}

func TestUnmarshal_Errors(t *testing.T) {
	httpRequest, _ := http.NewRequest("POST", "/", nil)
	httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	target1 := new(testErrTarget)
	err := New(httpRequest).Unmarshal(target1)
	assert.Error(t, err)
}
