package configloader

import (
	"fmt"
	"reflect"
)

type RequiredValidator struct{}

func (v *RequiredValidator) Validate(field reflect.Value, tags reflect.StructTag) error {
	if tags.Get("required") != "true" {
		return nil
	}

	if isZeroValue(field) {
		return fmt.Errorf("required field is empty")
	}

	return nil
}

func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Slice, reflect.Map:
		return v.Len() == 0
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	default:
		return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
	}
}
