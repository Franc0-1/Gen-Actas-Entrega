package generator

import (
	"archive/zip"
	"fmt"
	"io"
	"os"

	"actas-project/internal/models"
)

const nombreDocumentoXML = "word/document.xml"

// Generate arma un DOCX a partir de la plantilla ubicada en templatePath,
// reemplazando los datos variables por los del acta, y lo guarda en
// outputPath. No modifica la plantilla original.
func Generate(acta models.Acta, templatePath string, outputPath string) error {
	if err := acta.Validate(); err != nil {
		return fmt.Errorf("acta inválida: %w", err)
	}

	r, err := zip.OpenReader(templatePath)
	if err != nil {
		return fmt.Errorf("no se pudo abrir la plantilla: %w", err)
	}
	defer r.Close()

	archivos := map[string][]byte{}
	orden := make([]string, 0, len(r.File))
	var docXML string

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("no se pudo leer %s de la plantilla: %w", f.Name, err)
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return fmt.Errorf("no se pudo leer %s de la plantilla: %w", f.Name, err)
		}

		orden = append(orden, f.Name)
		if f.Name == nombreDocumentoXML {
			docXML = string(data)
		} else {
			archivos[f.Name] = data
		}
	}

	if docXML == "" {
		return fmt.Errorf("la plantilla no contiene %s", nombreDocumentoXML)
	}

	docXML = aplicarPlaceholdersSimples(docXML, acta)

	docXML, err = aplicarTablas(docXML, acta.Elementos)
	if err != nil {
		return fmt.Errorf("no se pudieron armar las tablas: %w", err)
	}

	archivos[nombreDocumentoXML] = []byte(docXML)

	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("no se pudo crear el archivo de salida: %w", err)
	}
	defer out.Close()

	zw := zip.NewWriter(out)
	for _, nombre := range orden {
		w, err := zw.Create(nombre)
		if err != nil {
			return fmt.Errorf("no se pudo escribir %s en el documento generado: %w", nombre, err)
		}
		if _, err := w.Write(archivos[nombre]); err != nil {
			return fmt.Errorf("no se pudo escribir %s en el documento generado: %w", nombre, err)
		}
	}

	return zw.Close()
}
