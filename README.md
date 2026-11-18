# Generador de Actas de Entrega

Aplicación web interna de la Oficina de Sistemas para generar en PDF:

- **Actas** de entrega, recibimiento o ambos, con su tabla de elementos.
- **Notas de retiro** dirigidas a la Guardia Policial de Casa de Gobierno.

El backend (Go, sin dependencias externas) completa una plantilla DOCX y la
convierte a PDF con LibreOffice. El frontend es HTML, CSS y JavaScript sin
frameworks.

## Ejecutar con Docker (producción)

```bash
docker compose up -d --build
```

- `actas`: el servidor Go más LibreOffice en modo listener (supervisord).
- `caddy`: proxy HTTPS delante del servidor; la IP del host está en `Caddyfile`.
- Los documentos generados quedan en `./output` (montado como volumen).

## Ejecutar en local

Requisitos: Go 1.27 y LibreOffice instalado.

```bash
go run ./cmd/server
```

La app queda en http://localhost:8080. Se ejecuta desde la raíz del
repositorio porque las rutas a `templates/` y `output/` son relativas.

| Variable | Uso | Por defecto |
|---|---|---|
| `PORT` | Puerto HTTP | `8080` |
| `SOFFICE_PATH` | Ejecutable de LibreOffice (Windows) | `C:\Program Files\LibreOffice\program\soffice.exe` |
| `UNOCONV_PATH` | Cliente unoconv (Linux/Docker) | `unoconv` |

Se pueden definir en un archivo `.env` en la raíz (no se versiona).

## Tests

```bash
go test ./...
```

## Estructura

```
cmd/server/          servidor HTTP y rutas
internal/handlers/   endpoints JSON
internal/services/   nombres de archivo, carpetas de salida y orquestación
internal/generator/  armado del DOCX, conversión a PDF y fechas en palabras
internal/models/     acta, nota de retiro y sus validaciones
templates/docx/      plantillas usadas por el generador
templates/web/       frontend (index.html, app.js, style.css)
docs/referencias/    actas de ejemplo hechas a mano (no las usa el código)
docker/              configuración de supervisord
```

## API

Ambos endpoints reciben JSON por `POST` y responden `201` con
`{"mensaje", "archivo", "descargaUrl"}`, o un código de error con `{"error"}`.

### `POST /api/actas/generar`

```json
{
  "fecha": "2026-11-04T00:00:00Z",
  "tipo": "entrega | recibimiento | ambos",
  "quienEntrega": "Fiduciaria",
  "areaDepartamento": "Prensa",
  "elementos": [
    {
      "descripcion": "Monitor 24\"",
      "nroInventario": "12345",
      "cantidad": 1,
      "direccion": "retira | entrega",
      "observacion": ""
    }
  ]
}
```

- `retira`: el área se lleva el elemento; `entrega`: el área lo devuelve.
- Entrega solo admite `retira` y recibimiento solo `entrega`.
- `quienEntrega` (quién presenta) es obligatorio en todos los tipos y aparece
  en el texto: "se presenta <quienEntrega> por ante esta oficina de Sistemas".

### `POST /api/notas-retiro/generar`

```json
{
  "fecha": "2026-11-04T00:00:00Z",
  "empresa": "Bienes y Servicios 3M",
  "elementos": ["Impresora Pantum M6550NW nro. de serie CJ3X001670"]
}
```

### Otras rutas

- `GET /health`: estado del servicio.
- `GET /output/PDF/<archivo>.pdf`: descarga de un documento generado
  (no se listan las carpetas).

## Plantillas

| Plantilla | Marcadores |
|---|---|
| `plantilla_acta.docx` | `{{dia_texto}}`, `{{mes_texto}}`, `{{anio_texto}}`, `{{quien_presenta}}`, `{{tipo_texto}}`, `{{area_departamento}}`; en las tablas "Retira" y "Entrega": `{{descripcion}}`, `{{nro_inventario}}`, `{{cantidad}}`, `{{observacion}}` |
| `plantilla_retiro.docx` | `{{fecha_larga}}`, `{{empresa}}`, `{{detalle_elementos}}` |

Al editar una plantilla en Word, cada marcador tiene que quedar escrito de
corrido; si Word lo parte en varios fragmentos, el generador no lo reemplaza.
Los títulos "Retira" y "Entrega" ubican las tablas y no deben cambiarse.
