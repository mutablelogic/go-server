// Package static contains embedded static files for the OpenAPI handler.
package static

import _ "embed"

//go:embed openapi.html
var OpenAPIHTML []byte
