package textsplitter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTextSplitterConfigWithValidName(t *testing.T) {
	_, err := GetTextSplitterConfig("text")
	assert.NoError(t, err)
}

func TestGetTextSplitterConfigWithInvalidName(t *testing.T) {
	_, err := GetTextSplitterConfig("invalid")
	assert.Error(t, err)
}

func TestGetTextSplitterFuncWithValidNameAndNilConfig(t *testing.T) {
	_, err := GetTextSplitter("text", nil)
	assert.NoError(t, err)
}

func TestGetTextSplitterFuncWithValidNameAndInvalidConfig(t *testing.T) {
	_, err := GetTextSplitter("text", "invalid")
	assert.Error(t, err)
}

func TestGetTextSplitterFuncWithValidNameAndValidConfig(t *testing.T) {
	_, err := GetTextSplitter("text", NewTextSplitterOpts())
	assert.NoError(t, err)
}

func TestGetTextSplitterFuncWithInvalidName(t *testing.T) {
	_, err := GetTextSplitter("invalid", nil)
	assert.Error(t, err)
}
