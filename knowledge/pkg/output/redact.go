package output

import (
	"reflect"
	"slices"
	"strings"
)

var SensitiveFields = []string{
	"password",
	"apikey",
	"token",
	"secret",
	"credentials",
	"auth",
}

func RedactSensitive(s any, fields ...string) any {
	toRedact := fields
	if len(toRedact) == 0 {
		toRedact = SensitiveFields
	}

	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return s
	}

	redactedStruct := reflect.New(v.Type()).Elem()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := v.Type().Field(i)
		fieldName := strings.ToLower(fieldType.Name)

		// Skip unexported fields
		if !fieldType.IsExported() {
			continue
		}

		// Handle nested structs recursively
		if field.Kind() == reflect.Struct {
			redactedStruct.Field(i).Set(reflect.ValueOf(RedactSensitive(field.Interface(), toRedact...)))
			continue
		}

		if slices.Contains(toRedact, fieldName) {
			redactedStruct.Field(i).SetString("REDACTED")
		} else {
			redactedStruct.Field(i).Set(field)
		}
	}

	return redactedStruct.Interface()
}
