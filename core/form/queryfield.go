package form

import (
	"fmt"
	"strconv"
)

type ErrQueryChoiceNotFound int64

func (qn ErrQueryChoiceNotFound) Error() string {
	return fmt.Sprintf("QueryChoice with Key %d not found", qn)
}

type QueryChoice struct {
	Key   int64
	Value string
}

var emptyChoice = QueryChoice{0, "----------"}

type QueryField struct {
	Required bool
	Label    string
	Query    func() ([]QueryChoice, error)
	couples  map[int64]string
	choices  []QueryChoice
}

func (f *QueryField) populateCache() error {
	entries, err := f.Query()
	if err != nil {
		return err
	}

	f.couples = map[int64]string{}
	f.choices = []QueryChoice{}
	if !f.Required {
		f.couples[emptyChoice.Key] = emptyChoice.Value
		f.choices = append(f.choices, emptyChoice)
	}
	for _, entry := range entries {
		f.couples[entry.Key] = entry.Value
		f.choices = append(f.choices, entry)
	}

	return nil
}

func (f *QueryField) Clean(value string) (interface{}, error) {
	if value == "" {
		if f.Required {
			return nil, ErrRequired
		}
		return nil, nil
	}

	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, err
	}

	if id == 0 && f.Required {
		return nil, ErrInvalidValue
	}

	if f.couples == nil {
		if err = f.populateCache(); err != nil {
			return nil, err
		}
	}

	_, ok := f.couples[id]
	if !ok {
		return nil, ErrQueryChoiceNotFound(id)
	}

	return id, nil
}

func (f *QueryField) Update(name string, value interface{}, fv *FieldValue) {
	if f.Label != "" {
		fv.Label = f.Label
	} else {
		fv.Label = name
	}
	fv.Required = f.Required

	if f.couples == nil {
		_ = f.populateCache()
	}

	fv.Data = f.choices
	if value == nil {
		fv.Value = "0"
	} else if _, ok := f.couples[value.(int64)]; !ok {
		fv.Value = "0"
	} else {
		fv.Value = fmt.Sprintf("%d", value.(int64))
	}
}
