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
	return nil
}

// Acta representa un acta de entrega, recibimiento o ambos, con sus elementos.
type Acta struct {
	Fecha            time.Time  `json:"fecha"`
	QuienEntrega     string     `json:"quienEntrega"`
	QuienRecibe      string     `json:"quienRecibe"`
	Tipo             TipoActa   `json:"tipo"`
	AreaDepartamento string     `json:"areaDepartamento"`
	Elementos        []Elemento `json:"elementos"`
}

// valorPorDefectoQuienEntrega se asigna cuando la petición no informa
// QuienEntrega en una acta que lo requiere: la oficina que entrega equipos
// es siempre la misma.
const valorPorDefectoQuienEntrega = "Oficina de Sistemas"

// AplicarValoresPorDefecto completa QuienEntrega con el valor institucional
// fijo cuando viene vacío y el tipo de acta lo requiere (entrega o ambos).
// Debe llamarse antes de Validate().
func (a *Acta) AplicarValoresPorDefecto() {
	if a.QuienEntrega == "" && (a.Tipo == TipoActaEntrega || a.Tipo == TipoActaAmbos) {
		a.QuienEntrega = valorPorDefectoQuienEntrega
	}
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

	// Quién entrega y quién recibe se exigen según el tipo de acta: una
	// entrega pura no necesita quién recibe, y viceversa.
	switch a.Tipo {
	case TipoActaEntrega:
		if a.QuienEntrega == "" {
			return errors.New("el acta debe indicar quién entrega")
		}
	case TipoActaRecibimiento:
		if a.QuienRecibe == "" {
			return errors.New("el acta debe indicar quién recibe")
		}
	case TipoActaAmbos:
		if a.QuienEntrega == "" {
			return errors.New("el acta debe indicar quién entrega")
		}
		if a.QuienRecibe == "" {
			return errors.New("el acta debe indicar quién recibe")
		}
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
