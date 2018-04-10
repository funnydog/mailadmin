package form

import "net/mail"

type EmailField struct {
	Required bool
	Label    string
}

func (f *EmailField) Clean(value string) (interface{}, error) {
	if value == "" {
		if f.Required {
			return "", errRequired
		}
		return "", nil
	}

	parser := mail.AddressParser{}
	_, err := parser.Parse(value)
	if err != nil {
		return "", emailNotValid
	}

	return value, nil
}

func (f *EmailField) Update(name string, value interface{}, fv *FieldValue) {
	if value == nil {
		value = ""
	}

	if value == nil {
		fv.Value = ""
	} else {
		fv.Value = value.(string)
	}

	fv.Label = f.Label
	fv.Required = f.Required
}
