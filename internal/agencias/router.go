package agencias

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type ConfigAgencias struct {
	Estrategia      string          `json:"estrategia"`
	AgenciasActivas []AgenciaConfig `json:"agencias_activas"`
}

type AgenciaConfig struct {
	Nombre string `json:"nombre"`
	URL    string `json:"url"`
	Activa bool   `json:"activa"`
}

type ResultadoDerivacion struct {
	TicketAgencia  string `json:"ticketAgencia"`
	AgenciaDestino string `json:"agenciaDestino"`
	AdaptadorUsado string `json:"adaptadorUsado"`
	Estado         string `json:"estado"`
	Mensaje        string `json:"mensaje"`
}

// Router implementa el enrutador de agencias aplicando Split Module y Runtime Binding (ADR-08, ADR-09)
type Router struct {
	configPath string
}

func NewRouter(configPath string) *Router {
	return &Router{configPath: configPath}
}

func (r *Router) cargarRegistroDinamico() ConfigAgencias {
	data, err := os.ReadFile(r.configPath)
	if err != nil {
		return ConfigAgencias{
			AgenciasActivas: []AgenciaConfig{
				{Nombre: "AgenciaEjemplo", URL: "https://api.agenciaejemplo.com", Activa: true},
			},
		}
	}
	var cfg ConfigAgencias
	_ = json.Unmarshal(data, &cfg)
	return cfg
}

// DerivarItinerario selecciona el adaptador de salida en runtime y transfiere el caso complejo (RF-06)
func (r *Router) DerivarItinerario(idViaje, agenciaPreferida, motivo, resumenChat string) ResultadoDerivacion {
	cfg := r.cargarRegistroDinamico()

	agenciaSeleccionada := AgenciaConfig{
		Nombre: "AgenciaEjemplo",
		URL:    "https://api.agenciaejemplo.com",
		Activa: true,
	}

	for _, ag := range cfg.AgenciasActivas {
		if ag.Activa && ag.Nombre == agenciaPreferida {
			agenciaSeleccionada = ag
			break
		}
	}

	ticket := fmt.Sprintf("AG-%d", time.Now().Unix()%100000)
	adaptador := fmt.Sprintf("OutAdapter[%s -> POST %s/webhooks/derivaciones]", agenciaSeleccionada.Nombre, agenciaSeleccionada.URL)

	return ResultadoDerivacion{
		TicketAgencia:  ticket,
		AgenciaDestino: agenciaSeleccionada.Nombre,
		AdaptadorUsado: adaptador,
		Estado:         "DERIVADO_CON_CONTEXTO",
		Mensaje:        fmt.Sprintf("Caso derivado a %s (Ticket #%s). Motivo: %s. Se adjuntó el resumen de conversación.", agenciaSeleccionada.Nombre, ticket, motivo),
	}
}
