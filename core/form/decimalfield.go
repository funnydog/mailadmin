package form

import "github.com/funnydog/mailadmin/decimal"

type DecimalField struct {
	Required  bool
	Label     string
	Precision int
}

func (f *DecimalField) Clean(value string) (interface{}, error) {
	if value == "" {
		if f.Required {
			return nil, ErrRequired
		}
		return nil, nil
	}

	d, err := decimal.Parse(value)
	if err != nil {
		return nil, err
	}

	return d, nil

}

func (f *DecimalField) Update(name string, value interface{}, fv *FieldValue) {
	if f.Label != "" {
		fv.Label = f.Label
	} else {
		fv.Label = name
	}
	fv.Required = f.Required

	if f.Required && value == nil {
		value = decimal.Decimal(0)
	}

	if value != nil {
		var trunc int
		if 0 < f.Precision && f.Precision <= 4 {
			trunc = 4 - f.Precision
		} else {
			trunc = 5
		}

		fv.Value = value.(decimal.Decimal).String()
		fv.Value = fv.Value[0:(len(fv.Value) - trunc)]
	} else {
		fv.Value = ""
	}
}
