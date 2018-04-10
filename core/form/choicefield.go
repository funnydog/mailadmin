package form

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
