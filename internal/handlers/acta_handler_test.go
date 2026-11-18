package handlers

import (
	"net/url"
	"testing"
)

func TestURLDescarga(t *testing.T) {
	for _, nombre := range []string{
		"Acta de Entrega - Prensa - 2026-11-18.pdf",
		"Acta de Entrega - Depto #3 - 2026-11-18.pdf",
		"Nota de Retiro - 100% Servicios - 2026-11-18.pdf",
		"Acta de Entrega - Comunicación - 2026-11-18 (2).pdf",
	} {
		ruta := urlDescarga(nombre)
		// Lo que el navegador entiende como ruta (sin fragmento ni query) tiene
		// que ser exactamente el archivo pedido.
		u, err := url.Parse(ruta)
		if err != nil {
			t.Fatalf("%q: URL inválida %q: %v", nombre, ruta, err)
		}
		if u.Fragment != "" || u.RawQuery != "" || u.Path != "/output/PDF/"+nombre {
			t.Errorf("%q: la URL %q se interpreta como ruta %q (fragmento %q)", nombre, ruta, u.Path, u.Fragment)
		}
	}
}
