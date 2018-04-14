package form

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/funnydog/mailadmin/decimal"
)

type FieldValue struct {
	Name      string
	Label     string
	Value     string
	Error     string
	Required  bool
	Submitted bool
	Data      interface{}
}

type Subform struct {
	fields map[string]FormField
	Values []map[string]*FieldValue
}

func (s *Subform) Add(name string, field FormField) {
	s.fields[name] = field
}

func (s *Subform) SetValue(index int, name string, value interface{}) {
	field, ok := s.fields[name]
	if !ok {
		log.Panicf("Subform field %s not found\n", name)
	}

	for len(s.Values) <= index {
		s.Values = append(s.Values, map[string]*FieldValue{})
	}

	vname := fmt.Sprintf("%s_%d", name, index)
	fv, ok := s.Values[index][name]
	if !ok {
		fv = &FieldValue{Name: vname}
		s.Values[index][name] = fv
	}

	field.Update(name, value, fv)
}

type Form struct {
	fields    map[string]FormField
	cleanData map[string]interface{}
	Values    map[string]*FieldValue
	Formset   Subform
}

func (f *Form) Add(name string, field FormField) {
	f.fields[name] = field
	f.Values[name] = &FieldValue{Name: name}
	field.Update(name, nil, f.Values[name])
}

func (f *Form) SetString(fieldName string, value string) {
	field, ok := f.fields[fieldName]
	if !ok {
		log.Panicf("Field %s not found\n", fieldName)
	}

	field.Update(fieldName, value, f.Values[fieldName])
}

func (f *Form) SetTime(fieldName string, value time.Time) {
	field, ok := f.fields[fieldName]
	if !ok {
		log.Panicf("Field %s not found\n", fieldName)
	}

	field.Update(fieldName, value, f.Values[fieldName])
}

func (f *Form) SetBool(fieldName string, value bool) {
	field, ok := f.fields[fieldName]
	if !ok {
		log.Panicf("Field %s not found\n", fieldName)
	}

	field.Update(fieldName, value, f.Values[fieldName])
}

func (f *Form) SetInt64(fieldName string, value int64) {
	field, ok := f.fields[fieldName]
	if !ok {
		log.Panicf("Field %s not found\n", fieldName)
	}

	field.Update(fieldName, value, f.Values[fieldName])
}

func (f *Form) SetDecimal(fieldName string, value decimal.Decimal) {
	field, ok := f.fields[fieldName]
	if !ok {
		log.Panicf("Field %s not found\n", fieldName)
	}

	field.Update(fieldName, value, f.Values[fieldName])
}

func (f *Form) GetString(fieldName string) string {
	value, ok := f.cleanData[fieldName]
	if !ok {
		log.Panicf("Field %s not found\n", fieldName)
	}

	return value.(string)
}

func (f *Form) GetTime(fieldName string) time.Time {
	value, ok := f.cleanData[fieldName]
	if !ok {
		log.Panicf("Field %s not found\n", fieldName)
	}

	return value.(time.Time)
}

func (f *Form) GetBool(fieldName string) bool {
	value, ok := f.cleanData[fieldName]
	if !ok {
		log.Panicf("Field %s not found\n", fieldName)
	}

	return value.(bool)
}

func (f *Form) GetInt64(fieldName string) int64 {
	value, ok := f.cleanData[fieldName]
	if !ok {
		log.Panicf("Field %s not found\n", fieldName)
	}

	return value.(int64)
}

func (f *Form) GetDecimal(fieldName string) decimal.Decimal {
	value, ok := f.cleanData[fieldName]
	if !ok {
		log.Panicf("Field %s not found\n", fieldName)
	}

	return value.(decimal.Decimal)
}

func (f *Form) SetError(fieldname, err string) {
	fieldValue, ok := f.Values[fieldname]
	if ok {
		fieldValue.Error = err
	}
}

func (f *Form) Validate(r *http.Request) (valid bool) {
	valid = true
	for key, field := range f.fields {
		value := f.Values[key]
		value.Submitted = true

		clean, err := field.Clean(r.FormValue(key))
		if err != nil {
			field.Update(key, nil, value)
			value.Value = r.FormValue(key)
			value.Error = err.Error()
			valid = false
		} else {
			field.Update(key, clean, value)
			f.cleanData[key] = clean
		}
	}

	return valid
}

func Create() Form {
	return Form{
		fields:    map[string]FormField{},
		Values:    map[string]*FieldValue{},
		cleanData: map[string]interface{}{},
		Formset: Subform{
			fields: map[string]FormField{},
			Values: []map[string]*FieldValue{},
		},
	}
}
