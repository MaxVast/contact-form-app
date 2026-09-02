package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/maxvast/contact-form-app/backend/internal/model"
)

type fakeRepository struct {
	saved     []model.ContactMessage
	saveErr   error
	listErr   error
	listReply []model.ContactMessage
}

func (f *fakeRepository) Ping(ctx context.Context) error {
	return nil
}

func (f *fakeRepository) Save(_ context.Context, c *model.ContactMessage) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	c.ID = uuid.NewString()
	f.saved = append(f.saved, *c)
	return nil
}

func (f *fakeRepository) List(_ context.Context, _ int) ([]model.ContactMessage, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listReply, nil
}

func validMessage() *model.ContactMessage {
	return &model.ContactMessage{
		Name:    "Alice",
		Email:   "alice@example.com",
		Subject: "Devis",
		Message: "Bonjour, un devis svp.",
	}
}

func TestContactService_Submit_Valid(t *testing.T) {
	repo := &fakeRepository{}
	svc := NewContactService(repo)

	msg := validMessage()
	if err := svc.Submit(context.Background(), msg); err != nil {
		t.Fatalf("Submit() erreur inattendue: %v", err)
	}
	if len(repo.saved) != 1 {
		t.Fatalf("attendu 1 message sauvegardé, reçu %d", len(repo.saved))
	}
	if msg.ID == "" {
		t.Error("l'ID aurait dû être renseigné après sauvegarde")
	}
}

func TestContactService_Submit_InvalidDoesNotHitRepo(t *testing.T) {
	repo := &fakeRepository{}
	svc := NewContactService(repo)

	msg := &model.ContactMessage{} // tous les champs vides -> invalide
	err := svc.Submit(context.Background(), msg)

	if err == nil {
		t.Fatal("attendu une erreur de validation, reçu nil")
	}
	if !errors.Is(err, model.ErrValidation) {
		t.Errorf("attendu ErrValidation, reçu: %v", err)
	}
	if len(repo.saved) != 0 {
		t.Error("le repository ne devrait pas être appelé si la validation échoue")
	}
}

func TestContactService_Submit_RepositoryError(t *testing.T) {
	repo := &fakeRepository{saveErr: errors.New("db down")}
	svc := NewContactService(repo)

	err := svc.Submit(context.Background(), validMessage())
	if err == nil {
		t.Fatal("attendu une erreur du repository, reçu nil")
	}
	if errors.Is(err, model.ErrValidation) {
		t.Error("une erreur de base de données ne doit pas être une ErrValidation")
	}
}

func TestContactService_Recent(t *testing.T) {
	want := []model.ContactMessage{{ID: "550e8400-e29b-41d4-a716-446655440000", Name: "Alice"}, {ID: "550e8400-e29b-41d4-a716-446655440001", Name: "Bob"}}
	repo := &fakeRepository{listReply: want}
	svc := NewContactService(repo)

	got, err := svc.Recent(context.Background())
	if err != nil {
		t.Fatalf("Recent() erreur inattendue: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("attendu 2 messages, reçu %d", len(got))
	}
}
