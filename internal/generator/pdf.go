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

// nombreComandoUnoconv ubica el cliente unoconv, usado en Linux/Docker para
// delegar la conversión al listener UNO persistente (ver
// docker/supervisord.conf, puerto 127.0.0.1:2002) en vez de arrancar una
// instancia de soffice por cada conversión. Se puede sobreescribir con la
// variable de entorno UNOCONV_PATH.
var nombreComandoUnoconv = resolverComandoUnoconv()

func resolverComandoUnoconv() string {
	if ruta := os.Getenv("UNOCONV_PATH"); ruta != "" {
		return ruta
	}
	return "unoconv"
}

// ConvertirAPDF convierte un .docx ya generado a PDF. No modifica ni
// reemplaza el DOCX de entrada. Devuelve la ruta del PDF generado.
//
// En Linux (contenedor Docker) reutiliza el listener UNO persistente vía
// unoconv, evitando el costo de arrancar soffice en cada conversión. En
// Windows (desarrollo local) arranca una instancia de soffice puntual, como
// antes, porque no hay listener persistente corriendo ahí.
func ConvertirAPDF(docxPath, outputDirPDF string) (string, error) {
	nombreBase := strings.TrimSuffix(filepath.Base(docxPath), filepath.Ext(docxPath))
	pdfPath := filepath.Join(outputDirPDF, nombreBase+".pdf")

	var err error
	if runtime.GOOS == "windows" {
		err = convertirConSofficeDirecto(docxPath, outputDirPDF)
	} else {
		err = convertirConUnoconv(docxPath, pdfPath)
	}
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(pdfPath); err != nil {
		return "", fmt.Errorf("la conversión a PDF no generó el archivo esperado (%s): %w", pdfPath, err)
	}

	return pdfPath, nil
}

// convertirConUnoconv delega la conversión a la instancia de soffice ya
// activa (listener UNO en 127.0.0.1:2002) a través de unoconv, sin arrancar
// un soffice nuevo.
func convertirConUnoconv(docxPath, pdfPath string) error {
	if _, err := exec.LookPath(nombreComandoUnoconv); err != nil {
		return fmt.Errorf("unoconv (%s) no está instalado: %w", nombreComandoUnoconv, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutConversionPDF)
	defer cancel()

	cmd := exec.CommandContext(ctx, nombreComandoUnoconv,
		"-f", "pdf",
		"-o", pdfPath,
		docxPath,
	)

	salida, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("la conversión a PDF superó el tiempo máximo configurado (%s)", timeoutConversionPDF)
		}
		return fmt.Errorf("no se pudo convertir %s a PDF: %w (salida: %s)", docxPath, err, strings.TrimSpace(string(salida)))
	}
	return nil
}

// convertirConSofficeDirecto arranca una instancia de soffice puntual para
// la conversión (modo de desarrollo local en Windows, sin listener UNO
// persistente).
func convertirConSofficeDirecto(docxPath, outputDirPDF string) error {
	if _, err := exec.LookPath(nombreComandoSoffice); err != nil {
		return fmt.Errorf("LibreOffice (%s) no está instalado: %w", nombreComandoSoffice, err)
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
			return fmt.Errorf("la conversión a PDF superó el tiempo máximo configurado (%s)", timeoutConversionPDF)
		}
		return fmt.Errorf("no se pudo convertir %s a PDF: %w (salida: %s)", docxPath, err, strings.TrimSpace(string(salida)))
	}
	return nil
}
