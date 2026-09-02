package model

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrValidation = errors.New("validation")

type ContactMessage struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Subject   string    `json:"subject"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

func (c *ContactMessage) Validate() error {
	c.Name = strings.TrimSpace(c.Name)
	c.Email = strings.TrimSpace(c.Email)
	c.Subject = strings.TrimSpace(c.Subject)
	c.Message = strings.TrimSpace(c.Message)

	if c.Name == "" || len(c.Name) > 100 {
		return fmt.Errorf("%w: le nom est requis (100 caractères max)", ErrValidation)
	}
	if c.Email == "" || !strings.Contains(c.Email, "@") || len(c.Email) > 150 {
		return fmt.Errorf("%w: l'email est invalide", ErrValidation)
	}
	if c.Subject == "" || len(c.Subject) > 150 {
		return fmt.Errorf("%w: le sujet est requis (150 caractères max)", ErrValidation)
	}
	if c.Message == "" || len(c.Message) > 5000 {
		return fmt.Errorf("%w: le message est requis (5000 caractères max)", ErrValidation)
	}
	return nil
}
