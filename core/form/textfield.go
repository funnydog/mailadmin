package form

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
