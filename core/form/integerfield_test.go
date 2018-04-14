package form

import (
	"fmt"
	"strconv"
	"testing"

	. "github.com/funnydog/mailadmin/testutils"
)

func TestIntegerFieldClean(t *testing.T) {
	field := IntegerField{Required: true, Label: "label"}

	_, err := field.Clean("")
	if err == nil {
		t.Error("Expected error but got no error")
		return
	}

	field.Required = false
	value, err := field.Clean("")
	if err != nil {
		t.Error(err)
		return
	}

	if value != nil {
		t.Errorf("Expected nil value but got (%v) instead", value)
		return
	}

	_, err = field.Clean("blablab")
	if err == nil {
		t.Error("Expected error but got no error instead")
		return
	}

	value, err = field.Clean("12345")
	if err != nil {
		t.Error(err)
		return
	}

	AssertStringEqual(t, fmt.Sprint(value.(int64)), "12345")
}

func TestIntegerFieldUpdate(t *testing.T) {
	value := FieldValue{}
	field := IntegerField{Required: false, Label: "label"}

	field.Update("int", nil, &value)
	AssertStringEqual(t, value.Value, "")
	AssertStringEqual(t, value.Label, field.Label)
	AssertBoolEqual(t, value.Required, field.Required)

	int, _ := strconv.ParseInt("12345", 10, 64)
	if int != int64(12345) {
		t.Errorf("Expected (%d); Actual data (%d)", 12345, int)
		return
	}

	field.Required = true
	field.Update("int", nil, &value)
	AssertStringEqual(t, value.Value, "0")

	field.Label = ""
	field.Update("int", nil, &value)
	AssertStringEqual(t, value.Label, "int")

	field.Update("int", int64(54321), &value)
	AssertStringEqual(t, value.Value, "54321")
}
