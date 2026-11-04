package services

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"actas-project/internal/models"
)

// usarDirectoriosTemporales apunta outputDir/outputDirPDF a carpetas
// temporales durante el test y las restaura al terminar.
func usarDirectoriosTemporales(t *testing.T) {
	t.Helper()
	docxOriginal, pdfOriginal := outputDir, outputDirPDF
	t.Cleanup(func() { outputDir, outputDirPDF = docxOriginal, pdfOriginal })

	raiz := t.TempDir()
	outputDir = filepath.Join(raiz, "Docx")
	outputDirPDF = filepath.Join(raiz, "PDF")
	if err := PrepararDirectorios(); err != nil {
		t.Fatal(err)
	}
}

func crearArchivo(t *testing.T, ruta string) {
	t.Helper()
	if err := os.WriteFile(ruta, nil, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestNombreBaseDisponible(t *testing.T) {
	const base = "Acta de Entrega - Prensa - 2026-11-04"

	t.Run("sin archivos previos usa el nombre base", func(t *testing.T) {
		usarDirectoriosTemporales(t)
		if got := nombreBaseDisponible(base); got != base {
			t.Errorf("got %q, want %q", got, base)
		}
	})

	t.Run("DOCX existente agrega sufijo", func(t *testing.T) {
		usarDirectoriosTemporales(t)
		crearArchivo(t, filepath.Join(outputDir, base+".docx"))
		if got, want := nombreBaseDisponible(base), base+" (2)"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	// Regresión: el PDF vive en outputDirPDF, no junto al DOCX. Antes se
	// buscaba en outputDir y un PDF previo se sobrescribía si su DOCX ya no
	// estaba.
	t.Run("solo PDF existente agrega sufijo", func(t *testing.T) {
		usarDirectoriosTemporales(t)
		crearArchivo(t, filepath.Join(outputDirPDF, base+".pdf"))
		if got, want := nombreBaseDisponible(base), base+" (2)"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("sufijos ocupados avanzan al siguiente libre", func(t *testing.T) {
		usarDirectoriosTemporales(t)
		crearArchivo(t, filepath.Join(outputDir, base+".docx"))
		crearArchivo(t, filepath.Join(outputDirPDF, base+" (2).pdf"))
		if got, want := nombreBaseDisponible(base), base+" (3)"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestNombreBaseActa(t *testing.T) {
	fecha := time.Date(2026, time.November, 4, 0, 0, 0, 0, time.UTC)
	casos := []struct {
		tipo models.TipoActa
		area string
		want string
	}{
		{models.TipoActaEntrega, "Prensa", "Acta de Entrega - Prensa - 2026-11-04"},
		{models.TipoActaRecibimiento, "Prensa", "Acta de Recibimiento - Prensa - 2026-11-04"},
		{models.TipoActaAmbos, "Prensa", "Acta de Entrega y Recibimiento - Prensa - 2026-11-04"},
		{models.TipoActaEntrega, "Área: 1/2", "Acta de Entrega - Área 12 - 2026-11-04"},
	}

	for _, c := range casos {
		acta := models.Acta{Fecha: fecha, Tipo: c.tipo, AreaDepartamento: c.area}
		if got := nombreBaseActa(acta); got != c.want {
			t.Errorf("%s/%q: got %q, want %q", c.tipo, c.area, got, c.want)
		}
	}
}
