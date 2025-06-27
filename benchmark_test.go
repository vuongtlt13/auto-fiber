package autofiber_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	autofiber "github.com/vuongtlt13/auto-fiber"
)

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		app := autofiber.New()
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
		dg := autofiber.NewDocsGenerator("/api/v1")
		assert.NotNil(b, dg)
	}
}

func BenchmarkDocsGenerator_GenerateOpenAPISpec(b *testing.B) {
	dg := autofiber.NewDocsGenerator("/api/v1")
	info := autofiber.OpenAPIInfo{
		Title:   "Test API",
		Version: "1.0.0",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		spec := dg.GenerateOpenAPISpec(info)
		assert.NotNil(b, spec)
	}
}
