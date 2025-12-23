// Package config is a minimal util to parse environment variables into structs.
//
// Example:
//
//	type Config struct {
//	    Port int `env:"PORT" env_default:"8080"`
//	}
//
//
//	var cfg Config
//	if err := config.Parse(&cfg); err != nil {
//		log.Fatal(err)
//	}
package config

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
)

const (
	envTag        = "env"
	envDefaultTag = "env_default"
)

// Parse parses a struct containing `env` tags and loads its values from environment variables.
//
// Supported types: string, int (all sizes), bool, and nested structs.
func Parse(cfg interface{}) error {
	refValue, err := getRefValue(cfg)
	if err != nil {
		return err
	}

	return parse(refValue)
}

func parse(refValue reflect.Value) error {
	refType := refValue.Type()
	var errs []error

	for i := range refType.NumField() {
		field := refType.Field(i)
		fieldValue := refValue.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		if field.Type.Kind() == reflect.Struct {
			if err := parse(fieldValue); err != nil {
				errs = append(errs, err)
			}
			continue
		}

		env, err := getEnv(field)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		switch field.Type.Kind() {
		case reflect.String:
			fieldValue.SetString(env)
		case reflect.Bool:
			value, err := strconv.ParseBool(env)
			if err != nil {
				errs = append(errs, ErrParseErr{
					Field: field.Name,
					Type:  field.Type,
					Err:   err,
				})
				continue
			}

			fieldValue.SetBool(value)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			value, err := strconv.ParseInt(env, 10, field.Type.Bits())
			if err != nil {
				errs = append(errs, ErrParseErr{
					Field: field.Name,
					Type:  field.Type,
					Err:   err,
				})
				continue
			}

			fieldValue.SetInt(value)
		default:
			errs = append(
				errs,
				fmt.Errorf("unsupported type %q for field %s", field.Type, field.Name),
			)
		}
	}

	return errors.Join(errs...)
}

func getRefValue(cfg interface{}) (reflect.Value, error) {
	refValue := reflect.ValueOf(cfg)
	if refValue.Kind() != reflect.Pointer || refValue.IsNil() {
		return reflect.Value{}, ErrNotStructPtr{}
	}

	refValue = refValue.Elem()
	if refValue.Kind() != reflect.Struct {
		return reflect.Value{}, ErrNotStructPtr{}
	}

	return refValue, nil
}

func getEnv(v reflect.StructField) (string, error) {
	envKey, ok := v.Tag.Lookup(envTag)
	if !ok || envKey == "" {
		return "", nil
	}

	envValue := os.Getenv(envKey)
	if envValue != "" {
		return envValue, nil
	}

	defValue := v.Tag.Get(envDefaultTag)
	if defValue != "" {
		return defValue, nil
	}

	return "", ErrEmptyValue{Field: v.Name, Key: envKey}
}
