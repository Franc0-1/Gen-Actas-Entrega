package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strings"

	"actas-project/internal/handlers"
	"actas-project/internal/services"
)

// loadEnv lee un archivo .env simple (KEY=VALUE por línea) y setea
// las variables de entorno si no están ya definidas.
func loadEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return // .env no existe, seguimos con defaults
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, value)
		}
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// sinListado responde 404 a cualquier ruta de carpeta (terminada en "/"),
// para que http.FileServer no publique el listado de documentos generados;
// los archivos se siguen sirviendo por su ruta exacta. Una carpeta pedida
// sin "/" final la redirige FileServer a la versión con "/", que cae acá.
func sinListado(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func main() {
	loadEnv(".env")

	if err := services.PrepararDirectorios(); err != nil {
		log.Fatal(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/actas/generar", handlers.GenerarActaHandler)
	mux.HandleFunc("/api/notas-retiro/generar", handlers.GenerarNotaRetiroHandler)
	mux.Handle("/", http.FileServer(http.Dir("templates/web")))
	mux.Handle("/output/", sinListado(http.StripPrefix("/output/", http.FileServer(http.Dir("output")))))

	addr := ":" + port
	log.Printf("Servidor escuchando en http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
