package store

import (
	"context"

	"github.com/Devlaner/devlane/api/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// InstanceAdminStore handles instance_admins persistence — the set of users
// authorized to manage instance settings (mirrors Plane's InstanceAdmin).
type InstanceAdminStore struct{ db *gorm.DB }

func NewInstanceAdminStore(db *gorm.DB) *InstanceAdminStore {
	return &InstanceAdminStore{db: db}
}

// IsAdmin reports whether the user is an instance admin (role >= RoleAdmin).
func (s *InstanceAdminStore) IsAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&model.InstanceAdmin{}).
		Where("user_id = ? AND role >= ?", userID, model.RoleAdmin).
		Count(&count).Error
	return count > 0, err
}

// Create inserts an instance admin row.
func (s *InstanceAdminStore) Create(ctx context.Context, a *model.InstanceAdmin) error {
	return s.db.WithContext(ctx).Create(a).Error
}

// GetByUserID returns the (non-deleted) admin row for a user, if any.
func (s *InstanceAdminStore) GetByUserID(ctx context.Context, userID uuid.UUID) (*model.InstanceAdmin, error) {
	var a model.InstanceAdmin
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

// List returns all instance admins with the user's email/display name joined for display.
func (s *InstanceAdminStore) List(ctx context.Context) ([]model.InstanceAdmin, error) {
	var admins []model.InstanceAdmin
	err := s.db.WithContext(ctx).
		Table("instance_admins").
		Select("instance_admins.*, users.email AS user_email, users.display_name AS user_display_name").
		Joins("LEFT JOIN users ON users.id = instance_admins.user_id").
		Where("instance_admins.deleted_at IS NULL").
		Order("instance_admins.created_at ASC").
		Scan(&admins).Error
	return admins, err
}

// Count returns the number of active instance admins.
func (s *InstanceAdminStore) Count(ctx context.Context) (int64, error) {
	var n int64
	err := s.db.WithContext(ctx).Model(&model.InstanceAdmin{}).Count(&n).Error
	return n, err
}

// DeleteByPK soft-deletes an instance admin by its primary key.
func (s *InstanceAdminStore) DeleteByPK(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Delete(&model.InstanceAdmin{}, "id = ?", id).Error
}
