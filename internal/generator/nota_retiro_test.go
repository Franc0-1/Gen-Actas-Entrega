package generator

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"actas-project/internal/models"
)

func TestGenerateNotaRetiro(t *testing.T) {
	nota := models.NotaRetiro{
		Fecha:     time.Date(2026, time.October, 27, 0, 0, 0, 0, time.UTC),
		Empresa:   "Bienes & Servicios 3M",
		Elementos: []string{"Impresora marca Pantum modelo M6550NW nro. de serie CJ3X001670", "Monitor Samsung."},
	}
	outputPath := filepath.Join(t.TempDir(), "nota_retiro_prueba.docx")
	if err := GenerateNotaRetiro(nota, "../../templates/docx/plantilla_retiro.docx", outputPath); err != nil {
		t.Fatalf("no se pudo generar la nota: %v", err)
	}

	texto := extraerTexto(t, leerZip(t, outputPath)["word/document.xml"])
	for _, esperado := range []string{
		"Resistencia, 27 de octubre de 2026",
		"firma Bienes & Servicios 3M,",
		"1) Impresora marca Pantum modelo M6550NW nro. de serie CJ3X001670; 2) Monitor Samsung.",
	} {
		if !strings.Contains(texto, esperado) {
			t.Errorf("falta %q en el documento:\n%s", esperado, texto)
		}
	}
	if strings.Contains(texto, "{{") {
		t.Errorf("quedaron placeholders sin reemplazar:\n%s", texto)
	}
}
