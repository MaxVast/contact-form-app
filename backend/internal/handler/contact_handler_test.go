package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/maxvast/contact-form-app/backend/internal/model"
	"github.com/maxvast/contact-form-app/backend/internal/service"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeRepository struct {
	saved []model.ContactMessage
}

func TestContactHandler_Ready(t *testing.T) {
	repo := &fakeRepository{}
	svc := service.NewContactService(repo)
	handler := NewContactHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	handler.Ready(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	expected := `{"status":"ready"}`
	if strings.TrimSpace(rec.Body.String()) != expected {
		t.Fatalf("expected body %q, got %q", expected, rec.Body.String())
	}
}

func (f *fakeRepository) Ping(ctx context.Context) error {
	return nil
}

func (f *fakeRepository) Save(_ context.Context, c *model.ContactMessage) error {
	c.ID = uuid.NewString()
	f.saved = append(f.saved, *c)
	return nil
}

func (f *fakeRepository) List(_ context.Context, _ int) ([]model.ContactMessage, error) {
	return f.saved, nil
}

func newTestHandler() (*ContactHandler, *fakeRepository) {
	repo := &fakeRepository{}
	svc := service.NewContactService(repo)
	return NewContactHandler(svc), repo
}

func TestContactHandler_Create_Success(t *testing.T) {
	h, repo := newTestHandler()

	body := `{"name":"Alice","email":"alice@example.com","subject":"Devis","message":"Bonjour !"}`
	req := httptest.NewRequest(http.MethodPost, "/api/contact/", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("attendu status 201, reçu %d (body: %s)", rec.Code, rec.Body.String())
	}
	if len(repo.saved) != 1 {
		t.Fatalf("attendu 1 message sauvegardé, reçu %d", len(repo.saved))
	}

	var got model.ContactMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("réponse JSON invalide: %v", err)
	}
	if got.Name != "Alice" {
		t.Errorf("attendu name=Alice, reçu %q", got.Name)
	}
}

func TestContactHandler_Create_InvalidBody(t *testing.T) {
	h, _ := newTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/contact/", bytes.NewBufferString("pas du json"))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("attendu status 400, reçu %d", rec.Code)
	}
}

func TestContactHandler_Create_ValidationError(t *testing.T) {
	h, repo := newTestHandler()

	body := `{"name":"","email":"pas-un-email","subject":"","message":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/contact/", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("attendu status 422, reçu %d", rec.Code)
	}
	if len(repo.saved) != 0 {
		t.Error("aucun message ne devrait être sauvegardé si invalide")
	}
}

func TestContactHandler_List(t *testing.T) {
	h, repo := newTestHandler()
	repo.saved = []model.ContactMessage{{ID: "550e8400-e29b-41d4-a716-446655440000", Name: "Alice"}, {ID: "550e8400-e29b-41d4-a716-446655440001", Name: "Bob"}}

	req := httptest.NewRequest(http.MethodGet, "/api/contact/", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("attendu status 200, reçu %d", rec.Code)
	}

	var got []model.ContactMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("réponse JSON invalide: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("attendu 2 messages, reçu %d", len(got))
	}
}

func TestContactHandler_Health(t *testing.T) {
	h, _ := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	h.Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("attendu status 200, reçu %d", rec.Code)
	}
}
