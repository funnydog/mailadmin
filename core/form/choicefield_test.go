package form

import (
	"testing"

	. "github.com/funnydog/mailadmin/testutils"
)

func TestChoiceFieldClean(t *testing.T) {
	field := ChoiceField{
		Required: true,
		Label:    "choicefield",
		Choices: []Choice{
			Choice{Key: "1", Value: "Option 1"},
			Choice{Key: "2", Value: "Option 2"},
			Choice{Key: "3", Value: "Option 3"},
		},
	}

	_, err := field.Clean("")
	if err == nil {
		t.Error("Expected error but got no error instead")
		return
	}

	value, err := field.Clean("1")
	if err != nil {
		t.Error(err)
		return
	}
	AssertStringEqual(t, value.(string), "1")

	value, err = field.Clean("4")
	if err == nil {
		t.Error("Expected error but got no error instead")
		return
	}
}

func TestChoiceFieldUpdate(t *testing.T) {
	value := FieldValue{}
	field := ChoiceField{
		Required: true,
		Label:    "choicefield",
		Choices: []Choice{
			Choice{Key: "1", Value: "Option 1"},
			Choice{Key: "2", Value: "Option 2"},
			Choice{Key: "3", Value: "Option 3"},
		},
	}

	field.Update("field", nil, &value)
	AssertStringEqual(t, value.Value, field.Choices[0].Key)
	AssertStringEqual(t, value.Label, field.Label)
	AssertBoolEqual(t, value.Required, field.Required)
	if value.Data == nil {
		t.Error("Data field not populated")
		return
	}

	field.Update("field", "1", &value)
	AssertStringEqual(t, value.Value, "1")

	field.Update("field", "2", &value)
	AssertStringEqual(t, value.Value, "2")

	field.Required = false
	field.Update("field", "", &value)
	AssertStringEqual(t, value.Value, "")

	field.Label = ""
	field.Update("field", "", &value)
	AssertStringEqual(t, value.Label, "field")
}
