package form

import (
	"errors"
)

var (
	errRequired    = errors.New("This field cannot be empty.")
	lengthExceeded = errors.New("Length exceeded.")
	valueNotValid  = errors.New("Value not valid.")
	emailNotValid  = errors.New("Please insert a valid email address.")
)

type FormField interface {
	Clean(string) (interface{}, error)
	Update(name string, value interface{}, fv *FieldValue)
}
