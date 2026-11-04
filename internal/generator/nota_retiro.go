package generator

import (
	"fmt"
	"strings"

	"actas-project/internal/models"
)

// GenerateNotaRetiro arma el DOCX de la nota de retiro a partir de la
// plantilla, reemplazando fecha, empresa y el detalle de elementos.
func GenerateNotaRetiro(nota models.NotaRetiro, templatePath, outputPath string) error {
	if err := nota.Validate(); err != nil {
		return fmt.Errorf("nota inválida: %w", err)
	}

	// Formato de la nota: "27 de octubre de 2026".
	fecha := fmt.Sprintf("%d de %s de %d", nota.Fecha.Day(), strings.ToLower(meses[nota.Fecha.Month()-1]), nota.Fecha.Year())

	return reescribirDocumento(templatePath, outputPath, func(docXML string) (string, error) {
		return strings.NewReplacer(
			"{{fecha_larga}}", fecha,
			"{{empresa}}", escaparXML(strings.TrimSpace(nota.Empresa)),
			"{{detalle_elementos}}", escaparXML(detalleElementos(nota.Elementos)),
		).Replace(docXML), nil
	})
}

// detalleElementos numera los elementos como en el modelo de nota:
// "1) Impresora ...; 2) Monitor ...".
func detalleElementos(elementos []string) string {
	partes := make([]string, len(elementos))
	for i, e := range elementos {
		partes[i] = fmt.Sprintf("%d) %s", i+1, strings.TrimRight(strings.TrimSpace(e), "."))
	}
	return strings.Join(partes, "; ")
}
