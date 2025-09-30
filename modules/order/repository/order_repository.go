package repository

import (
	"context"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type (
	OrderRepository interface {
		Create(ctx context.Context, tx *gorm.DB, o entities.Order) (entities.Order, error)
		FindByID(ctx context.Context, tx *gorm.DB, id string) (entities.Order, error)
		List(ctx context.Context, tx *gorm.DB, limit, offset int) ([]entities.Order, int64, error)
		Delete(ctx context.Context, tx *gorm.DB, id string) error
	}

	orderRepository struct {
		db *gorm.DB
	}
)

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *orderRepository) Create(ctx context.Context, tx *gorm.DB, o entities.Order) (entities.Order, error) {
	db := r.getDB(tx)
	if err := db.WithContext(ctx).Create(&o).Error; err != nil {
		return entities.Order{}, err
	}
	return o, nil
}

func (r *orderRepository) FindByID(ctx context.Context, tx *gorm.DB, id string) (entities.Order, error) {
	db := r.getDB(tx)
	var o entities.Order
	if err := db.WithContext(ctx).Preload("Product").Where("id = ?", id).Take(&o).Error; err != nil {
		return entities.Order{}, err
	}
	return o, nil
}

func (r *orderRepository) List(ctx context.Context, tx *gorm.DB, limit, offset int) ([]entities.Order, int64, error) {
	db := r.getDB(tx)
	var (
		items []entities.Order
		total int64
	)
	if err := db.WithContext(ctx).Model(&entities.Order{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.WithContext(ctx).Model(&entities.Order{}).
		Preload("Product").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *orderRepository) Delete(ctx context.Context, tx *gorm.DB, id string) error {
	db := r.getDB(tx)
	return db.WithContext(ctx).Clauses(clause.Returning{}).Delete(&entities.Order{}, "id = ?", id).Error
}
