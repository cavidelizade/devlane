package store

import (
	"context"
	"errors"

	"github.com/Devlaner/devlane/api/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EmailChangeRequestStore persists pending email-change verifications.
type EmailChangeRequestStore struct{ db *gorm.DB }

func NewEmailChangeRequestStore(db *gorm.DB) *EmailChangeRequestStore {
	return &EmailChangeRequestStore{db: db}
}

// GetByUserID returns the user's pending email change, or nil when none exists.
func (s *EmailChangeRequestStore) GetByUserID(ctx context.Context, userID uuid.UUID) (*model.EmailChangeRequest, error) {
	var r model.EmailChangeRequest
	err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&r).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}

// Upsert stores the user's pending email change, replacing any earlier one so
// only the most recent request is ever valid.
func (s *EmailChangeRequestStore) Upsert(ctx context.Context, r *model.EmailChangeRequest) error {
	existing, err := s.GetByUserID(ctx, r.UserID)
	if err != nil {
		return err
	}
	if existing != nil {
		existing.NewEmail = r.NewEmail
		existing.CodeHash = r.CodeHash
		existing.ExpiresAt = r.ExpiresAt
		return s.db.WithContext(ctx).Save(existing).Error
	}
	return s.db.WithContext(ctx).Create(r).Error
}

// DeleteByUserID removes the user's pending email change.
func (s *EmailChangeRequestStore) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&model.EmailChangeRequest{}).Error
}
