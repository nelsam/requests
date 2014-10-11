package requests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCodecs(t *testing.T) {
	codecService = nil
	assert.NotNil(t, Codecs())

	s := Codecs()
	codecService = nil
	SetCodecs(s)
	assert.Equal(t, s, Codecs())
}
