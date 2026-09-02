package model

import (
	"errors"
	"testing"
	"time"
)

func TestContactMessage_Validate(t *testing.T) {
	tests := []struct {
		name    string
		contact ContactMessage
		wantErr bool
	}{
		{
			name: "contact valide",
			contact: ContactMessage{
				Name:    "Alice DOE",
				Email:   "alice@example.com",
				Subject: "Demande de contact",
				Message: "Bonjour, je souhaite vous contacter.",
			},
			wantErr: false,
		},
		{
			name: "nom vide",
			contact: ContactMessage{
				Name:    "",
				Email:   "alice@example.com",
				Subject: "Demande de contact",
				Message: "Bonjour",
			},
			wantErr: true,
		},
		{
			name: "nom trop long",
			contact: ContactMessage{
				Name:    string(make([]byte, 101)),
				Email:   "alice@example.com",
				Subject: "Demande de contact",
				Message: "Bonjour",
			},
			wantErr: true,
		},
		{
			name: "email vide",
			contact: ContactMessage{
				Name:    "Alice",
				Email:   "",
				Subject: "Demande de contact",
				Message: "Bonjour",
			},
			wantErr: true,
		},
		{
			name: "email sans arobase",
			contact: ContactMessage{
				Name:    "Alice",
				Email:   "Alice.example.com",
				Subject: "Demande de contact",
				Message: "Bonjour",
			},
			wantErr: true,
		},
		{
			name: "email trop long",
			contact: ContactMessage{
				Name:    "Alice",
				Email:   string(make([]byte, 151)),
				Subject: "Demande de contact",
				Message: "Bonjour",
			},
			wantErr: true,
		},
		{
			name: "sujet vide",
			contact: ContactMessage{
				Name:    "Alice",
				Email:   "alice@example.com",
				Subject: "",
				Message: "Bonjour",
			},
			wantErr: true,
		},
		{
			name: "sujet trop long",
			contact: ContactMessage{
				Name:    "Alice",
				Email:   "alice@example.com",
				Subject: string(make([]byte, 151)),
				Message: "Bonjour",
			},
			wantErr: true,
		},
		{
			name: "message vide",
			contact: ContactMessage{
				Name:    "Alice",
				Email:   "alice@example.com",
				Subject: "Demande de contact",
				Message: "",
			},
			wantErr: true,
		},
		{
			name: "message trop long",
			contact: ContactMessage{
				Name:    "Alice",
				Email:   "alice@example.com",
				Subject: "Demande de contact",
				Message: string(make([]byte, 5001)),
			},
			wantErr: true,
		},
		{
			name: "espaces autour des champs",
			contact: ContactMessage{
				Name:    "  Alice  ",
				Email:   "  alice@example.com  ",
				Subject: "  Demande de contact  ",
				Message: "  Bonjour  ",
			},
			wantErr: false,
		},
		{
			name: "nom composé uniquement d'espaces",
			contact: ContactMessage{
				Name:    "     ",
				Email:   "alice@example.com",
				Subject: "Demande",
				Message: "Bonjour",
			},
			wantErr: true,
		},
		{
			name: "email composé uniquement d'espaces",
			contact: ContactMessage{
				Name:    "Alice",
				Email:   "     ",
				Subject: "Demande",
				Message: "Bonjour",
			},
			wantErr: true,
		},
		{
			name: "sujet composé uniquement d'espaces",
			contact: ContactMessage{
				Name:    "Alice",
				Email:   "alice@example.com",
				Subject: "     ",
				Message: "Bonjour",
			},
			wantErr: true,
		},
		{
			name: "message composé uniquement d'espaces",
			contact: ContactMessage{
				Name:    "Alice",
				Email:   "alice@example.com",
				Subject: "Demande",
				Message: "     ",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.contact.Validate()

			if tt.wantErr && err == nil {
				t.Fatalf("Validate() devrait retourner une erreur")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("Validate() retourne une erreur inattendue : %v", err)
			}

			if tt.wantErr && !errors.Is(err, ErrValidation) {
				t.Fatalf("erreur attendue de type ErrValidation, obtenu : %v", err)
			}
		})
	}
}

func TestContactMessage_Validate_TrimsFields(t *testing.T) {
	contact := ContactMessage{
		Name:    "  Alice DOE  ",
		Email:   "  alice@example.com  ",
		Subject: "  Demande de contact  ",
		Message: "  Bonjour, ceci est un message.  ",
	}

	err := contact.Validate()

	if err != nil {
		t.Fatalf("Validate() retourne une erreur inattendue : %v", err)
	}

	if contact.Name != "Alice DOE" {
		t.Errorf("Name = %q, attendu %q", contact.Name, "Alice DOE")
	}

	if contact.Email != "alice@example.com" {
		t.Errorf("Email = %q, attendu %q", contact.Email, "alice@example.com")
	}

	if contact.Subject != "Demande de contact" {
		t.Errorf("Subject = %q, attendu %q", contact.Subject, "Demande de contact")
	}

	if contact.Message != "Bonjour, ceci est un message." {
		t.Errorf(
			"Message = %q, attendu %q",
			contact.Message,
			"Bonjour, ceci est un message.",
		)
	}
}

func TestContactMessage_Validate_Boundaries(t *testing.T) {
	tests := []struct {
		name    string
		contact ContactMessage
		wantErr bool
	}{
		{
			name: "nom exactement 100 caractères",
			contact: ContactMessage{
				Name:    string(make([]byte, 100)),
				Email:   "alice@example.com",
				Subject: "Demande",
				Message: "Bonjour",
			},
			wantErr: false,
		},
		{
			name: "sujet exactement 150 caractères",
			contact: ContactMessage{
				Name:    "Alice",
				Email:   "alice@example.com",
				Subject: string(make([]byte, 150)),
				Message: "Bonjour",
			},
			wantErr: false,
		},
		{
			name: "message exactement 5000 caractères",
			contact: ContactMessage{
				Name:    "Alice",
				Email:   "alice@example.com",
				Subject: "Demande",
				Message: string(make([]byte, 5000)),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.contact.Validate()

			if tt.wantErr && err == nil {
				t.Fatal("une erreur était attendue")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("erreur inattendue : %v", err)
			}
		})
	}
}

func TestContactMessage_CanStoreMetadata(t *testing.T) {
	now := time.Now()

	contact := ContactMessage{
		ID:        "550e8400-e29b-41d4-a716-446655440000",
		Name:      "Alice",
		Email:     "alice@example.com",
		Subject:   "Test",
		Message:   "Bonjour",
		CreatedAt: now,
	}

	if contact.ID == "" {
		t.Error("ID ne devrait pas être vide")
	}

	if contact.CreatedAt.IsZero() {
		t.Error("CreatedAt ne devrait pas être vide")
	}
}
