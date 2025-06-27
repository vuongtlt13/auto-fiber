package autofiber_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	autofiber "github.com/vuongtlt13/auto-fiber"
)

func TestNew(t *testing.T) {
	app := autofiber.New()
	assert.NotNil(t, app)
	assert.NotNil(t, app.App)
}

func TestNew_WithConfig(t *testing.T) {
	app := autofiber.New()
	assert.NotNil(t, app)
	assert.NotNil(t, app.App)
}
