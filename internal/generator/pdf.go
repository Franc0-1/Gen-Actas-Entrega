package generator

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const timeoutConversionPDF = 90 * time.Second

// nombreComandoSoffice ubica el ejecutable de LibreOffice. Se puede
// sobreescribir con la variable de entorno SOFFICE_PATH (usada en el
// contenedor Linux, donde soffice ya está en el PATH); si no está seteada,
// se usa un valor por defecto según el sistema operativo.
var nombreComandoSoffice = resolverComandoSoffice()

func resolverComandoSoffice() string {
	if ruta := os.Getenv("SOFFICE_PATH"); ruta != "" {
		return ruta
	}
	if runtime.GOOS == "windows" {
		// Windows no suele tener soffice en el PATH del proceso de servicio;
		// exec.LookPath acepta una ruta absoluta igual (no busca en PATH si
		// el argumento ya contiene un separador de ruta).
		return `C:\Program Files\LibreOffice\program\soffice.exe`
	}
	return "soffice"
}

// perfilLibreOfficeURI apunta a un perfil de usuario propio y persistente
// (no el de un usuario real de Windows) para los procesos de conversión.
// Evita que LibreOffice reconstruya su perfil (registro, cachés, fuentes)
// en cada arranque, que es buena parte del costo de los 3-5s observados.
// Esto reduce el overhead por conversión, pero no elimina el costo base de
// arrancar el proceso: para eso hace falta una instancia persistente con
// listener UNO, que es un cambio de arquitectura mayor a esta función.
var perfilLibreOfficeURI = "file:///" + filepath.ToSlash(filepath.Join(os.TempDir(), "actas-libreoffice-profile"))

// ConvertirAPDF convierte un .docx ya generado a PDF usando LibreOffice en
// modo headless. No modifica ni reemplaza el DOCX de entrada: el PDF se
// escribe en outputDir con el mismo nombre base. Devuelve la ruta del PDF
// generado.
func ConvertirAPDF(docxPath, outputDirPDF string) (string, error) {
	if _, err := exec.LookPath(nombreComandoSoffice); err != nil {
		return "", fmt.Errorf("LibreOffice (%s) no está instalado: %w", nombreComandoSoffice, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutConversionPDF)
	defer cancel()

	cmd := exec.CommandContext(ctx, nombreComandoSoffice,
		"--headless",
		"--nologo",
		"--nodefault",
		"--norestore",
		"-env:UserInstallation="+perfilLibreOfficeURI,
		"--convert-to", "pdf:writer_pdf_Export",
		"--outdir", outputDirPDF,
		docxPath,
	)

	salida, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("la conversión a PDF superó el tiempo máximo configurado (%s)", timeoutConversionPDF)
		}
		return "", fmt.Errorf("no se pudo convertir %s a PDF: %w (salida: %s)", docxPath, err, strings.TrimSpace(string(salida)))
	}

	nombreBase := strings.TrimSuffix(filepath.Base(docxPath), filepath.Ext(docxPath))
	pdfPath := filepath.Join(outputDirPDF, nombreBase+".pdf")

	if _, err := os.Stat(pdfPath); err != nil {
		return "", fmt.Errorf("la conversión a PDF no generó el archivo esperado (%s): %w", pdfPath, err)
	}

	return pdfPath, nil
}
