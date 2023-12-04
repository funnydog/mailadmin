package form

type CheckboxField struct {
	Required bool
	Label    string
}

func (f *CheckboxField) Clean(value string) (interface{}, error) {
	if value != "" {
		return true, nil
	} else if f.Required {
		return nil, ErrRequired
	} else {
		return false, nil
	}
}

func (f *CheckboxField) Update(name string, value interface{}, fv *FieldValue) {
	if f.Label != "" {
		fv.Label = f.Label
	} else {
		fv.Label = name
	}

	fv.Required = f.Required

	if value != nil && value.(bool) {
		fv.Value = "on"
	} else {
		fv.Value = ""
	}
}
