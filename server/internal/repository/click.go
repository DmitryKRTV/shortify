package repository

import (
	"context"
	"shortify/server/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ClickRepository struct {
	db *gorm.DB
}

func NewClickRepository(db *gorm.DB) *ClickRepository {
	return &ClickRepository{db: db}
}

func (r *ClickRepository) Create(ctx context.Context, click *domain.Click) error {
	return r.db.WithContext(ctx).Create(click).Error
}

func (r *ClickRepository) CountByLink(ctx context.Context, linkID uuid.UUID) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&domain.Click{}).
		Where("link_id = ?", linkID).
		Count(&total).Error
	return total, err
}

func (r *ClickRepository) ListByLink(ctx context.Context, linkID uuid.UUID, limit int) ([]domain.Click, error) {
	var clicks []domain.Click
	err := r.db.WithContext(ctx).
		Where("link_id = ?", linkID).
		Order("created_at DESC").
		Limit(limit).
		Find(&clicks).Error
	return clicks, err
}
