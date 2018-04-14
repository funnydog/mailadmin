package form

import (
	"testing"

	. "github.com/funnydog/mailadmin/testutils"
)

func queryFunc() ([]QueryChoice, error) {
	return []QueryChoice{
		QueryChoice{Key: 1, Value: "Option 1"},
		QueryChoice{Key: 2, Value: "Option 2"},
		QueryChoice{Key: 3, Value: "Option 3"},
	}, nil
}

func TestQueryFieldClean(t *testing.T) {
	field := QueryField{Label: "queryfield", Required: true, Query: queryFunc}

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
	if value.(int64) != int64(1) {
		t.Errorf("Expected value %d; Actual value %d\n",
			int64(1), value.(int64))
		return
	}

	_, err = field.Clean("4")
	if err == nil {
		t.Error("Expected error but got no error instead")
		return
	}

	field = QueryField{Label: "queryfield", Required: false, Query: queryFunc}
	value, err = field.Clean("")
	if err != nil {
		t.Error(err)
		return
	}
	if value != nil {
		t.Errorf("Expected value nil; Actual value %v", value)
		return
	}

	value, err = field.Clean("0")
	if err != nil {
		t.Error(err)
		return
	}
	if value.(int64) != 0 {
		t.Errorf("Expected value %d; Actual value %d\n",
			0, value.(int64))
		return
	}
}

func TestQueryFieldUpdate(t *testing.T) {
	value := FieldValue{}
	field := QueryField{Label: "queryfield", Required: true, Query: queryFunc}

	field.Update("field", nil, &value)
	AssertStringEqual(t, value.Value, "0")
	AssertStringEqual(t, value.Label, field.Label)
	AssertBoolEqual(t, value.Required, field.Required)
	if value.Data == nil {
		t.Error("Data field not populated")
		return
	}

	field.Update("field", int64(1), &value)
	AssertStringEqual(t, value.Value, "1")
	field.Update("field", int64(2), &value)
	AssertStringEqual(t, value.Value, "2")
	field.Update("field", int64(4), &value)
	AssertStringEqual(t, value.Value, "0")

	field.Label = ""
	field.Update("field", nil, &value)
	AssertStringEqual(t, value.Label, "field")
}
