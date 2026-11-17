package httpapi

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	dataBytes, err := json.Marshal(data)
	if err != nil {
		h.log.Error("marshal response", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte(`{"error": "failed to marshal response"}`)); err != nil {
			h.log.Error("write response", "err", err)
		}
		return
	}
	w.WriteHeader(status)
	if _, err := w.Write(dataBytes); err != nil {
		h.log.Error("write response", "err", err)
	}
}

func (h *Handler) writeError(w http.ResponseWriter, status int, message string) {
	errorResponse := map[string]string{"error": message}
	h.writeJSON(w, status, errorResponse)
}
