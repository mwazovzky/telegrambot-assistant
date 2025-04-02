package configloader

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequiredValidator_Validate(t *testing.T) {
	tests := []struct {
		name     string
		field    interface{}
		required bool
		wantErr  bool
	}{
		{"empty string required", "", true, true},
		{"non-empty string required", "test", true, false},
		{"empty string not required", "", false, false},
		{"empty slice required", []string{}, true, true},
		{"non-empty slice required", []string{"test"}, true, false},
		{"zero int required", 0, true, true},
		{"non-zero int required", 42, true, false},
	}

	validator := &RequiredValidator{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := reflect.ValueOf(tt.field)
			tag := reflect.StructTag("")
			if tt.required {
				tag = `required:"true"`
			}

			err := validator.Validate(value, tag)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_isZeroValue(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  bool
	}{
		{"empty string", "", true},
		{"non-empty string", "test", false},
		{"zero int", 0, true},
		{"non-zero int", 42, false},
		{"empty slice", []string{}, true},
		{"non-empty slice", []string{"test"}, false},
		{"zero float", 0.0, true},
		{"non-zero float", 3.14, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isZeroValue(reflect.ValueOf(tt.value))
			assert.Equal(t, tt.want, got)
		})
	}
}
