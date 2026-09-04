package services

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"actas-project/internal/generator"
	"actas-project/internal/models"
)

const (
	plantillaPath = "templates/docx/plantilla_acta.docx"
	outputDir     = "output"
)

// caracteresInvalidosArchivo son los caracteres que Windows (y, por
// prudencia, el resto de los SO) no permiten en un nombre de archivo.
var caracteresInvalidosArchivo = regexp.MustCompile(`[\\/:*?"<>|]`)

var espaciosDuplicados = regexp.MustCompile(`\s{2,}`)

// sanitizarNombreArchivo deja un fragmento de nombre de archivo seguro:
// sin caracteres inválidos para el sistema de archivos, sin espacios
// duplicados ni en los extremos.
func sanitizarNombreArchivo(s string) string {
	s = caracteresInvalidosArchivo.ReplaceAllString(s, "")
	s = espaciosDuplicados.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// nombreBaseActa arma el nombre descriptivo del acta (sin extensión) según
// el tipo: quién entrega, quién recibe o ambos, más la fecha del acta en
// formato ISO (AAAA-MM-DD).
func nombreBaseActa(acta models.Acta) string {
	fecha := acta.Fecha.Format("2006-01-02")

	var base string
	switch acta.Tipo {
	case models.TipoActaEntrega:
		base = fmt.Sprintf("Acta de Entrega - %s - %s", acta.QuienEntrega, fecha)
	case models.TipoActaRecibimiento:
		base = fmt.Sprintf("Acta de Recibimiento - %s - %s", acta.QuienRecibe, fecha)
	default: // models.TipoActaAmbos
		base = fmt.Sprintf("Acta - %s - %s - %s", acta.QuienEntrega, acta.QuienRecibe, fecha)
	}

	return sanitizarNombreArchivo(base)
}

// nombreBaseDisponible agrega un sufijo incremental " (2)", " (3)"... si ya
// existe un .docx o .pdf con ese nombre base en el directorio de salida,
// para no sobrescribir una acta generada antes.
func nombreBaseDisponible(base string) string {
	candidato := base
	for intento := 2; existeConEsaBase(candidato); intento++ {
		candidato = fmt.Sprintf("%s (%d)", base, intento)
	}
	return candidato
}

func existeConEsaBase(base string) bool {
	for _, ext := range []string{".docx", ".pdf"} {
		if _, err := os.Stat(filepath.Join(outputDir, base+ext)); err == nil {
			return true
		}
	}
	return false
}

// GenerarActa genera el DOCX correspondiente al acta, lo convierte a PDF y
// devuelve el nombre del PDF resultante (sin ruta de sistema, para que el
// llamador pueda armar una URL sin depender del separador del SO). El DOCX
// generado no se borra: queda en el directorio de salida como base del
// sistema. Ambos archivos comparten el mismo nombre descriptivo, cambiando
// solo la extensión.
func GenerarActa(acta models.Acta) (string, error) {
	base := nombreBaseDisponible(nombreBaseActa(acta))
	docxPath := filepath.Join(outputDir, base+".docx")

	if err := generator.Generate(acta, plantillaPath, docxPath); err != nil {
		return "", err
	}

	pdfPath, err := generator.ConvertirAPDF(docxPath, outputDir)
	if err != nil {
		return "", err
	}

	return filepath.Base(pdfPath), nil
}
