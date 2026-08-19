// Command typlyd serves the Typly editor and render API.
// It is intentionally deployment-agnostic; Railway deployment configuration
// is left to the application owner.
package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/EmiiFont/typly/pkg/typly"
	"github.com/starfederation/datastar-go/datastar"
)

//go:embed web/index.html
var webFS embed.FS

var indexTemplate = template.Must(template.ParseFS(webFS, "web/index.html"))

func main() {
	addr := ":8080"
	if value := os.Getenv("PORT"); value != "" {
		addr = ":" + value
	}
	log.Printf("typlyd listening on %s", addr)
	if err := http.ListenAndServe(addr, routes()); err != nil {
		log.Fatal(err)
	}
}

func routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handleIndex)
	mux.HandleFunc("GET /healthz", handleHealth)
	mux.HandleFunc("GET /api/spec/default", handleDefaultSpec)
	mux.HandleFunc("POST /api/render/gif", handleGIF)
	mux.HandleFunc("POST /api/render/mp4", handleMP4)
	mux.HandleFunc("POST /api/spec/validate", handleValidate)
	return withJSONHeaders(mux)
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}

func handleIndex(w http.ResponseWriter, _ *http.Request) {
	if err := indexTemplate.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleDefaultSpec(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, typly.DefaultRenderSpec())
}

func handleGIF(w http.ResponseWriter, r *http.Request) {
	spec, err := decodeSpec(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	data, err := renderGIF(spec)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Content-Disposition", `attachment; filename="typly.gif"`)
	_, _ = w.Write(data)
}

func handleMP4(w http.ResponseWriter, r *http.Request) {
	spec, err := decodeSpec(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	path, cleanup, err := renderMP4(spec)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer cleanup()
	data, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, "read rendered video: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Content-Disposition", `attachment; filename="typly.mp4"`)
	_, _ = w.Write(data)
}

// handleValidate demonstrates the Datastar read-patch-response flow. The
// web editor can use this endpoint without writing client-side state logic.
func handleValidate(w http.ResponseWriter, r *http.Request) {
	var spec typly.RenderSpec
	if err := datastar.ReadSignals(r, &spec); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sse := datastar.NewSSE(w, r)
	if err := spec.Validate(); err != nil {
		_ = sse.PatchElements(fmt.Sprintf(`<p id="validation" class="error">%s</p>`, template.HTMLEscapeString(err.Error())))
		return
	}
	_ = sse.PatchElements(`<p id="validation" class="valid">Ready to render.</p>`)
}

func decodeSpec(r *http.Request) (typly.RenderSpec, error) {
	defer r.Body.Close()
	var spec typly.RenderSpec
	decoder := json.NewDecoder(io.LimitReader(r.Body, 128*1024))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&spec); err != nil {
		return spec, fmt.Errorf("decode render spec: %w", err)
	}
	if err := spec.Validate(); err != nil {
		return spec, err
	}
	return spec, nil
}

func renderGIF(spec typly.RenderSpec) ([]byte, error) {
	renderer, err := rendererFor(spec)
	if err != nil {
		return nil, err
	}
	defer renderer.Close()
	frames, err := typly.Animate(renderer, spec.Sentences)
	if err != nil {
		return nil, err
	}
	var output bytes.Buffer
	if err := typly.EncodeGIF(&output, frames); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func renderMP4(spec typly.RenderSpec) (string, func(), error) {
	renderer, err := rendererFor(spec)
	if err != nil {
		return "", func() {}, err
	}
	defer renderer.Close()
	frames, err := typly.Animate(renderer, spec.Sentences)
	if err != nil {
		return "", func() {}, err
	}
	file, err := os.CreateTemp("", "typly-*.mp4")
	if err != nil {
		return "", func() {}, err
	}
	path := file.Name()
	if err := file.Close(); err != nil {
		os.Remove(path)
		return "", func() {}, err
	}
	if err := typly.EncodeMP4(path, frames, spec.FPS); err != nil {
		os.Remove(path)
		return "", func() {}, err
	}
	return path, func() { _ = os.Remove(path) }, nil
}

func rendererFor(spec typly.RenderSpec) (*typly.Renderer, error) {
	cfg, err := spec.Config()
	if err != nil {
		return nil, err
	}
	return typly.NewRenderer(cfg)
}

func withJSONHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
