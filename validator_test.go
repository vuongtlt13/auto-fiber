package autofiber_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	autofiber "github.com/vuongtlt13/auto-fiber"
)

func TestGetValidator(t *testing.T) {
	validator := autofiber.GetValidator()
	assert.NotNil(t, validator)
}
