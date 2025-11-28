package swagger

import (
	"bytes"
	"net/http"
	"text/template"

	swaggerdocs "fin-track-app/api/swagger"
)

type pageData struct {
	SpecURL string
}

const pageTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = () => {
      SwaggerUIBundle({
        url: "{{ .SpecURL }}",
        dom_id: '#swagger-ui'
      });
    };
  </script>
</body>
</html>`

func UIHandler(specURL string) http.HandlerFunc {
	tmpl := template.Must(template.New("swagger").Parse(pageTemplate))

	return func(w http.ResponseWriter, r *http.Request) {
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, pageData{SpecURL: specURL}); err != nil {
			http.Error(w, "failed to render swagger page", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(buf.Bytes())
	}
}

func SpecHandler(content []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		_, _ = w.Write(content)
	}
}

func FinAPISpec() []byte {
	return swaggerdocs.FinAPISpec()
}

func FinAnalyticsSpec() []byte {
	return swaggerdocs.FinAnalyticsSpec()
}
