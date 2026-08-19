package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EmiiFont/typly/pkg/typly"
)

func TestDefaultSpecEndpoint(t *testing.T) {
	recorder := httptest.NewRecorder()
	routes().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/spec/default", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	var spec typly.RenderSpec
	if err := json.Unmarshal(recorder.Body.Bytes(), &spec); err != nil {
		t.Fatal(err)
	}
	if spec.Emoji != "color" || spec.Width != 1280 {
		t.Errorf("unexpected default spec: %+v", spec)
	}
}

func TestHealthEndpoint(t *testing.T) {
	recorder := httptest.NewRecorder()
	routes().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if recorder.Code != http.StatusOK || recorder.Body.String() != "ok\n" {
		t.Errorf("health response = %d %q", recorder.Code, recorder.Body.String())
	}
}

func TestGIFEndpoint(t *testing.T) {
	spec := typly.DefaultRenderSpec()
	spec.Sentences = []string{"Hi 🌍"}
	spec.Width, spec.Height, spec.FontSize, spec.Blinks = 320, 180, 24, 0
	body, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/render/gif", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	routes().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if recorder.Header().Get("Content-Type") != "image/gif" {
		t.Errorf("content type = %q, want image/gif", recorder.Header().Get("Content-Type"))
	}
	if !bytes.HasPrefix(recorder.Body.Bytes(), []byte("GIF")) {
		t.Fatal("response is not a GIF")
	}
}

func TestGIFEndpointRejectsUnknownFields(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/render/gif", bytes.NewBufferString(`{"sentences":["x"],"unknown":true}`))
	routes().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", recorder.Code)
	}
}
