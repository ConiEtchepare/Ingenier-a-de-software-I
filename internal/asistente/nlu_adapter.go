package asistente

import "strings"

type IntentResult struct {
	Intencion string
	Entidad   string
	Confianza float64
}

// IntentClassifier es el contrato interno encapsulado (ADR-12) que permite cambiar de motor NLU/LLM sin tocar el orquestador
type IntentClassifier interface {
	Classify(texto string) IntentResult
}

// LocalRuleNLUAdapter es la implementación determinística de baja latencia (< 5 ms) para el prototipo
type LocalRuleNLUAdapter struct{}

func NewLocalRuleNLUAdapter() *LocalRuleNLUAdapter {
	return &LocalRuleNLUAdapter{}
}

func (a *LocalRuleNLUAdapter) Classify(texto string) IntentResult {
	t := strings.ToLower(strings.TrimSpace(texto))

	// 1. Casos complejos que requieren intervención de la agencia (RF-08 / RF-06)
	palabrasComplejas := []string{"cancelar", "perdi", "perdí", "reembolso", "emergencia", "tormenta", "reclamo", "agente", "humano", "agencia", "conexion", "conexión"}
	for _, p := range palabrasComplejas {
		if strings.Contains(t, p) {
			return IntentResult{
				Intencion: "ESCALAR_AGENCIA",
				Entidad:   p,
				Confianza: 0.94,
			}
		}
	}

	// 2. Reprogramaciones menores (check-in tardío, cambio horario simple)
	palabrasReprogramacion := []string{"reprogramar", "check-in", "checkin", "horario", "tarde", "mover", "cambiar hora"}
	for _, p := range palabrasReprogramacion {
		if strings.Contains(t, p) {
			return IntentResult{
				Intencion: "REPROGRAMAR_MENOR",
				Entidad:   "Late Check-in / Ajuste horario +2hs",
				Confianza: 0.88,
			}
		}
	}

	// 3. Consultas de estado de viaje o reservas
	palabrasEstado := []string{"estado", "vuelo", "hotel", "puerta", "sale", "reserva", "itinerario", "donde", "dónde", "cuando", "cuándo"}
	for _, p := range palabrasEstado {
		if strings.Contains(t, p) {
			return IntentResult{
				Intencion: "CONSULTAR_ESTADO",
				Entidad:   "estado_general",
				Confianza: 0.95,
			}
		}
	}

	// 4. Baja confianza (< 0.75) -> dispara política de aclaración / escalamiento según ADR-13
	return IntentResult{
		Intencion: "DESCONOCIDA",
		Entidad:   "",
		Confianza: 0.45,
	}
}
