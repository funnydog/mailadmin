package form

import "time"

var date_formats = []string{
	"02/01/2006",
	"02/01/06",
}

type DateField struct {
	Required bool
	Label    string
}

func (f *DateField) Clean(value string) (interface{}, error) {
	if value == "" {
		if f.Required {
			return nil, ErrRequired
		}

		return nil, nil
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
	if f.Label != "" {
		fv.Label = f.Label
	} else {
		fv.Label = name
	}

	fv.Required = f.Required

	if value != nil {
		fv.Value = value.(time.Time).Format("02/01/2006")
	} else if f.Required {
		fv.Value = time.Now().Format("02/01/2006")
	} else {
		fv.Value = ""
	}
}
