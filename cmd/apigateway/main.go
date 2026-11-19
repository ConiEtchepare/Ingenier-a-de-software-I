package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"trotamundos/internal/agencias"
	"trotamundos/internal/asistente"
	"trotamundos/internal/core"
)

func main() {
	// Instanciación de componentes arquitectónicos según ADD (Iteraciones 1 a 6)
	coreService := core.NewItinerarioService()
	agenciaRouter := agencias.NewRouter("internal/config/agencias.json")
	sessionCache := asistente.NewSessionCache(30 * time.Minute)
	nluAdapter := asistente.NewLocalRuleNLUAdapter()
	botOrquestador := asistente.NewOrquestador(sessionCache, nluAdapter, coreService, agenciaRouter)

	mux := http.NewServeMux()

	// Servir interfaz web del prototipo
	mux.Handle("/", http.FileServer(http.Dir("./web")))

	// 1. Endpoint GET/POST /itinerarios (RF-04, RF-05)
	mux.HandleFunc("/itinerarios", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(coreService.ListarViajes())
		case http.MethodPost:
			var req core.ReservaManualRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"JSON inválido"}`, http.StatusBadRequest)
				return
			}
			viajeActualizado := coreService.AgregarReservaManual(req)
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(viajeActualizado)
		default:
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		}
	})

	// 2. Endpoint POST /itinerarios/{idViaje}/derivar (RF-06 - Iteración 4)
	mux.HandleFunc("/itinerarios/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/derivar") {
			http.Error(w, "Ruta o método no soportado", http.StatusNotFound)
			return
		}
		idViaje := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/itinerarios/"), "/derivar")
		viaje, ok := coreService.ObtenerViaje(idViaje)
		if !ok {
			http.Error(w, `{"error":"Viaje no encontrado"}`, http.StatusNotFound)
			return
		}
		res := agenciaRouter.DerivarItinerario(viaje.IDViaje, viaje.AgenciaPreferida, "Derivación manual desde panel", "")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	})

	// 3. Endpoint POST /assistant/messages (RF-08 - Iteraciones 5 y 6)
	mux.HandleFunc("/assistant/messages", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
			return
		}
		var req asistente.MensajeUsuarioRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Mensaje) == "" {
			http.Error(w, `{"error":"El campo 'mensaje' es obligatorio"}`, http.StatusBadRequest)
			return
		}
		resp := botOrquestador.ProcesarMensaje(req)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// 4. Endpoint GET /assistant/sessions/{sessionId} (ADR-11)
	mux.HandleFunc("/assistant/sessions/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
			return
		}
		sessionID := strings.TrimPrefix(r.URL.Path, "/assistant/sessions/")
		sesion, ok := botOrquestador.ObtenerSesion(sessionID)
		if !ok {
			http.Error(w, `{"error":"Sesión expirada o inexistente"}`, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sesion)
	})

	fmt.Println("🚀 API Gateway de Trotamundos (v2.0 con Asistente Virtual) corriendo en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
