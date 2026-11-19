package core

import (
	"fmt"
	"strings"
	"sync"
)

type Reserva struct {
	Codigo        string `json:"codigo"`
	Tipo          string `json:"tipo"`
	Detalle       string `json:"detalle"`
	FechaHora     string `json:"fechaHora"`
	EstadoReserva string `json:"estadoReserva"`
}

type Viaje struct {
	IDViaje          string    `json:"idViaje"`
	Destino          string    `json:"destino"`
	Estado           string    `json:"estado"`
	AgenciaPreferida string    `json:"agenciaPreferida"`
	Reservas         []Reserva `json:"reservas"`
}

type ReservaManualRequest struct {
	Destino     string `json:"destino"`
	Tipo        string `json:"tipo"`
	Detalle     string `json:"detalle"`
	FechaInicio string `json:"fechaInicio"`
}

// ItinerarioService maneja la lógica central de agrupación y gestión manual (RF-04, RF-05)
type ItinerarioService struct {
	mu     sync.RWMutex
	viajes []Viaje
}

func NewItinerarioService() *ItinerarioService {
	return &ItinerarioService{
		viajes: []Viaje{
			{
				IDViaje:          "123e4567-e89b-12d3-a456-426614174000",
				Destino:          "Madrid",
				Estado:           "activo",
				AgenciaPreferida: "AgenciaEjemplo",
				Reservas: []Reserva{
					{
						Codigo:        "IB-6842",
						Tipo:          "vuelo",
						Detalle:       "Iberia EZE -> MAD (Puerta B14)",
						FechaHora:     "2026-12-10 21:30",
						EstadoReserva: "En horario",
					},
					{
						Codigo:        "HT-901",
						Tipo:          "hotel",
						Detalle:       "Hotel Gran Vía Madrid - Check-in 15:00",
						FechaHora:     "2026-12-11 15:00",
						EstadoReserva: "Confirmada",
					},
				},
			},
			{
				IDViaje:          "987fcdeb-51a2-43d1-9012-345678901234",
				Destino:          "Bariloche",
				Estado:           "activo",
				AgenciaPreferida: "DespegarCorporate",
				Reservas: []Reserva{
					{
						Codigo:        "AR-1680",
						Tipo:          "vuelo",
						Detalle:       "Aerolíneas Argentinas AEP -> BRC",
						FechaHora:     "2027-01-15 08:15",
						EstadoReserva: "Confirmado",
					},
				},
			},
		},
	}
}

func (s *ItinerarioService) ListarViajes() []Viaje {
	s.mu.RLock()
	defer s.mu.RUnlock()
	copia := make([]Viaje, len(s.viajes))
	copy(copia, s.viajes)
	return copia
}

func (s *ItinerarioService) ObtenerViaje(id string) (Viaje, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, v := range s.viajes {
		if v.IDViaje == id {
			return v, true
		}
	}
	if len(s.viajes) > 0 && id == "" {
		return s.viajes[0], true
	}
	return Viaje{}, false
}

func (s *ItinerarioService) AgregarReservaManual(req ReservaManualRequest) Viaje {
	s.mu.Lock()
	defer s.mu.Unlock()

	nuevaReserva := Reserva{
		Codigo:        fmt.Sprintf("MAN-%d", len(s.viajes)*100+15),
		Tipo:          req.Tipo,
		Detalle:       req.Detalle,
		FechaHora:     req.FechaInicio,
		EstadoReserva: "Confirmada (Manual)",
	}

	for i := range s.viajes {
		if strings.EqualFold(s.viajes[i].Destino, req.Destino) {
			s.viajes[i].Reservas = append(s.viajes[i].Reservas, nuevaReserva)
			return s.viajes[i]
		}
	}

	nuevoViaje := Viaje{
		IDViaje:          fmt.Sprintf("viaje-%d", len(s.viajes)+100),
		Destino:          req.Destino,
		Estado:           "activo",
		AgenciaPreferida: "AgenciaEjemplo",
		Reservas:         []Reserva{nuevaReserva},
	}
	s.viajes = append(s.viajes, nuevoViaje)
	return nuevoViaje
}

// ReprogramarMenor ejecuta cambios simples permitidos en primer nivel (ej. horario de check-in o asiento)
func (s *ItinerarioService) ReprogramarMenor(idViaje, nuevoHorario string) (Viaje, string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.viajes {
		if s.viajes[i].IDViaje == idViaje || (idViaje == "" && i == 0) {
			if len(s.viajes[i].Reservas) > 0 {
				res := &s.viajes[i].Reservas[len(s.viajes[i].Reservas)-1]
				res.EstadoReserva = fmt.Sprintf("Reprogramado (%s)", nuevoHorario)
				return s.viajes[i], fmt.Sprintf("Se actualizó la reserva %s (%s) a: %s.", res.Codigo, res.Detalle, nuevoHorario), true
			}
		}
	}
	return Viaje{}, "", false
}
