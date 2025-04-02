package configloader

import (
	"fmt"
	"os"
	"reflect"
	"time"
)

// ValueParser is responsible for parsing string values into specific types
type ValueParser interface {
	Parse(value string, field reflect.Value) error
}

// Validator is responsible for validating field values
type Validator interface {
	Validate(field reflect.Value, tags reflect.StructTag) error
}

// EnvLoader loads values from environment variables
type EnvLoader struct {
	parsers    map[reflect.Kind]ValueParser
	validators []Validator
}

// Option represents a configuration option for EnvLoader
type Option func(*EnvLoader)

// WithParser adds a custom parser for a specific type
func WithParser(kind reflect.Kind, parser ValueParser) Option {
	return func(l *EnvLoader) {
		l.parsers[kind] = parser
	}
}

// WithValidator adds a custom validator
func WithValidator(validator Validator) Option {
	return func(l *EnvLoader) {
		l.validators = append(l.validators, validator)
	}
}

var defaultLoader = NewEnvLoader()

// LoadConfig maintains backward compatibility using the default loader
func LoadConfig(cfg interface{}) error {
	return defaultLoader.LoadConfig(cfg)
}

// NewEnvLoader creates a new EnvLoader with default parsers and validators
func NewEnvLoader(opts ...Option) *EnvLoader {
	l := &EnvLoader{
		parsers: map[reflect.Kind]ValueParser{
			reflect.String: &StringParser{},
			reflect.Int64:  &Int64Parser{},
			reflect.Int:    &IntParser{},
			reflect.Slice:  &SliceParser{},
		},
		validators: []Validator{&RequiredValidator{}},
	}

	// Apply custom options
	for _, opt := range opts {
		opt(l)
	}

	return l
}

// LoadConfig loads configuration from environment variables
func (l *EnvLoader) LoadConfig(cfg interface{}) error {
	v := reflect.ValueOf(cfg)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("config must be a pointer")
	}

	return l.loadStruct(v.Elem())
}

func (l *EnvLoader) loadStruct(v reflect.Value) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if err := l.loadField(field, fieldType); err != nil {
			return fmt.Errorf("field %s: %w", fieldType.Name, err)
		}
	}

	return nil
}

func (l *EnvLoader) loadField(field reflect.Value, fieldType reflect.StructField) error {
	envKey := fieldType.Tag.Get("env")
	if envKey == "" {
		return nil
	}

	envValue := os.Getenv(envKey)

	// Special handling for time.Duration
	if fieldType.Type == reflect.TypeOf(time.Duration(0)) {
		parser := &DurationParser{}
		if err := parser.Parse(envValue, field); err != nil {
			return err
		}
		return nil
	}

	// Parse other types
	parser, ok := l.parsers[field.Kind()]
	if !ok {
		return fmt.Errorf("unsupported type: %v", field.Kind())
	}

	if err := parser.Parse(envValue, field); err != nil {
		return err
	}

	// Validate value
	for _, validator := range l.validators {
		if err := validator.Validate(field, fieldType.Tag); err != nil {
			return err
		}
	}

	return nil
}
