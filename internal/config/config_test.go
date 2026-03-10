package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	cfg, err := LoadConfig("")
	assert.Nil(t, err)
	assert.NotNil(t, cfg)

	t.Logf("%+v", cfg)
}
