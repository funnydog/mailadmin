package form

import (
	"testing"

	. "github.com/funnydog/mailadmin/testutils"
)

func TestTextFieldClean(t *testing.T) {
	field := TextField{MaxLength: 10, Required: false, Label: "Label"}

	_, err := field.Clean("0123456789A")
	if err == nil {
		t.Error("Expected error but got no error instead")
		return
	}

	_, err = field.Clean("")
	if err != nil {
		t.Error(err)
		return
	}

	value, err := field.Clean("0123456789")
	if err != nil {
		t.Error(err)
		return
	}

	AssertStringEqual(t, value.(string), "0123456789")
}

func TestTextFieldUpdate(t *testing.T) {
	value := FieldValue{}
	field := TextField{Required: false, Label: "label"}

	field.Update("text", nil, &value)

	AssertStringEqual(t, value.Value, "")
	AssertStringEqual(t, value.Label, field.Label)
	AssertBoolEqual(t, value.Required, field.Required)
}
