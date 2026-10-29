package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Viaje representa la estructura básica del itinerario (RF-05)
type Viaje struct {
	IDViaje string `json:"idViaje"`
	Destino string `json:"destino"`
	Estado  string `json:"estado"`
}

func main() {
	// API Gateway: Enrutador principal de Trotamundos
	mux := http.NewServeMux()

	// Endpoint síncrono para obtener itinerarios
	mux.HandleFunc("/itinerarios", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			
			// Mock de datos para demostrar el contrato
			viajes := []Viaje{
				{IDViaje: "123e4567-e89b-12d3-a456-426614174000", Destino: "Madrid", Estado: "activo"},
			}
			json.NewEncoder(w).Encode(viajes)
			return
		}
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	})

	fmt.Println("🚀 API Gateway de Trotamundos iniciando en el puerto 8080...")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
