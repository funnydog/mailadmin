package form

import (
	"errors"
	"net/mail"
)

var (
	ErrInvalidEmail = errors.New("Please insert a valid email address.")
	emailParser     = mail.AddressParser{}
)

type EmailField struct {
	Required bool
	Label    string
}

func (f *EmailField) Clean(value string) (interface{}, error) {
	if value != "" {
		_, err := emailParser.Parse(value)
		if err != nil {
			return nil, ErrInvalidEmail
		}
		return value, nil
	} else if f.Required {
		return nil, ErrRequired
	} else {
		return nil, nil
	}
}

func (f *EmailField) Update(name string, value interface{}, fv *FieldValue) {
	if f.Label != "" {
		fv.Label = f.Label
	} else {
		fv.Label = name
	}

	fv.Required = f.Required

	if value != nil {
		fv.Value = value.(string)
	} else {
		fv.Value = ""
	}
}
