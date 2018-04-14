package form

import (
	"testing"

	. "github.com/funnydog/mailadmin/testutils"
)

func TestEmailFieldClean(t *testing.T) {
	field := EmailField{Required: true, Label: "email"}

	_, err := field.Clean("")
	if err == nil {
		t.Error("Expected error but got no error instead")
		return
	}

	value, err := field.Clean("toast@example.com")
	if err != nil {
		t.Error(err)
		return
	}
	AssertStringEqual(t, value.(string), "toast@example.com")

	field.Required = false
	_, err = field.Clean("")
	if err != nil {
		t.Error(err)
		return
	}
}

func TestEmailFieldUpdate(t *testing.T) {
	value := FieldValue{}
	field := EmailField{Required: true, Label: "email"}

	field.Update("field", nil, &value)
	AssertStringEqual(t, value.Value, "")
	AssertStringEqual(t, value.Label, field.Label)
	AssertBoolEqual(t, value.Required, field.Required)

	field.Update("field", "example@example.org", &value)
	AssertStringEqual(t, value.Value, "example@example.org")

	field.Required = false
	field.Update("field", nil, &value)
	AssertStringEqual(t, value.Value, "")

	field.Label = ""
	field.Update("field", nil, &value)
	AssertStringEqual(t, value.Label, "field")
}
