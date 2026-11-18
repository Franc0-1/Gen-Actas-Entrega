package models

import (
	"errors"
	"time"
)

// TipoActa indica si el acta es de entrega, de recibimiento o de ambos.
type TipoActa string

const (
	TipoActaEntrega      TipoActa = "entrega"
	TipoActaRecibimiento TipoActa = "recibimiento"
	TipoActaAmbos        TipoActa = "ambos"
)

// esValido indica si el valor es uno de los tipos de acta soportados.
func (t TipoActa) esValido() bool {
	switch t {
	case TipoActaEntrega, TipoActaRecibimiento, TipoActaAmbos:
		return true
	default:
		return false
	}
}

// DireccionElemento indica si un elemento lo retira la persona (se lo lleva)
// o lo entrega la persona (lo devuelve a la oficina).
type DireccionElemento string

const (
	DireccionRetira  DireccionElemento = "retira"
	DireccionEntrega DireccionElemento = "entrega"
)

// esValido indica si el valor es una dirección de elemento soportada.
func (d DireccionElemento) esValido() bool {
	switch d {
	case DireccionRetira, DireccionEntrega:
		return true
	default:
		return false
	}
}

// Elemento representa un ítem individual entregado/recibido dentro de un acta.
type Elemento struct {
	Descripcion   string            `json:"descripcion"`
	NroInventario string            `json:"nroInventario,omitempty"`
	Cantidad      int               `json:"cantidad"`
	Observacion   string            `json:"observacion,omitempty"`
	Direccion     DireccionElemento `json:"direccion"`
}

// Validate verifica que el elemento tenga los datos mínimos correctos.
func (e Elemento) Validate() error {
	if e.Descripcion == "" {
		return errors.New("el elemento debe tener una descripción")
	}
	if e.Cantidad <= 0 {
		return errors.New("el elemento debe tener una cantidad mayor a cero")
	}
	if !e.Direccion.esValido() {
		return errors.New("la dirección del elemento no es válida")
	}
	if e.NroInventario != "" && !esSoloDigitos(e.NroInventario) {
		return errors.New("el número de inventario solo debe contener dígitos")
	}
	return nil
}

// esSoloDigitos indica si el string contiene únicamente caracteres 0-9.
func esSoloDigitos(s string) bool {
	if len(s) == 0 || len(s) > 8 {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// Acta representa un acta de entrega, recibimiento o ambos, con sus elementos.
type Acta struct {
	Fecha            time.Time  `json:"fecha"`
	QuienEntrega     string     `json:"quienEntrega"`
	Tipo             TipoActa   `json:"tipo"`
	AreaDepartamento string     `json:"areaDepartamento"`
	Elementos        []Elemento `json:"elementos"`
}

// Validate verifica que el acta tenga los datos mínimos necesarios para
// poder generarse. No valida reglas de negocio más allá de eso.
func (a Acta) Validate() error {
	if a.Fecha.IsZero() {
		return errors.New("el acta debe tener una fecha")
	}
	if !a.Tipo.esValido() {
		return errors.New("el tipo de acta no es válido")
	}
	if a.Fecha.After(time.Now()) {
		return errors.New("la fecha del acta no puede ser futura")
	}

	// QuienEntrega es "quién presenta": figura en el texto de todos los tipos
	// de acta ("se presenta ... por ante esta oficina"), así que siempre se exige.
	if a.QuienEntrega == "" {
		return errors.New("el acta debe indicar quién presenta")
	}

	if a.AreaDepartamento == "" {
		return errors.New("el acta debe indicar el área o departamento")
	}
	if len(a.Elementos) == 0 {
		return errors.New("el acta debe tener al menos un elemento")
	}
	for _, elemento := range a.Elementos {
		if err := elemento.Validate(); err != nil {
			return err
		}
		if a.Tipo == TipoActaEntrega && elemento.Direccion != DireccionRetira {
			return errors.New("en un acta de tipo entrega todos los elementos deben ser de dirección retira")
		}
		if a.Tipo == TipoActaRecibimiento && elemento.Direccion != DireccionEntrega {
			return errors.New("en un acta de tipo recibimiento todos los elementos deben ser de dirección entrega")
		}
	}
	return nil
}
