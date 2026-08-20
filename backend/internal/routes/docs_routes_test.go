package routes

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestSwaggerDocumentationRoutes(t *testing.T) {
	app := fiber.New()
	Register(app, nil, nil, nil, nil, nil, nil, "test-secret")

	tests := []struct {
		name        string
		path        string
		contentType string
		bodyText    string
	}{
		{name: "UI", path: "/docs", contentType: "text/html", bodyText: "Clothing Store API Docs"},
		{name: "OpenAPI specification", path: "/docs/openapi.yaml", contentType: "application/yaml", bodyText: "openapi: 3.0.3"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response, err := app.Test(httptest.NewRequest("GET", test.path, nil))
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			if response.StatusCode != fiber.StatusOK {
				t.Fatalf("status = %d, want %d", response.StatusCode, fiber.StatusOK)
			}
			if got := response.Header.Get(fiber.HeaderContentType); !strings.Contains(got, test.contentType) {
				t.Fatalf("content type = %q, want it to contain %q", got, test.contentType)
			}
			body, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(body), test.bodyText) {
				t.Fatalf("response body does not contain %q", test.bodyText)
			}
		})
	}
}
