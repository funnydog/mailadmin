package form

import (
	"testing"
	"time"

	. "github.com/funnydog/mailadmin/testutils"
)

func TestDateFieldClean(t *testing.T) {
	field := DateField{Required: true, Label: "label"}

	value, err := field.Clean("11/04/2018")
	if err != nil {
		t.Error(err)
		return
	}

	AssertStringEqual(t, value.(time.Time).Format("02/01/2006"), "11/04/2018")

	value, err = field.Clean("11/04/18")
	if err != nil {
		t.Error(err)
		return
	}

	AssertStringEqual(t, value.(time.Time).Format("02/01/2006"), "11/04/2018")

	field = DateField{Required: false, Label: "label"}
	value, err = field.Clean("")
	if err != nil {
		t.Error(err)
		return
	}

	if value != nil {
		t.Errorf("Expected nil value, got %v", value)
		return
	}
}

func TestDateFieldUpdate(t *testing.T) {
	dateValue := FieldValue{}
	dateField := DateField{Required: false, Label: "label"}

	dateField.Update("date", nil, &dateValue)
	AssertStringEqual(t, dateValue.Value, "")
	AssertStringEqual(t, dateValue.Label, dateField.Label)
	AssertBoolEqual(t, dateValue.Required, dateField.Required)

	date, _ := time.Parse(date_formats[0], "11/04/2018")
	dateField.Update("date", date, &dateValue)
	AssertStringEqual(t, dateValue.Value, "11/04/2018")

	dateField.Required = true
	dateField.Update("date", nil, &dateValue)
	AssertStringNotEqual(t, dateValue.Value, "")

	dateField.Label = ""
	dateField.Update("date", nil, &dateValue)
	AssertStringEqual(t, dateValue.Label, "date")
}
