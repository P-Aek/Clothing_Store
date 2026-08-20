package apidocs

import (
	_ "embed"

	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
)

// openAPISpec is embedded so API documentation works regardless of the
// directory from which the compiled server is started.
//
//go:embed openapi.yaml
var openAPISpec []byte

// Handler serves Swagger UI at /docs and the OpenAPI document at
// /docs/openapi.yaml.
func Handler() fiber.Handler {
	return swagger.New(swagger.Config{
		BasePath:         "/",
		FilePath:         "docs/openapi.yaml",
		FileContent:      openAPISpec,
		Path:             "docs",
		Title:            "Clothing Store API Docs",
		CacheAge:         300,
		SwaggerURL:       "https://unpkg.com/swagger-ui-dist@5.32.0/swagger-ui-bundle.js",
		SwaggerPresetURL: "https://unpkg.com/swagger-ui-dist@5.32.0/swagger-ui-standalone-preset.js",
		SwaggerStylesURL: "https://unpkg.com/swagger-ui-dist@5.32.0/swagger-ui.css",
	})
}
