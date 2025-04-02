package configloader

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type StringParser struct{}

func (p *StringParser) Parse(value string, field reflect.Value) error {
	field.SetString(value)
	return nil
}

type Int64Parser struct{}

func (p *Int64Parser) Parse(value string, field reflect.Value) error {
	if value == "" {
		return nil
	}
	v, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return err
	}
	field.SetInt(v)
	return nil
}

type IntParser struct{}

func (p *IntParser) Parse(value string, field reflect.Value) error {
	if value == "" {
		return nil
	}
	v, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	field.SetInt(int64(v))
	return nil
}

type SliceParser struct{}

func (p *SliceParser) Parse(value string, field reflect.Value) error {
	if value == "" {
		return nil
	}

	values := strings.Split(value, ",")
	slice := reflect.MakeSlice(field.Type(), 0, len(values))

	elemParser, ok := defaultParsers[field.Type().Elem().Kind()]
	if !ok {
		return fmt.Errorf("unsupported slice element type: %v", field.Type().Elem().Kind())
	}

	for _, v := range values {
		elem := reflect.New(field.Type().Elem()).Elem()
		if err := elemParser.Parse(v, elem); err != nil {
			return err
		}
		slice = reflect.Append(slice, elem)
	}

	field.Set(slice)
	return nil
}

type DurationParser struct{}

func (p *DurationParser) Parse(value string, field reflect.Value) error {
	if value == "" {
		return nil
	}

	// If no time unit is specified, assume seconds
	if _, err := strconv.Atoi(value); err == nil {
		value += "s"
	}

	d, err := time.ParseDuration(value)
	if err != nil {
		return err
	}
	field.Set(reflect.ValueOf(d))
	return nil
}

var defaultParsers = map[reflect.Kind]ValueParser{
	reflect.String: &StringParser{},
	reflect.Int64:  &Int64Parser{},
	reflect.Int:    &IntParser{},
	reflect.Slice:  &SliceParser{},
}
