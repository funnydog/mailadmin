package form

import "fmt"

type ErrChoiceNotFound string

func (nf ErrChoiceNotFound) Error() string {
	return fmt.Sprintf("Choice '%s' not found", string(nf))
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
	if value != "" {
		for _, c := range f.Choices {
			if c.Key == value {
				return c.Key, nil
			}
		}
		return nil, ErrChoiceNotFound(value)
	} else if f.Required {
		return nil, ErrRequired
	} else {
		return nil, nil
	}
}

func (f *ChoiceField) Update(name string, value interface{}, fv *FieldValue) {
	if f.Label != "" {
		fv.Label = f.Label
	} else {
		fv.Label = name
	}

	fv.Required = f.Required
	fv.Data = f.Choices

	if value != nil {
		fv.Value = value.(string)
	} else if f.Required && len(f.Choices) > 0 {
		fv.Value = f.Choices[0].Key
	} else {
		fv.Value = ""
	}

}
