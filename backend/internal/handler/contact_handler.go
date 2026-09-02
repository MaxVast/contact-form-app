package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/maxvast/contact-form-app/backend/internal/model"
	"github.com/maxvast/contact-form-app/backend/internal/service"
)

type ContactHandler struct {
	service *service.ContactService
}

func NewContactHandler(s *service.ContactService) *ContactHandler {
	return &ContactHandler{service: s}
}

func (h *ContactHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.service.Ping(ctx); err != nil {
		http.Error(w, "database unavailable", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, _ = w.Write([]byte(`{"status":"ready"}`))
}

func (h *ContactHandler) Create(w http.ResponseWriter, r *http.Request) {
	var msg model.ContactMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		writeError(w, http.StatusBadRequest, "corps de requête invalide")
		return
	}

	if err := h.service.Submit(r.Context(), &msg); err != nil {
		if errors.Is(err, model.ErrValidation) {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		log.Printf("erreur enregistrement contact: %v", err)
		writeError(w, http.StatusInternalServerError, "erreur interne")
		return
	}

	writeJSON(w, http.StatusCreated, msg)
}

func (h *ContactHandler) List(w http.ResponseWriter, r *http.Request) {
	messages, err := h.service.Recent(r.Context())
	if err != nil {
		log.Printf("erreur lecture contacts: %v", err)
		writeError(w, http.StatusInternalServerError, "erreur interne")
		return
	}
	writeJSON(w, http.StatusOK, messages)
}

func (h *ContactHandler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
