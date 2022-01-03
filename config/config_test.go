package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	ParseCommandLine()
	err := Init()
	assert.NoError(t, err)
}
