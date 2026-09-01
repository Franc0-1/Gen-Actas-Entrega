package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strings"

	"actas-project/internal/handlers"
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

func main() {
	loadEnv(".env")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/actas/generar", handlers.GenerarActaHandler)
	mux.Handle("/", http.FileServer(http.Dir("templates/web")))
	mux.Handle("/output/", http.StripPrefix("/output/", http.FileServer(http.Dir("output"))))

	addr := ":" + port
	log.Printf("Servidor escuchando en http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
