package configloader

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type TestConfig struct {
	StringField   string        `env:"STRING_FIELD" required:"true"`
	IntField      int64         `env:"INT_FIELD" required:"true"`
	SliceField    []int64       `env:"SLICE_FIELD" required:"true"`
	DurationField time.Duration `env:"DURATION_FIELD" required:"true"`
	OptionalField string        `env:"OPTIONAL_FIELD" required:"false"`
}

func TestLoadConfig(t *testing.T) {
	// Set environment variables for testing
	os.Setenv("STRING_FIELD", "test_string")
	os.Setenv("INT_FIELD", "12345")
	os.Setenv("SLICE_FIELD", "12345,67890")
	os.Setenv("DURATION_FIELD", "60")
	os.Setenv("OPTIONAL_FIELD", "optional_value")

	cfg := &TestConfig{}
	err := LoadConfig(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, "test_string", cfg.StringField)
	assert.Equal(t, int64(12345), cfg.IntField)
	assert.Equal(t, []int64{12345, 67890}, cfg.SliceField)
	assert.Equal(t, time.Duration(60)*time.Second, cfg.DurationField)
	assert.Equal(t, "optional_value", cfg.OptionalField)
}

func TestLoadConfigMissingRequired(t *testing.T) {
	// Unset environment variables to test missing required fields
	os.Unsetenv("STRING_FIELD")
	os.Unsetenv("INT_FIELD")
	os.Unsetenv("SLICE_FIELD")
	os.Unsetenv("DURATION_FIELD")

	cfg := &TestConfig{}
	err := LoadConfig(cfg)
	assert.Error(t, err)
}

func TestLoadConfigInvalidValues(t *testing.T) {
	// Set invalid environment variables for testing
	os.Setenv("STRING_FIELD", "test_string")
	os.Setenv("INT_FIELD", "invalid")
	os.Setenv("SLICE_FIELD", "invalid")
	os.Setenv("DURATION_FIELD", "invalid")

	cfg := &TestConfig{}
	err := LoadConfig(cfg)
	assert.Error(t, err)
}

func TestLoadConfigEmptyValues(t *testing.T) {
	// Set empty environment variables for testing
	os.Setenv("STRING_FIELD", "")
	os.Setenv("INT_FIELD", "")
	os.Setenv("SLICE_FIELD", "")
	os.Setenv("DURATION_FIELD", "")

	cfg := &TestConfig{}
	err := LoadConfig(cfg)
	assert.Error(t, err)
}

func TestLoadConfigPartialValues(t *testing.T) {
	// Set partial environment variables for testing
	os.Setenv("STRING_FIELD", "test_string")
	os.Setenv("INT_FIELD", "12345")
	os.Unsetenv("SLICE_FIELD")
	os.Unsetenv("DURATION_FIELD")

	cfg := &TestConfig{}
	err := LoadConfig(cfg)
	assert.Error(t, err)
}
