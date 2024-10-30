package output

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactSensitiveWithNoCustomFields(t *testing.T) {
	original := struct {
		ApiKey   string
		Username string
	}{"secret", "user"}

	redacted := RedactSensitive(original)

	v := reflect.ValueOf(redacted)

	assert.Equal(t, "REDACTED", v.FieldByName("ApiKey").String())
	assert.Equal(t, "user", v.FieldByName("Username").String())
}

func TestRedactSensitiveWithCustomFields(t *testing.T) {
	original := struct {
		Foo  string
		Spam string
		Bar  string
	}{"some", "weird", "fields"}

	redacted := RedactSensitive(original, "spam")

	v := reflect.ValueOf(redacted)

	assert.Equal(t, "some", v.FieldByName("Foo").String())
	assert.Equal(t, "REDACTED", v.FieldByName("Spam").String())
	assert.Equal(t, "fields", v.FieldByName("Bar").String())
}

func TestRedactSensitiveWithUnsupportedType(t *testing.T) {
	original := "string"
	redacted := RedactSensitive(original)
	assert.Equal(t, original, redacted)
}

func TestRedactSensitiveWithPointerToStruct(t *testing.T) {
	original := &struct {
		Token string
		Info  string
	}{"token", "info"}

	redacted := RedactSensitive(original)

	v := reflect.ValueOf(redacted)

	assert.Equal(t, "REDACTED", v.FieldByName("Token").String())
	assert.Equal(t, "info", v.FieldByName("Info").String())
}

func TestRedactSensitiveWithNestedStruct(t *testing.T) {
	original := struct {
		Credentials struct {
			Password string
		}
		Detail string
	}{struct{ Password string }{"pass"}, "detail"}

	redacted := RedactSensitive(original)

	v := reflect.ValueOf(redacted)

	assert.Equal(t, "REDACTED", v.FieldByName("Credentials").FieldByName("Password").String())
	assert.Equal(t, "detail", v.FieldByName("Detail").String())
}

func TestRedactSensitiveUnexportedField(t *testing.T) {
	original := struct {
		Password string
		secret   string
	}{"pass", "secret"}

	redacted := RedactSensitive(original)

	v := reflect.ValueOf(redacted)

	assert.Equal(t, "REDACTED", v.FieldByName("Password").String())
	assert.Equal(t, "", v.FieldByName("secret").String())
}
