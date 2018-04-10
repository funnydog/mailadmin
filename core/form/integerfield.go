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
			return nil, errRequired
		}
		value = "0"
	}
	return strconv.ParseInt(value, 10, 64)
}

func (f *IntegerField) Update(name string, value interface{}, fv *FieldValue) {
	if value == nil {
		fv.Value = ""
	} else {
		fv.Value = fmt.Sprint(value.(int64))
	}

	fv.Label = f.Label
	fv.Required = f.Required
}
