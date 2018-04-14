package form

import (
	"fmt"
	"strconv"
)

type IntegerField struct {
	Required bool
	Label    string
}

func (f *IntegerField) Clean(value string) (interface{}, error) {
	if value == "" {
		if f.Required {
			return nil, ErrRequired
		}

		return nil, nil
	}
	return strconv.ParseInt(value, 10, 64)
}

func (f *IntegerField) Update(name string, value interface{}, fv *FieldValue) {
	if f.Label != "" {
		fv.Label = f.Label
	} else {
		fv.Label = name
	}
	fv.Required = f.Required

	if value != nil {
		fv.Value = fmt.Sprint(value.(int64))
	} else if f.Required {
		fv.Value = "0"
	} else {
		fv.Value = ""
	}
}
