package form

import "time"

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
