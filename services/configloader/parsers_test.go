package configloader

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringParser_Parse(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"normal string", "test", "test"},
		{"empty string", "", ""},
		{"with spaces", "test value", "test value"},
		{"with special chars", "test@123!", "test@123!"},
	}

	parser := &StringParser{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := reflect.New(reflect.TypeOf("")).Elem()
			err := parser.Parse(tt.value, field)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, field.String())
		})
	}
}

func TestInt64Parser_Parse(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    int64
		wantErr bool
	}{
		{"valid number", "123", 123, false},
		{"empty string", "", 0, false},
		{"invalid number", "abc", 0, true},
		{"negative number", "-123", -123, false},
	}

	parser := &Int64Parser{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := reflect.New(reflect.TypeOf(int64(0))).Elem()
			err := parser.Parse(tt.value, field)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, field.Int())
			}
		})
	}
}

func TestSliceParser_Parse(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		typ     reflect.Type
		want    interface{}
		wantErr bool
	}{
		{
			name:  "string slice",
			value: "a,b,c",
			typ:   reflect.TypeOf([]string{}),
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "int64 slice",
			value: "1,2,3",
			typ:   reflect.TypeOf([]int64{}),
			want:  []int64{1, 2, 3},
		},
		{
			name:    "invalid int slice",
			value:   "1,a,3",
			typ:     reflect.TypeOf([]int64{}),
			wantErr: true,
		},
		{
			name:  "empty slice",
			value: "",
			typ:   reflect.TypeOf([]string{}),
			want:  []string(nil),
		},
	}

	parser := &SliceParser{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := reflect.New(tt.typ).Elem()
			err := parser.Parse(tt.value, field)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, field.Interface())
			}
		})
	}
}
