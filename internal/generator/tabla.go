package generator

import (
	"fmt"
	"strings"

	"actas-project/internal/models"
)

// aplicarTablas arma las tablas "Retira" y "Entrega" del documento: clona la
// fila molde una vez por cada elemento correspondiente, o elimina el bloque
// completo (encabezado + tabla) si no hay elementos para esa dirección.
func aplicarTablas(docXML string, elementos []models.Elemento) (string, error) {
	retira := filtrarPorDireccion(elementos, models.DireccionRetira)
	entrega := filtrarPorDireccion(elementos, models.DireccionEntrega)

	docXML, err := aplicarBloqueTabla(docXML, "Retira", retira)
	if err != nil {
		return "", err
	}

	docXML, err = aplicarBloqueTabla(docXML, "Entrega", entrega)
	if err != nil {
		return "", err
	}

	return docXML, nil
}

func filtrarPorDireccion(elementos []models.Elemento, direccion models.DireccionElemento) []models.Elemento {
	var resultado []models.Elemento
	for _, e := range elementos {
		if e.Direccion == direccion {
			resultado = append(resultado, e)
		}
	}
	return resultado
}

// aplicarBloqueTabla ubica el bloque "encabezado + tabla" identificado por el
// texto del encabezado (ej. "Retira") y lo reemplaza según corresponda.
func aplicarBloqueTabla(docXML, encabezado string, elementos []models.Elemento) (string, error) {
	bloque, inicio, fin, err := extraerBloque(docXML, encabezado)
	if err != nil {
		return "", err
	}

	if len(elementos) == 0 {
		return docXML[:inicio] + docXML[fin:], nil
	}

	bloqueConFilas, err := construirBloqueConFilas(bloque, elementos)
	if err != nil {
		return "", err
	}

	return docXML[:inicio] + bloqueConFilas + docXML[fin:], nil
}

// extraerBloque devuelve el fragmento de XML que va desde el párrafo que
// contiene el texto de encabezado hasta el cierre de la tabla que le sigue,
// junto con sus posiciones de inicio y fin dentro del documento completo.
func extraerBloque(docXML, encabezado string) (bloque string, inicio, fin int, err error) {
	marcaTexto := "<w:t>" + encabezado + "</w:t>"
	idxTexto := strings.Index(docXML, marcaTexto)
	if idxTexto == -1 {
		return "", 0, 0, fmt.Errorf("no se encontró el encabezado %q en la plantilla", encabezado)
	}

	inicio = strings.LastIndex(docXML[:idxTexto], "<w:p ")
	if inicio == -1 {
		return "", 0, 0, fmt.Errorf("no se encontró el párrafo del encabezado %q", encabezado)
	}

	idxCierreTabla := strings.Index(docXML[idxTexto:], "</w:tbl>")
	if idxCierreTabla == -1 {
		return "", 0, 0, fmt.Errorf("no se encontró la tabla siguiente al encabezado %q", encabezado)
	}
	fin = idxTexto + idxCierreTabla + len("</w:tbl>")

	return docXML[inicio:fin], inicio, fin, nil
}

// extraerFilas devuelve todas las filas (<w:tr>...</w:tr>) presentes en un
// fragmento de XML, en el orden en que aparecen.
func extraerFilas(xmlFragmento string) ([]string, error) {
	var filas []string
	pos := 0
	for {
		start := strings.Index(xmlFragmento[pos:], "<w:tr ")
		if start == -1 {
			break
		}
		start += pos

		relEnd := strings.Index(xmlFragmento[start:], "</w:tr>")
		if relEnd == -1 {
			return nil, fmt.Errorf("fila de tabla sin cierre en la plantilla")
		}
		end := start + relEnd + len("</w:tr>")

		filas = append(filas, xmlFragmento[start:end])
		pos = end
	}
	return filas, nil
}

// construirBloqueConFilas toma el bloque original (con su fila molde) y
// devuelve el bloque con una fila por cada elemento, ya con los placeholders
// de la fila reemplazados por los datos de cada elemento.
func construirBloqueConFilas(bloque string, elementos []models.Elemento) (string, error) {
	filas, err := extraerFilas(bloque)
	if err != nil {
		return "", err
	}
	if len(filas) < 2 {
		return "", fmt.Errorf("la tabla no tiene la fila molde esperada")
	}
	molde := filas[1]

	var filasGeneradas strings.Builder
	for _, elemento := range elementos {
		fila := molde
		fila = strings.Replace(fila, "{{descripcion}}", escaparXML(elemento.Descripcion), 1)
		fila = strings.Replace(fila, "{{nro_inventario}}", escaparXML(elemento.NroInventario), 1)
		fila = strings.Replace(fila, "{{cantidad}}", fmt.Sprintf("%d", elemento.Cantidad), 1)
		filasGeneradas.WriteString(fila)
	}

	return strings.Replace(bloque, molde, filasGeneradas.String(), 1), nil
}
