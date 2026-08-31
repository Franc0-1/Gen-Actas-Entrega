package generator

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"strings"
	"testing"
	"time"

	"actas-project/internal/models"
)

const plantillaPrueba = "../../templates/docx/plantilla_acta.docx"

func actaDePrueba() models.Acta {
	return models.Acta{
		Fecha:            time.Date(2026, time.August, 31, 0, 0, 0, 0, time.UTC),
		QuienEntrega:     "Juan Pérez",
		QuienRecibe:      "María Gómez",
		Tipo:             models.TipoActaAmbos,
		AreaDepartamento: "Dirección de Sistemas",
		Elementos: []models.Elemento{
			{Descripcion: "Notebook Dell Latitude", NroInventario: "1001", Cantidad: 1, Direccion: models.DireccionRetira},
			{Descripcion: "Mouse inalámbrico", NroInventario: "1002", Cantidad: 2, Direccion: models.DireccionRetira},
			{Descripcion: "Impresora HP LaserJet", NroInventario: "2001", Cantidad: 1, Direccion: models.DireccionEntrega},
			{Descripcion: "Monitor Samsung 24\"", NroInventario: "2002", Cantidad: 1, Direccion: models.DireccionEntrega},
			{Descripcion: "Teclado USB", NroInventario: "", Cantidad: 3, Direccion: models.DireccionEntrega},
		},
	}
}

// leerZip devuelve el contenido de cada archivo interno del .docx, indexado
// por nombre.
func leerZip(t *testing.T, path string) map[string][]byte {
	t.Helper()
	r, err := zip.OpenReader(path)
	if err != nil {
		t.Fatalf("no se pudo abrir %s: %v", path, err)
	}
	defer r.Close()

	archivos := map[string][]byte{}
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("no se pudo abrir %s dentro de %s: %v", f.Name, path, err)
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Fatalf("no se pudo leer %s dentro de %s: %v", f.Name, path, err)
		}
		archivos[f.Name] = data
	}
	return archivos
}

// extraerTexto recorre el XML con un decoder (que decodifica entidades como
// &#34; automáticamente) y concatena el contenido de cada elemento <w:t>.
func extraerTexto(t *testing.T, docXML []byte) string {
	t.Helper()
	dec := xml.NewDecoder(bytes.NewReader(docXML))

	var sb strings.Builder
	dentroDeT := false
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("error recorriendo el XML: %v", err)
		}
		switch el := tok.(type) {
		case xml.StartElement:
			dentroDeT = el.Name.Local == "t"
		case xml.EndElement:
			if el.Name.Local == "t" {
				dentroDeT = false
				sb.WriteByte('\n')
			}
		case xml.CharData:
			if dentroDeT {
				sb.Write(el)
			}
		}
	}
	return sb.String()
}

