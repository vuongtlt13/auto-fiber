package autofiber_test

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	autofiber "github.com/vuongtlt13/auto-fiber"
)

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		app := autofiber.New(fiber.Config{})
		assert.NotNil(b, app)
		assert.NotNil(b, app.App)
	}
}

func BenchmarkGetValidator(b *testing.B) {
	for i := 0; i < b.N; i++ {
		validator := autofiber.GetValidator()
		assert.NotNil(b, validator)
	}
}

func BenchmarkNewDocsGenerator(b *testing.B) {
	for i := 0; i < b.N; i++ {
		dg := autofiber.NewDocsGenerator()
		assert.NotNil(b, dg)
	}
}

func BenchmarkDocsGenerator_GenerateOpenAPISpec(b *testing.B) {
	dg := autofiber.NewDocsGenerator()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		spec := dg.GenerateOpenAPISpec()
		assert.NotNil(b, spec)
	}
}
