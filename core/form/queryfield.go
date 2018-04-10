package form

import (
	"fmt"
	"strconv"
)

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
