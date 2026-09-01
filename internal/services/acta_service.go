package services

import (
	"fmt"
	"path/filepath"
	"time"

	"actas-project/internal/generator"
	"actas-project/internal/models"
)

const (
	plantillaPath = "templates/docx/plantilla_acta.docx"
	outputDir     = "output"
)

// GenerarActa genera el DOCX correspondiente al acta en el directorio de
// salida y devuelve el nombre del archivo creado (sin ruta de sistema, para
// que el llamador pueda armar una URL sin depender del separador del SO).
func GenerarActa(acta models.Acta) (string, error) {
	nombreArchivo := fmt.Sprintf("acta_%d.docx", time.Now().UnixNano())
	outputPath := filepath.Join(outputDir, nombreArchivo)

	if err := generator.Generate(acta, plantillaPath, outputPath); err != nil {
		return "", err
	}

	return nombreArchivo, nil
}
