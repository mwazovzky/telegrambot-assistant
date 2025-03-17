package configloader

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func LoadConfig(cfg interface{}) error {
	if err := loadConfig(cfg); err != nil {
		return err
	}
	if err := validateConfig(cfg); err != nil {
		return err
	}
	return nil
}

func loadConfig(cfg interface{}) error {
	v := reflect.ValueOf(cfg).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		envKey := fieldType.Tag.Get("env")
		required := fieldType.Tag.Get("required") == "true"

		if envKey == "" {
			if err := loadConfig(field.Addr().Interface()); err != nil {
				return err
			}
			continue
		}

		envValue := os.Getenv(envKey)
		if envValue == "" && required {
			return fmt.Errorf("missing required environment variable: %s", envKey)
		}

		if err := setFieldValue(field, fieldType, envValue); err != nil {
			return err
		}
	}
	return nil
}

func setFieldValue(field reflect.Value, fieldType reflect.StructField, envValue string) error {
	if fieldType.Type == reflect.TypeOf(time.Duration(0)) {
		if value, err := strconv.Atoi(envValue); err == nil {
			field.Set(reflect.ValueOf(time.Duration(value) * time.Second))
		} else {
			return err
		}
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(envValue)
	case reflect.Int64:
		if value, err := strconv.ParseInt(envValue, 10, 64); err == nil {
			field.SetInt(value)
		} else {
			return err
		}
	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.Int64 {
			values := []int64{}
			for _, v := range strings.Split(envValue, ",") {
				if value, err := strconv.ParseInt(v, 10, 64); err == nil {
					values = append(values, value)
				} else {
					return err
				}
			}
			field.Set(reflect.ValueOf(values))
		}
	}
	return nil
}

func validateConfig(cfg interface{}) error {
	v := reflect.ValueOf(cfg).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		required := fieldType.Tag.Get("required") == "true"

		if required && isZeroValue(field) {
			return fmt.Errorf("missing required configuration: %s", fieldType.Name)
		}
	}
	return nil
}

func isZeroValue(v reflect.Value) bool {
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}
