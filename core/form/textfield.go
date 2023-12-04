package form

import "fmt"

type ErrLengthExceeded int

func (el ErrLengthExceeded) Error() string {
	return fmt.Sprintf("Max Length of %d exceeded", el)
}

type TextField struct {
	MaxLength int
	Required  bool
	Label     string
}

func (f *TextField) Clean(value string) (interface{}, error) {
	if value != "" {
		if 0 < f.MaxLength && f.MaxLength < len(value) {
			return nil, ErrLengthExceeded(f.MaxLength)
		}
		return value, nil
	} else if f.Required {
		return nil, ErrRequired
	} else {
		return value, nil
	}

}

func (f *TextField) Update(name string, value interface{}, fv *FieldValue) {
	if f.Label != "" {
		fv.Label = f.Label
	} else {
		fv.Label = name
	}
	fv.Required = f.Required
	fv.Data = f.MaxLength

	if value == nil {
		fv.Value = ""
	} else {
		fv.Value = value.(string)
	}
}