func TestGenerate_ActaAmbos(t *testing.T) {
	acta := actaDePrueba()
	outputPath := "../../output/acta_prueba_ambos.docx"

	if err := Generate(acta, plantillaPrueba, outputPath); err != nil {
		t.Fatalf("Generate devolvió error: %v", err)
	}

	plantillaArchivos := leerZip(t, plantillaPrueba)
	generadoArchivos := leerZip(t, outputPath)

	// 1. El documento generado debe tener exactamente los mismos archivos
	// internos que la plantilla (mismos estilos, header, footer, media...).
	if len(plantillaArchivos) != len(generadoArchivos) {
		t.Fatalf("cantidad de archivos internos distinta: plantilla=%d generado=%d", len(plantillaArchivos), len(generadoArchivos))
	}
	for nombre, original := range plantillaArchivos {
		generado, ok := generadoArchivos[nombre]
		if !ok {
			t.Fatalf("el documento generado no contiene %s", nombre)
		}
		if nombre == "word/document.xml" {
			continue // este archivo se espera que cambie
		}
		if !bytes.Equal(original, generado) {
			t.Errorf("%s cambió respecto a la plantilla y no debería (estilos/formato)", nombre)
		}
	}

	docXML := generadoArchivos["word/document.xml"]

	// 2. El XML resultante debe ser válido (no se rompió ninguna etiqueta).
	dec := xml.NewDecoder(bytes.NewReader(docXML))
	for {
		_, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("el XML generado no es válido: %v", err)
		}
	}

	// 3. No debe quedar ningún placeholder sin reemplazar.
	if strings.Contains(string(docXML), "{{") {
		t.Errorf("quedó al menos un placeholder sin reemplazar en el documento generado")
	}

	texto := extraerTexto(t, docXML)

	// 4. Fecha en palabras.
	for _, esperado := range []string{"TREINTA Y UN", "AGOSTO", "DOS MIL VEINTISEIS"} {
		if !strings.Contains(texto, esperado) {
			t.Errorf("el texto generado no contiene %q (fecha en palabras)", esperado)
		}
	}

	// 5. Tipo de acta.
	if !strings.Contains(texto, "la entrega y el recibimiento") {
		t.Errorf("el texto generado no contiene la frase esperada para tipo 'ambos'")
	}

	// 6. Área/departamento.
	if !strings.Contains(texto, "Dirección de Sistemas") {
		t.Errorf("el texto generado no contiene el área/departamento")
	}

	// 7. Elementos de "retira" presentes.
	for _, esperado := range []string{"Notebook Dell Latitude", "INV-1001", "Mouse inalámbrico", "INV-1002"} {
		if !strings.Contains(texto, esperado) {
			t.Errorf("falta el elemento de retira %q en el texto generado", esperado)
		}
	}

	// 8. Elementos de "entrega" presentes.
	for _, esperado := range []string{"Impresora HP LaserJet", "INV-2001", "Monitor Samsung 24\"", "INV-2002", "Teclado USB"} {
		if !strings.Contains(texto, esperado) {
			t.Errorf("falta el elemento de entrega %q en el texto generado", esperado)
		}
	}

	// 8b. "Teclado USB" no tiene NroInventario: la celda debe quedar vacía,
	// sin el prefijo "INV-" colgando. De los 5 elementos, solo 4 tienen
	// número de inventario, así que "INV-" debe aparecer exactamente 4 veces.
	if n := strings.Count(texto, "INV-"); n != 4 {
		t.Errorf("cantidad de prefijos INV- inesperada: got %d, want 4 (el elemento sin NroInventario no debe mostrar el prefijo)", n)
	}

	// 9. Los elementos de retira no deben terminar en la tabla de entrega y
	// viceversa: separamos el texto en el punto donde empieza cada bloque.
	idxRetira := strings.Index(texto, "Retira")
	idxEntrega := strings.Index(texto, "Entrega")
	if idxRetira == -1 || idxEntrega == -1 || idxRetira > idxEntrega {
		t.Fatalf("no se encontraron los encabezados Retira/Entrega en el orden esperado")
	}
	bloqueRetira := texto[idxRetira:idxEntrega]
	bloqueEntrega := texto[idxEntrega:]

	for _, item := range []string{"Notebook Dell Latitude", "Mouse inalámbrico"} {
		if !strings.Contains(bloqueRetira, item) {
			t.Errorf("%q debería estar en el bloque Retira", item)
		}
		if strings.Contains(bloqueEntrega, item) {
			t.Errorf("%q no debería aparecer en el bloque Entrega", item)
		}
	}
	for _, item := range []string{"Impresora HP LaserJet", "Monitor Samsung 24\"", "Teclado USB"} {
		if !strings.Contains(bloqueEntrega, item) {
			t.Errorf("%q debería estar en el bloque Entrega", item)
		}
		if strings.Contains(bloqueRetira, item) {
			t.Errorf("%q no debería aparecer en el bloque Retira", item)
		}
	}

	// 10. Cantidad de filas clonadas: la molde ya no debe aparecer, y la
	// cantidad de <w:tr> en cada tabla debe ser encabezado + N elementos.
	if strings.Contains(texto, "Mouse DX-110") || strings.Contains(texto, "Mouse HDC") {
		t.Errorf("la fila molde original de la plantilla no debería aparecer en el documento generado")
	}

	filasRetiraYEntrega := strings.Count(string(docXML), "<w:tr ")
	// 2 filas de encabezado + 2 elementos retira + 3 elementos entrega = 7
	if filasRetiraYEntrega != 7 {
		t.Errorf("cantidad de filas <w:tr> inesperada: got %d, want 7", filasRetiraYEntrega)
	}
}
