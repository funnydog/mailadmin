package form

import (
	"errors"
	"fmt"
	"net/mail"
	"strconv"
	"time"

	"github.com/funnydog/measure/decimal"
)

var (
	errRequired    = errors.New("This field cannot be empty.")
	lengthExceeded = errors.New("length exceeded")
	valueNotValid  = errors.New("value not valid")
	emailNotValid  = errors.New("Please insert a valid email address.")
)

type FormField interface {
	Clean(string) (interface{}, error)
	Update(name string, value interface{}, fv *FieldValue)
}

type DateField struct {
	Required bool
	Label    string
}

func (f *DateField) Clean(value string) (interface{}, error) {
	if f.Required && value == "" {
		return time.Time{}, errRequired
	}

	date_formats := []string{
		"02/01/2006",
		"02/01/06",
	}

	var (
		err  error
		date time.Time
	)
	for _, format := range date_formats {
		date, err = time.Parse(format, value)
		if err == nil {
			break
		}
	}
	return date, err
}

func (f *DateField) Update(name string, value interface{}, fv *FieldValue) {
	if value == nil {
		value = time.Now()
	}

	if f.Label != "" {
		fv.Label = f.Label
	} else {
		fv.Label = name
	}

	fv.Value = value.(time.Time).Format("02/01/2006")
	fv.Required = f.Required
}

type TextField struct {
	MaxLength int
	Required  bool
	Label     string
}

func (f *TextField) Clean(value string) (interface{}, error) {
	if f.Required && value == "" {
		return "", errRequired
	}

	if f.MaxLength > 0 && len(value) > f.MaxLength {
		return "", lengthExceeded
	}

	return value, nil
}

func (f *TextField) Update(name string, value interface{}, fv *FieldValue) {
	if value == nil {
		fv.Value = ""
	} else {
		fv.Value = value.(string)
	}

	fv.Label = f.Label
	fv.Required = f.Required
}

type CheckboxField struct {
	Required bool
	Label    string
}

func (f *CheckboxField) Clean(value string) (interface{}, error) {
	if value == "" {
		if f.Required {
			return false, errRequired
		}

		return false, nil
	}

	return true, nil
}

func (f *CheckboxField) Update(name string, value interface{}, fv *FieldValue) {
	if value == nil {
		value = false
	}

	if f.Label != "" {
		fv.Label = f.Label
	} else {
		fv.Label = name
	}

	if value == nil {
		fv.Value = ""
	} else if value.(bool) {
		fv.Value = "checked"
	} else {
		fv.Value = ""
	}

	fv.Required = f.Required
}

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

type Choice struct {
	Key   string
	Value string
}

type ChoiceField struct {
	Required bool
	Label    string
	Choices  []Choice
}

func (f *ChoiceField) Clean(value string) (interface{}, error) {
	if value == "" && f.Required {
		return value, errRequired
	}

	return value, nil
}

func (f *ChoiceField) Update(name string, value interface{}, fv *FieldValue) {
	if value == nil {
		fv.Value = ""
	} else {
		fv.Value = value.(string)
	}

	fv.Label = f.Label
	fv.Data = f.Choices
	fv.Required = f.Required
}

type QueryField struct {
	Required bool
	Label    string
	Query    func() ([]Choice, error)
}

func (f *QueryField) Clean(value string) (interface{}, error) {
	if value == "" {
		if f.Required {
			return nil, errRequired
		}
		return int64(0), nil
	}

	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, err
	}

	if id == 0 && f.Required {
		return nil, errRequired
	}

	return id, nil
}

func (f *QueryField) Update(name string, value interface{}, fv *FieldValue) {
	if fv.Data == nil {
		couples := []Choice{}
		if !f.Required {
			couples = append(couples, Choice{"0", "--------"})
		}
		entries, err := f.Query()
		if err == nil {
			for _, entry := range entries {
				couples = append(couples, entry)
			}
		}
		fv.Data = couples
	}

	if value == nil {
		fv.Value = "0"
	} else {
		fv.Value = fmt.Sprintf("%d", value.(int64))
	}

	fv.Label = f.Label
	fv.Required = f.Required
}
