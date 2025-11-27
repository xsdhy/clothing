package api

import "github.com/sirupsen/logrus"

type sseMessage struct {
	event string
	data  interface{}
}

func (h *HTTPHandler) registerSSEClient(clientID string, ch chan sseMessage) {
	if h == nil || ch == nil || clientID == "" {
		return
	}
	h.sseMu.Lock()
	defer h.sseMu.Unlock()

	if h.sseClients == nil {
		h.sseClients = make(map[string][]chan sseMessage)
	}
	h.sseClients[clientID] = append(h.sseClients[clientID], ch)
}

func (h *HTTPHandler) unregisterSSEClient(clientID string, target chan sseMessage) {
	if h == nil || target == nil || clientID == "" {
		return
	}
	h.sseMu.Lock()
	defer h.sseMu.Unlock()

	current := h.sseClients[clientID]
	if len(current) == 0 {
		return
	}

	remaining := current[:0]
	for _, ch := range current {
		if ch == target {
			continue
		}
		remaining = append(remaining, ch)
	}

	if len(remaining) == 0 {
		delete(h.sseClients, clientID)
		return
	}

	h.sseClients[clientID] = remaining
}

func (h *HTTPHandler) publishSSEMessage(clientID string, msg sseMessage) {
	if h == nil || clientID == "" {
		return
	}

	h.sseMu.Lock()
	channels := append([]chan sseMessage(nil), h.sseClients[clientID]...)
	h.sseMu.Unlock()

	for _, ch := range channels {
		select {
		case ch <- msg:
		default:
			logrus.WithFields(logrus.Fields{
				"client_id": clientID,
				"event":     msg.event,
			}).Warn("dropping sse message due to slow consumer")
		}
	}
}
