package handlers

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed ../../docs/openapi.yaml
var openapiSpec []byte

// OpenAPISpec serves the raw OpenAPI YAML file.
func OpenAPISpec(c *gin.Context) {
	c.Data(http.StatusOK, "application/yaml", openapiSpec)
}

// SwaggerUI serves a self-contained Swagger UI page.
func SwaggerUI(c *gin.Context) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8"/>
  <title>Lambare API – Swagger UI</title>
  <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/5.17.14/swagger-ui.min.css"/>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/5.17.14/swagger-ui-bundle.min.js"></script>
  <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/5.17.14/swagger-ui-standalone-preset.min.js"></script>
  <script>
    SwaggerUIBundle({
      url: "/openapi.yaml",
      dom_id: "#swagger-ui",
      presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
      layout: "StandaloneLayout",
      deepLinking: true,
      persistAuthorization: true,
    });
  </script>
</body>
</html>`
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}
