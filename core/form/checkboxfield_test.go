package form

import (
	"testing"

	. "github.com/funnydog/mailadmin/testutils"
)

func TestCheckboxFieldClean(t *testing.T) {
	field := CheckboxField{Required: true, Label: "checkbox"}

	_, err := field.Clean("")
	if err == nil {
		t.Error("Expected error but got no error instead")
		return
	}

	value, err := field.Clean("on")
	if err != nil {
		t.Error(err)
		return
	}

	field.Required = false
	value, err = field.Clean("")
	if err != nil {
		t.Error(err)
		return
	}

	AssertBoolEqual(t, value.(bool), false)
}

func TestCheckboxFieldUpdate(t *testing.T) {
	value := FieldValue{}
	field := CheckboxField{Required: false, Label: "checkbox"}

	field.Update("field", nil, &value)
	AssertStringEqual(t, value.Value, "")
	AssertStringEqual(t, value.Label, field.Label)
	AssertBoolEqual(t, value.Required, field.Required)

	field.Update("field", true, &value)
	AssertStringEqual(t, value.Value, "on")

	field.Update("field", false, &value)
	AssertStringEqual(t, value.Value, "")

	field.Required = true
	field.Update("field", nil, &value)
	AssertStringEqual(t, value.Value, "")

	field.Label = ""
	field.Update("field", nil, &value)
	AssertStringEqual(t, value.Label, "field")
}
