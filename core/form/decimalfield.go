package form

import (
	"fmt"
	"github.com/funnydog/mailadmin/decimal"
)

type DecimalField struct {
	Required  bool
	Label     string
	Precision int
}

func (f *DecimalField) Clean(value string) (interface{}, error) {
	if value == "" {
		if f.Required {
			return nil, errRequired
		}
		value = "0"
	}

	d, err := decimal.Parse(value)
	if err != nil {
		return d, err
	}

	return d, nil

}

func (f *DecimalField) Update(name string, value interface{}, fv *FieldValue) {
	if value == nil {
		value = ""
	} else {
		var trunc int
		if f.Precision > 0 && f.Precision <= 4 {
			trunc = 4 - f.Precision
		} else if f.Precision == 0 {
			trunc = 5
		}

		fv.Value = fmt.Sprint(value.(decimal.Decimal))
		fv.Value = fv.Value[0:(len(fv.Value) - trunc)]
	}

	fv.Label = f.Label
	fv.Required = f.Required
}
