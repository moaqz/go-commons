package config

import (
	"fmt"
	"reflect"
)

// ErrEmptyValue is returned when a required field is not set.
type ErrEmptyValue struct {
	Field string
	Key   string
}

func (e ErrEmptyValue) Error() string {
	return fmt.Sprintf("required environment variable %q is not set", e.Key)
}

// ErrNotStructPtr is returned when Parse receives an invalid pointer.
type ErrNotStructPtr struct{}

func (e ErrNotStructPtr) Error() string {
	return "expected a non-nil pointer to a struct"
}

// ErrParseErr is returned when a field value cannot be parsed to its target type.
// Parse may return multiple ErrParseErr values joined with errors.Join()
type ErrParseErr struct {
	Field string
	Type  reflect.Type
	Err   error
}

func (e ErrParseErr) Error() string {
	return fmt.Sprintf("parse error on field %q of type %q: %v", e.Field, e.Type, e.Err)
}
