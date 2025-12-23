package config

import (
	"reflect"
	"testing"
)

type testCase struct {
	Name      string
	EnvVars   map[string]string
	Input     interface{}
	Output    interface{}
	ExpectErr bool
}

func TestParse(t *testing.T) {
	testCases := []testCase{
		{
			Name:      "non-pointer struct",
			EnvVars:   map[string]string{},
			Input:     struct{}{},
			ExpectErr: true,
		},
		{
			Name:      "empty struct pointer",
			EnvVars:   map[string]string{},
			Input:     &struct{}{},
			Output:    &struct{}{},
			ExpectErr: false,
		},
		{
			Name:    "missing env var with default",
			EnvVars: map[string]string{},
			Input: &struct {
				AppName string `env:"APP_NAME" env_default:"placeholder"`
			}{},
			Output: &struct {
				AppName string `env:"APP_NAME" env_default:"placeholder"`
			}{
				AppName: "placeholder",
			},
		},
		{
			Name:    "empty required value",
			EnvVars: map[string]string{},
			Input: &struct {
				AppName string `env:"APP_NAME"`
			}{},
			ExpectErr: true,
		},
		{
			Name:    "unexported fields",
			EnvVars: map[string]string{},
			Input: &struct {
				port       int    `env:"PORT"`
				secret_key string `env:"SECRET_KEY" env_default:"123"`
			}{},
			Output: &struct {
				port       int    `env:"PORT"`
				secret_key string `env:"SECRET_KEY" env_default:"123"`
			}{},
			ExpectErr: false,
		},
		{
			Name: "string, number and boolean fields",
			EnvVars: map[string]string{
				"PORT":     "8080",
				"APP_NAME": "go-commons",
			},
			Input: &struct {
				AppName string `env:"APP_NAME"`
				Port    int    `env:"PORT"`
				Debug   bool   `env:"DEBUG" env_default:"true"`
			}{},
			Output: &struct {
				AppName string `env:"APP_NAME"`
				Port    int    `env:"PORT"`
				Debug   bool   `env:"DEBUG" env_default:"true"`
			}{
				AppName: "go-commons",
				Port:    8080,
				Debug:   true,
			},
			ExpectErr: false,
		},
		{
			Name: "nested struct",
			EnvVars: map[string]string{
				"APP_NAME": "test-app",
				"DB_HOST":  "127.0.0.1",
				"DB_PORT":  "27017",
			},
			Input: &struct {
				AppName  string `env:"APP_NAME"`
				Database struct {
					Host    string `env:"DB_HOST"`
					Port    int    `env:"DB_PORT"`
					SSLMode bool   `env:"DB_SSL_MODE" env_default:"true"`
				}
			}{},
			Output: &struct {
				AppName  string `env:"APP_NAME"`
				Database struct {
					Host    string `env:"DB_HOST"`
					Port    int    `env:"DB_PORT"`
					SSLMode bool   `env:"DB_SSL_MODE" env_default:"true"`
				}
			}{
				AppName: "test-app",
				Database: struct {
					Host    string `env:"DB_HOST"`
					Port    int    `env:"DB_PORT"`
					SSLMode bool   `env:"DB_SSL_MODE" env_default:"true"`
				}{
					Host:    "127.0.0.1",
					Port:    27017,
					SSLMode: true,
				},
			},
			ExpectErr: false,
		},
		{
			Name: "unsupported type",
			EnvVars: map[string]string{
				"DATA": "test",
			},
			Input: &struct {
				Data map[string]string `env:"DATA"`
			}{},
			ExpectErr: true,
		},
		{
			Name: "invalid int format",
			EnvVars: map[string]string{
				"PORT": "NaN",
			},
			Input: &struct {
				Port int `env:"PORT"`
			}{},
			ExpectErr: true,
		},
		{
			Name: "invalid bool format",
			EnvVars: map[string]string{
				"DEBUG": "maybe",
			},
			Input: &struct {
				Debug bool `env:"DEBUG"`
			}{},
			ExpectErr: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			for key, val := range test.EnvVars {
				t.Setenv(key, val)
			}

			err := Parse(test.Input)

			if test.ExpectErr && err == nil {
				t.Fatal("expected error but got nil")
			}

			if !test.ExpectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !test.ExpectErr && !reflect.DeepEqual(test.Input, test.Output) {
				t.Errorf("expected %+v to be equal to %+v", test.Input, test.Output)
			}
		})
	}
}
