package models

import (
	"testing"
	"time"
)

func actaBase(tipo TipoActa) Acta {
	return Acta{
		Fecha:            time.Date(2026, time.August, 31, 0, 0, 0, 0, time.UTC),
		Tipo:             tipo,
		AreaDepartamento: "Sistemas",
	}
}

func elementoRetira() Elemento {
	return Elemento{Descripcion: "Notebook", Cantidad: 1, Direccion: DireccionRetira}
}

func elementoEntrega() Elemento {
	return Elemento{Descripcion: "Impresora", Cantidad: 1, Direccion: DireccionEntrega}
}

func TestValidate_Entrega(t *testing.T) {
	t.Run("válida sin QuienRecibe", func(t *testing.T) {
		acta := actaBase(TipoActaEntrega)
		acta.QuienEntrega = "Juan Pérez"
		acta.Elementos = []Elemento{elementoRetira()}

		if err := acta.Validate(); err != nil {
			t.Errorf("no debería exigir QuienRecibe en tipo entrega, error: %v", err)
		}
	})

	t.Run("inválida sin QuienEntrega", func(t *testing.T) {
		acta := actaBase(TipoActaEntrega)
		acta.Elementos = []Elemento{elementoRetira()}

		if err := acta.Validate(); err == nil {
			t.Error("debería exigir QuienEntrega en tipo entrega")
		}
	})

	t.Run("inválida con elemento de dirección entrega", func(t *testing.T) {
		acta := actaBase(TipoActaEntrega)
		acta.QuienEntrega = "Juan Pérez"
		acta.Elementos = []Elemento{elementoEntrega()}

		if err := acta.Validate(); err == nil {
			t.Error("debería rechazar elementos de dirección entrega en tipo entrega")
		}
	})
}

func TestValidate_Recibimiento(t *testing.T) {
	t.Run("válida sin QuienEntrega", func(t *testing.T) {
		acta := actaBase(TipoActaRecibimiento)
		acta.QuienRecibe = "María Gómez"
		acta.Elementos = []Elemento{elementoEntrega()}

		if err := acta.Validate(); err != nil {
			t.Errorf("no debería exigir QuienEntrega en tipo recibimiento, error: %v", err)
		}
	})

	t.Run("inválida sin QuienRecibe", func(t *testing.T) {
		acta := actaBase(TipoActaRecibimiento)
		acta.Elementos = []Elemento{elementoEntrega()}

		if err := acta.Validate(); err == nil {
			t.Error("debería exigir QuienRecibe en tipo recibimiento")
		}
	})

	t.Run("inválida con elemento de dirección retira", func(t *testing.T) {
		acta := actaBase(TipoActaRecibimiento)
		acta.QuienRecibe = "María Gómez"
		acta.Elementos = []Elemento{elementoRetira()}

		if err := acta.Validate(); err == nil {
			t.Error("debería rechazar elementos de dirección retira en tipo recibimiento")
		}
	})
}

func TestValidate_Ambos(t *testing.T) {
	t.Run("válida con ambas personas y ambas direcciones", func(t *testing.T) {
		acta := actaBase(TipoActaAmbos)
		acta.QuienEntrega = "Juan Pérez"
		acta.QuienRecibe = "María Gómez"
		acta.Elementos = []Elemento{elementoRetira(), elementoEntrega()}

		if err := acta.Validate(); err != nil {
			t.Errorf("acta válida de tipo ambos no debería fallar, error: %v", err)
		}
	})

	t.Run("inválida sin QuienEntrega", func(t *testing.T) {
		acta := actaBase(TipoActaAmbos)
		acta.QuienRecibe = "María Gómez"
		acta.Elementos = []Elemento{elementoRetira()}

		if err := acta.Validate(); err == nil {
			t.Error("debería exigir QuienEntrega en tipo ambos")
		}
	})

	t.Run("inválida sin QuienRecibe", func(t *testing.T) {
		acta := actaBase(TipoActaAmbos)
		acta.QuienEntrega = "Juan Pérez"
		acta.Elementos = []Elemento{elementoRetira()}

		if err := acta.Validate(); err == nil {
			t.Error("debería exigir QuienRecibe en tipo ambos")
		}
	})
}

func TestValidate_CamposSiempreObligatorios(t *testing.T) {
	t.Run("inválida sin fecha", func(t *testing.T) {
		acta := actaBase(TipoActaEntrega)
		acta.Fecha = time.Time{}
		acta.QuienEntrega = "Juan Pérez"
		acta.Elementos = []Elemento{elementoRetira()}

		if err := acta.Validate(); err == nil {
			t.Error("debería exigir fecha en cualquier tipo")
		}
	})

	t.Run("inválida sin área/departamento", func(t *testing.T) {
		acta := actaBase(TipoActaEntrega)
		acta.AreaDepartamento = ""
		acta.QuienEntrega = "Juan Pérez"
		acta.Elementos = []Elemento{elementoRetira()}

		if err := acta.Validate(); err == nil {
			t.Error("debería exigir área/departamento en cualquier tipo")
		}
	})

	t.Run("inválida sin elementos", func(t *testing.T) {
		acta := actaBase(TipoActaEntrega)
		acta.QuienEntrega = "Juan Pérez"

		if err := acta.Validate(); err == nil {
			t.Error("debería exigir al menos un elemento en cualquier tipo")
		}
	})

	t.Run("inválida con tipo desconocido", func(t *testing.T) {
		acta := actaBase(TipoActa("otro"))
		acta.QuienEntrega = "Juan Pérez"
		acta.QuienRecibe = "María Gómez"
		acta.Elementos = []Elemento{elementoRetira()}

		if err := acta.Validate(); err == nil {
			t.Error("debería rechazar un tipo de acta desconocido")
		}
	})
}
