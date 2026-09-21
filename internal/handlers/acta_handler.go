package handlers

import (
	"encoding/json"
	"net/http"

	"actas-project/internal/models"
	"actas-project/internal/services"
)

// GenerarActaHandler atiende POST /api/actas/generar: recibe un Acta en
// JSON, lo valida, genera el DOCX y lo convierte a PDF (services.GenerarActa
// se encarga de ambos pasos), y responde con la ruta del PDF resultante.
func GenerarActaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responderJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "método no permitido"})
		return
	}

	var acta models.Acta
	if err := json.NewDecoder(r.Body).Decode(&acta); err != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON inválido: " + err.Error()})
		return
	}

	acta.AplicarValoresPorDefecto()

	if err := acta.Validate(); err != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	nombreArchivo, err := services.GenerarActa(acta)
	if err != nil {
		responderJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	responderJSON(w, http.StatusCreated, map[string]string{
		"mensaje":     "acta generada correctamente",
		"archivo":     nombreArchivo,
		"descargaUrl": "/output/PDF/" + nombreArchivo,
	})
}

func responderJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}
