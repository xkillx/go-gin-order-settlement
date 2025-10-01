package repository

import (
	"context"

	"github.com/xkillx/go-gin-order-settlement/database/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type (
	ProductRepository interface {
		Create(ctx context.Context, tx *gorm.DB, p entities.Product) (entities.Product, error)
		FindByID(ctx context.Context, tx *gorm.DB, id string) (entities.Product, error)
		List(ctx context.Context, tx *gorm.DB, limit, offset int) ([]entities.Product, int64, error)
		Update(ctx context.Context, tx *gorm.DB, p entities.Product) (entities.Product, error)
		Delete(ctx context.Context, tx *gorm.DB, id string) error
		DecrementStock(ctx context.Context, tx *gorm.DB, productID uuid.UUID, qty int) (bool, error)
	}

	productRepository struct {
		db *gorm.DB
	}
)

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *productRepository) Create(ctx context.Context, tx *gorm.DB, p entities.Product) (entities.Product, error) {
	db := r.getDB(tx)
	if err := db.WithContext(ctx).Create(&p).Error; err != nil {
		return entities.Product{}, err
	}
	return p, nil
}

func (r *productRepository) FindByID(ctx context.Context, tx *gorm.DB, id string) (entities.Product, error) {
	db := r.getDB(tx)
	var p entities.Product
	if err := db.WithContext(ctx).Where("id = ?", id).Take(&p).Error; err != nil {
		return entities.Product{}, err
	}
	return p, nil
}

func (r *productRepository) List(ctx context.Context, tx *gorm.DB, limit, offset int) ([]entities.Product, int64, error) {
	db := r.getDB(tx)
	var (
		items []entities.Product
		total int64
	)
	if err := db.WithContext(ctx).Model(&entities.Product{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.WithContext(ctx).Model(&entities.Product{}).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *productRepository) Update(ctx context.Context, tx *gorm.DB, p entities.Product) (entities.Product, error) {
	db := r.getDB(tx)
	if err := db.WithContext(ctx).Clauses(clause.Returning{}).Updates(&p).Error; err != nil {
		return entities.Product{}, err
	}
	return p, nil
}

func (r *productRepository) Delete(ctx context.Context, tx *gorm.DB, id string) error {
	db := r.getDB(tx)
	return db.WithContext(ctx).Delete(&entities.Product{}, "id = ?", id).Error
}

func (r *productRepository) DecrementStock(ctx context.Context, tx *gorm.DB, productID uuid.UUID, qty int) (bool, error) {
	db := r.getDB(tx)
	res := db.WithContext(ctx).Model(&entities.Product{}).
		Where("id = ? AND stock >= ?", productID, qty).
		Update("stock", gorm.Expr("stock - ?", qty))
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}
