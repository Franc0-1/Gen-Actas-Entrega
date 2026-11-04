package models

import (
	"testing"
	"time"
)

func TestNotaRetiro_Validate(t *testing.T) {
	valida := NotaRetiro{
		Fecha:     time.Date(2026, time.October, 27, 0, 0, 0, 0, time.UTC),
		Empresa:   "Bienes y Servicios 3M",
		Elementos: []string{"Impresora Pantum M6550NW"},
	}
	if err := valida.Validate(); err != nil {
		t.Fatalf("nota válida no debería fallar: %v", err)
	}

	casos := map[string]func(*NotaRetiro){
		"sin fecha":      func(n *NotaRetiro) { n.Fecha = time.Time{} },
		"sin empresa":    func(n *NotaRetiro) { n.Empresa = "  " },
		"sin elementos":  func(n *NotaRetiro) { n.Elementos = nil },
		"elemento vacío": func(n *NotaRetiro) { n.Elementos = []string{""} },
	}
	for nombre, romper := range casos {
		n := valida
		romper(&n)
		if err := n.Validate(); err == nil {
			t.Errorf("%s: debería fallar", nombre)
		}
	}
}
