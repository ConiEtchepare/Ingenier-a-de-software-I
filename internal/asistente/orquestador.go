package asistente

import (
	"fmt"
	"strings"
	"time"

	"trotamundos/internal/agencias"
	"trotamundos/internal/core"
)

type MensajeUsuarioRequest struct {
	SessionID string `json:"sessionId"`
	IDViaje   string `json:"idViaje"`
	Mensaje   string `json:"mensaje"`
}

type RespuestaAsistente struct {
	SessionID          string                        `json:"sessionId"`
	IntencionDetectada string                        `json:"intencionDetectada"`
	ConfianzaNLU       float64                       `json:"confianzaNLU"`
	AccionEjecutada    string                        `json:"accionEjecutada"`
	Respuesta          string                        `json:"respuesta"`
	Derivacion         *agencias.ResultadoDerivacion `json:"derivacion,omitempty"`
}

// Orquestador coordina el flujo conversacional de primer nivel (Iteraciones 5 y 6)
type Orquestador struct {
	cache      *SessionCache
	nlu        IntentClassifier
	coreSvc    *core.ItinerarioService
	agenciaRtr *agencias.Router
}

func NewOrquestador(cache *SessionCache, nlu IntentClassifier, coreSvc *core.ItinerarioService, agenciaRtr *agencias.Router) *Orquestador {
	return &Orquestador{
		cache:      cache,
		nlu:        nlu,
		coreSvc:    coreSvc,
		agenciaRtr: agenciaRtr,
	}
}

func (o *Orquestador) ProcesarMensaje(req MensajeUsuarioRequest) RespuestaAsistente {
	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = fmt.Sprintf("sess-%d", time.Now().UnixNano()%1000000)
	}

	sesion, existe := o.cache.Obtener(sessionID)
	if !existe {
		sesion = SesionChat{
			SessionID:     sessionID,
			IDViajeActivo: req.IDViaje,
			Mensajes:      []MensajeChat{},
		}
	}
	if req.IDViaje != "" {
		sesion.IDViajeActivo = req.IDViaje
	}

	ahora := time.Now().Format("15:04:05")
	sesion.Mensajes = append(sesion.Mensajes, MensajeChat{
		Emisor:    "usuario",
		Texto:     req.Mensaje,
		Timestamp: ahora,
	})

	viaje, encontrado := o.coreSvc.ObtenerViaje(sesion.IDViajeActivo)
	if encontrado {
		sesion.IDViajeActivo = viaje.IDViaje
	}

	// 1. Clasificación de intención vía Adaptador NLU (ADR-12)
	resultadoNLU := o.nlu.Classify(req.Mensaje)

	var resp RespuestaAsistente
	resp.SessionID = sessionID
	resp.IntencionDetectada = resultadoNLU.Intencion
	resp.ConfianzaNLU = resultadoNLU.Confianza

	// 2. Política de decisión y escalamiento (ADR-13: umbral confianza < 0.75 o caso complejo)
	switch {
	case resultadoNLU.Intencion == "CONSULTAR_ESTADO" && resultadoNLU.Confianza >= 0.75:
		resp.AccionEjecutada = "CONSULTA_RESUELTA"
		if !encontrado {
			resp.Respuesta = "No encontré viajes activos asociados a tu cuenta en este momento."
		} else {
			var detalles []string
			for _, r := range viaje.Reservas {
				detalles = append(detalles, fmt.Sprintf("• [%s] %s (%s) — Estado: %s", r.Codigo, r.Detalle, r.FechaHora, r.EstadoReserva))
			}
			resp.Respuesta = fmt.Sprintf("Tu viaje a **%s** (%s) tiene las siguientes reservas:\n%s", viaje.Destino, viaje.Estado, strings.Join(detalles, "\n"))
		}

	case resultadoNLU.Intencion == "REPROGRAMAR_MENOR" && resultadoNLU.Confianza >= 0.75:
		resp.AccionEjecutada = "REPROGRAMACION_MENOR"
		_, detalleCambio, ok := o.coreSvc.ReprogramarMenor(sesion.IDViajeActivo, resultadoNLU.Entidad)
		if ok {
			resp.Respuesta = fmt.Sprintf("Listo, gestioné el ajuste menor en primer nivel: %s", detalleCambio)
		} else {
			resp.Respuesta = "No pude aplicar la reprogramación menor porque no hay reservas modificables en este viaje."
		}

	default:
		// Caso complejo o confianza NLU < 0.75 -> Escalamiento automático al Enrutador de Agencias (Iteración 4 + 6)
		resp.AccionEjecutada = "ESCALADO_AGENCIA"
		resumen := o.construirResumenSesion(sesion)
		derivacion := o.agenciaRtr.DerivarItinerario(viaje.IDViaje, viaje.AgenciaPreferida, req.Mensaje, resumen)
		resp.Derivacion = &derivacion
		resp.Respuesta = fmt.Sprintf(
			"Este caso requiere gestión especializada (Intención: %s, Confianza: %.2f). Lo derivé automáticamente a tu agencia preferida **%s** (Ticket **#%s**) junto con el contexto de nuestra charla para que no tengas que repetir nada.",
			resultadoNLU.Intencion, resultadoNLU.Confianza, derivacion.AgenciaDestino, derivacion.TicketAgencia,
		)
	}

	sesion.Mensajes = append(sesion.Mensajes, MensajeChat{
		Emisor:    "asistente",
		Texto:     resp.Respuesta,
		Timestamp: time.Now().Format("15:04:05"),
	})
	o.cache.Guardar(sesion)

	return resp
}

func (o *Orquestador) ObtenerSesion(sessionID string) (SesionChat, bool) {
	return o.cache.Obtener(sessionID)
}

func (o *Orquestador) construirResumenSesion(s SesionChat) string {
	var lineas []string
	for _, m := range s.Mensajes {
		lineas = append(lineas, fmt.Sprintf("[%s] %s: %s", m.Timestamp, m.Emisor, m.Texto))
	}
	return strings.Join(lineas, " | ")
}
