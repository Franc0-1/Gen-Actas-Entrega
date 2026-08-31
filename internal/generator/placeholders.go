package generator

import (
	"bytes"
	"encoding/xml"
	"strings"

	"actas-project/internal/models"
)

// escaparXML escapa caracteres especiales (&, <, >, comillas) para que un
// valor arbitrario pueda insertarse como texto dentro del XML del documento.
func escaparXML(s string) string {
	var buf bytes.Buffer
	if err := xml.EscapeText(&buf, []byte(s)); err != nil {
		return s
	}
	return buf.String()
}

// tipoATexto arma la frase que reemplaza a {{tipo_texto}} en el párrafo
// introductorio, según el tipo de acta.
func tipoATexto(tipo models.TipoActa) string {
	switch tipo {
	case models.TipoActaEntrega:
		return "la entrega"
	case models.TipoActaRecibimiento:
		return "el recibimiento"
	case models.TipoActaAmbos:
		return "la entrega y el recibimiento"
	default:
		return ""
	}
}

// aplicarPlaceholdersSimples reemplaza los placeholders de texto simple
// (fecha, tipo de acta, área/departamento) por los valores del acta.
func aplicarPlaceholdersSimples(docXML string, acta models.Acta) string {
	diaTexto, mesTexto, anioTexto := FechaEnPalabras(acta.Fecha)

	reemplazos := map[string]string{
		"dia_texto":         diaTexto,
		"mes_texto":         mesTexto,
		"anio_texto":        anioTexto,
		"tipo_texto":        tipoATexto(acta.Tipo),
		"area_departamento": escaparXML(acta.AreaDepartamento),
	}

	for clave, valor := range reemplazos {
		docXML = strings.ReplaceAll(docXML, "{{"+clave+"}}", valor)
	}

	return docXML
}
