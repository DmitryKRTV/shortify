package repository

import (
	"context"
	"errors"
	"shortify/server/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LinkRepository struct {
	db *gorm.DB
}

func NewLinkRepository(db *gorm.DB) *LinkRepository {
	return &LinkRepository{db: db}
}

func (r *LinkRepository) Create(ctx context.Context, link *domain.Link) error {
	return r.db.WithContext(ctx).Create(link).Error
}

func (r *LinkRepository) FindByShortCode(ctx context.Context, code string) (*domain.Link, error) {
	var link domain.Link
	err := r.db.WithContext(ctx).Where("short_code = ?", code).First(&link).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *LinkRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Link, error) {
	var link domain.Link
	err := r.db.WithContext(ctx).First(&link, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *LinkRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Link, error) {
	var links []domain.Link
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&links).Error
	return links, err
}

func (r *LinkRepository) Delete(ctx context.Context, userID uuid.UUID, linkID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", linkID, userID).
		Delete(&domain.Link{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
