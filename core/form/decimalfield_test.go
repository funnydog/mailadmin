package form

import (
	"testing"

	"github.com/funnydog/mailadmin/decimal"
	. "github.com/funnydog/mailadmin/testutils"
)

func TestDecimalFieldClean(t *testing.T) {
	field := DecimalField{Required: true, Label: "decimal", Precision: 2}

	_, err := field.Clean("")
	if err == nil {
		t.Error("Expected error but got no error")
	}

	_, err = field.Clean("blabla")
	if err == nil {
		t.Error("Expected error but got no error")
		return
	}

	value, err := field.Clean("12.333")
	if err != nil {
		t.Error(err)
		return
	}

	if value != decimal.Decimal(123330) {
		t.Errorf("Expected (%v) but got (%v) instead", value,
			decimal.Decimal(123330))
		return
	}

	field.Required = false
	_, err = field.Clean("")
	if err != nil {
		t.Error(err)
		return
	}
}

func TestDecimalFieldUpdate(t *testing.T) {
	value := FieldValue{}
	field := DecimalField{Required: true, Label: "decimal", Precision: 2}

	field.Update("field", nil, &value)
	AssertStringEqual(t, value.Value, "0.00")
	AssertStringEqual(t, value.Label, field.Label)
	AssertBoolEqual(t, value.Required, field.Required)

	field.Required = false
	field.Update("field", nil, &value)
	AssertStringEqual(t, value.Value, "")

	field.Update("field", decimal.Decimal(123330), &value)
	AssertStringEqual(t, value.Value, "12.33")

	field.Precision = 0
	field.Update("field", decimal.Decimal(123330), &value)
	AssertStringEqual(t, value.Value, "12")

	field.Precision = 4
	field.Update("field", decimal.Decimal(123330), &value)
	AssertStringEqual(t, value.Value, "12.3330")

	field.Precision = 5
	field.Update("field", decimal.Decimal(123330), &value)
	AssertStringEqual(t, value.Value, "12")

	field.Label = ""
	field.Update("field", nil, &value)
	AssertStringEqual(t, value.Label, "field")
}
