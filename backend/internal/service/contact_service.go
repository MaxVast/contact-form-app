package service

import (
	"context"

	"github.com/maxvast/contact-form-app/backend/internal/model"
)

type Repository interface {
	Save(ctx context.Context, c *model.ContactMessage) error
	List(ctx context.Context, limit int) ([]model.ContactMessage, error)
	Ping(ctx context.Context) error
}

type ContactService struct {
	repo Repository
}

func NewContactService(repo Repository) *ContactService {
	return &ContactService{repo: repo}
}

func (s *ContactService) Submit(ctx context.Context, c *model.ContactMessage) error {
	if err := c.Validate(); err != nil {
		return err
	}
	return s.repo.Save(ctx, c)
}

func (s *ContactService) Recent(ctx context.Context) ([]model.ContactMessage, error) {
	return s.repo.List(ctx, 50)
}

func (s *ContactService) Ping(ctx context.Context) error {
	return s.repo.Ping(ctx)
}
