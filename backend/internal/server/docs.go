package server

import (
	"net/http"
	"os"
	"sync"
)

var (
	openapiSpec     []byte
	openapiSpecOnce sync.Once
)

func loadOpenAPISpec() []byte {
	openapiSpecOnce.Do(func() {
		// Try common locations
		for _, path := range []string{
			"api/openapi.yaml",
			"../api/openapi.yaml",
			"/app/api/openapi.yaml",
		} {
			data, err := os.ReadFile(path)
			if err == nil {
				openapiSpec = data
				return
			}
		}
	})
	return openapiSpec
}

func handleOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	spec := loadOpenAPISpec()
	if spec == nil {
		http.Error(w, "OpenAPI spec not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Write(spec)
}

func handleDocs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<!doctype html>
<html>
<head>
  <title>MailBrief API</title>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
</head>
<body>
  <script id="api-reference" data-url="/api/openapi.yaml"></script>
  <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>`))
}
