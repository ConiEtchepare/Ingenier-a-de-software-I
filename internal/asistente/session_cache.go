package asistente

import (
	"sync"
	"time"
)

type MensajeChat struct {
	Emisor    string `json:"emisor"`
	Texto     string `json:"texto"`
	Timestamp string `json:"timestamp"`
}

type SesionChat struct {
	SessionID     string        `json:"sessionId"`
	IDViajeActivo string        `json:"idViajeActivo"`
	Mensajes      []MensajeChat `json:"mensajes"`
	ExpiraEn      time.Time     `json:"-"`
}

// SessionCache implementa almacenamiento efímero en memoria con TTL de 30 minutos (ADR-11)
type SessionCache struct {
	mu       sync.RWMutex
	sesiones map[string]SesionChat
	ttl      time.Duration
}

func NewSessionCache(ttl time.Duration) *SessionCache {
	return &SessionCache{
		sesiones: make(map[string]SesionChat),
		ttl:      ttl,
	}
}

func (c *SessionCache) Obtener(sessionID string) (SesionChat, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	s, ok := c.sesiones[sessionID]
	if !ok || time.Now().After(s.ExpiraEn) {
		return SesionChat{}, false
	}
	return s, true
}

func (c *SessionCache) Guardar(s SesionChat) {
	c.mu.Lock()
	defer c.mu.Unlock()
	s.ExpiraEn = time.Now().Add(c.ttl)
	c.sesiones[s.SessionID] = s
}
