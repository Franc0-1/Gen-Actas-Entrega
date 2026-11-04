package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestSinListado(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "PDF"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "PDF", "acta.pdf"), []byte("pdf"), 0o644); err != nil {
		t.Fatal(err)
	}

	handler := sinListado(http.StripPrefix("/output/", http.FileServer(http.Dir(dir))))

	casos := []struct {
		ruta   string
		estado int
	}{
		{"/output/", http.StatusNotFound},
		{"/output/PDF/", http.StatusNotFound},
		{"/output/PDF", http.StatusMovedPermanently}, // FileServer redirige a "/output/PDF/", que da 404
		{"/output/PDF/acta.pdf", http.StatusOK},
		{"/output/PDF/no-existe.pdf", http.StatusNotFound},
	}

	for _, c := range casos {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, c.ruta, nil))
		if rec.Code != c.estado {
			t.Errorf("GET %s: estado %d, se esperaba %d", c.ruta, rec.Code, c.estado)
		}
	}
}
