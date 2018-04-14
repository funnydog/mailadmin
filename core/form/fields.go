package form

import "errors"

var (
	ErrRequired     = errors.New("This field cannot be empty.")
	ErrInvalidValue = errors.New("Value not valid.")
)

type FormField interface {
	Clean(string) (interface{}, error)
	Update(name string, value interface{}, fv *FieldValue)
}
