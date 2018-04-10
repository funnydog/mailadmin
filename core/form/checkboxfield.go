package form

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
