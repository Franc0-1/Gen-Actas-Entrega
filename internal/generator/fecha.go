package generator

import (
	"strings"
	"time"
)

var meses = [...]string{
	"ENERO", "FEBRERO", "MARZO", "ABRIL", "MAYO", "JUNIO",
	"JULIO", "AGOSTO", "SEPTIEMBRE", "OCTUBRE", "NOVIEMBRE", "DICIEMBRE",
}

var unidades = [...]string{"", "UNO", "DOS", "TRES", "CUATRO", "CINCO", "SEIS", "SIETE", "OCHO", "NUEVE"}
var especiales10a19 = [...]string{"DIEZ", "ONCE", "DOCE", "TRECE", "CATORCE", "QUINCE", "DIECISEIS", "DIECISIETE", "DIECIOCHO", "DIECINUEVE"}
var veintenas = [...]string{"VEINTE", "VEINTIUNO", "VEINTIDOS", "VEINTITRES", "VEINTICUATRO", "VEINTICINCO", "VEINTISEIS", "VEINTISIETE", "VEINTIOCHO", "VEINTINUEVE"}
var decenas = [...]string{"", "", "", "TREINTA", "CUARENTA", "CINCUENTA", "SESENTA", "SETENTA", "OCHENTA", "NOVENTA"}
var centenas = [...]string{"", "CIENTO", "DOSCIENTOS", "TRESCIENTOS", "CUATROCIENTOS", "QUINIENTOS", "SEISCIENTOS", "SETECIENTOS", "OCHOCIENTOS", "NOVECIENTOS"}

// numeroATexto convierte un entero no negativo (hasta 9999) a su
// representación en palabras en español, en mayúsculas y sin tildes,
// siguiendo el estilo ya usado en la plantilla (ej. "AGOSTO", "VEINTIUNO").
func numeroATexto(n int) string {
	if n == 0 {
		return "CERO"
	}

	var partes []string

	miles := n / 1000
	n %= 1000
	if miles > 0 {
		if miles == 1 {
			partes = append(partes, "MIL")
		} else {
			partes = append(partes, numeroATexto(miles), "MIL")
		}
	}

	if n == 100 {
		partes = append(partes, "CIEN")
		n = 0
	} else {
		cientos := n / 100
		n %= 100
		if cientos > 0 {
			partes = append(partes, centenas[cientos])
		}
	}

	switch {
	case n == 0:
		// nada más que agregar
	case n < 10:
		partes = append(partes, unidades[n])
	case n < 20:
		partes = append(partes, especiales10a19[n-10])
	case n < 30:
		partes = append(partes, veintenas[n-20])
	default:
		d := n / 10
		u := n % 10
		if u == 0 {
			partes = append(partes, decenas[d])
		} else {
			partes = append(partes, decenas[d]+" Y "+unidades[u])
		}
	}

	return strings.Join(partes, " ")
}

// diaATexto convierte el día del mes a su forma apocopada, tal como se usa
// al anteponerlo a "días" (ej. "VEINTIUN días", no "VEINTIUNO días").
func diaATexto(dia int) string {
	texto := numeroATexto(dia)
	if strings.HasSuffix(texto, "UNO") {
		texto = strings.TrimSuffix(texto, "O")
	}
	return texto
}

// FechaEnPalabras descompone una fecha en día, mes y año en palabras, en el
// formato usado por la plantilla ("VEINTIUN", "AGOSTO", "DOS MIL VEINTISEIS").
func FechaEnPalabras(t time.Time) (diaTexto, mesTexto, anioTexto string) {
	diaTexto = diaATexto(t.Day())
	mesTexto = meses[t.Month()-1]
	anioTexto = numeroATexto(t.Year())
	return diaTexto, mesTexto, anioTexto
}
