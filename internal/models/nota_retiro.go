package models

import (
	"errors"
	"strings"
	"time"
)

// NotaRetiro es la nota dirigida a la Guardia Policial de Casa de Gobierno
// que informa que personal de una empresa retirará ciertos elementos. El
// destinatario es fijo y vive en la plantilla.
type NotaRetiro struct {
	Fecha     time.Time `json:"fecha"`
	Empresa   string    `json:"empresa"`
	Elementos []string  `json:"elementos"` // ej. "Impresora marca Pantum modelo M6550NW nro. de serie CJ3X001670"
}

// Validate verifica que la nota tenga los datos mínimos para generarse.
func (n NotaRetiro) Validate() error {
	if n.Fecha.IsZero() {
		return errors.New("la nota debe tener una fecha")
	}
	if strings.TrimSpace(n.Empresa) == "" {
		return errors.New("la nota debe indicar la empresa que retira")
	}
	if len(n.Elementos) == 0 {
		return errors.New("la nota debe tener al menos un elemento")
	}
	for _, e := range n.Elementos {
		if strings.TrimSpace(e) == "" {
			return errors.New("los elementos no pueden estar vacíos")
		}
	}
	return nil
}
