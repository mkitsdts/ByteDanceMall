package repository

import (
	"bytedancemall/inventory/model"
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repositories struct {
	db        *gorm.DB
	Inventory *InventoryRepository
	Deduction *DeductionRepository
}

func New(db *gorm.DB) *Repositories {
	repos := &Repositories{db: db}
	repos.Inventory = &InventoryRepository{db: db}
	repos.Deduction = &DeductionRepository{db: db}
	return repos
}

func (r *Repositories) Transaction(ctx context.Context, fn func(txRepos *Repositories) error) error {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if rec := recover(); rec != nil {
			tx.Rollback()
			panic(rec)
		}
	}()

	txRepos := New(tx)
	if err := fn(txRepos); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

type InventoryRepository struct {
	db *gorm.DB
}

func (r *InventoryRepository) FindByProductID(ctx context.Context, productID uint64) (*model.Inventory, error) {
	var inventory model.Inventory
	if err := r.db.WithContext(ctx).Where("product_id = ?", productID).First(&inventory).Error; err != nil {
		return nil, err
	}
	return &inventory, nil
}

func (r *InventoryRepository) FindByProductIDForUpdate(ctx context.Context, productID uint64) (*model.Inventory, error) {
	var inventory model.Inventory
	if err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("product_id = ?", productID).
		First(&inventory).Error; err != nil {
		return nil, err
	}
	return &inventory, nil
}

func (r *InventoryRepository) Create(ctx context.Context, inventory *model.Inventory) error {
	return r.db.WithContext(ctx).Create(inventory).Error
}

func (r *InventoryRepository) DeleteByProductID(ctx context.Context, productID uint64) error {
	return r.db.WithContext(ctx).Where("product_id = ?", productID).Delete(&model.Inventory{}).Error
}

func (r *InventoryRepository) IncreaseLockedStock(ctx context.Context, productID, amount uint64) error {
	return r.db.WithContext(ctx).
		Model(&model.Inventory{}).
		Where("product_id = ?", productID).
		Updates(map[string]any{
			"locked_stock": gorm.Expr("locked_stock + ?", amount),
			"version":      gorm.Expr("version + 1"),
		}).Error
}

func (r *InventoryRepository) DecreaseLockedStock(ctx context.Context, productID, amount uint64) error {
	return r.db.WithContext(ctx).
		Model(&model.Inventory{}).
		Where("product_id = ? AND locked_stock >= ?", productID, amount).
		Updates(map[string]any{
			"locked_stock": gorm.Expr("locked_stock - ?", amount),
			"version":      gorm.Expr("version + 1"),
		}).Error
}

func (r *InventoryRepository) CommitDeduction(ctx context.Context, productID, amount uint64) error {
	return r.db.WithContext(ctx).
		Model(&model.Inventory{}).
		Where("product_id = ?", productID).
		Updates(map[string]any{
			"total_stock":  gorm.Expr("total_stock - ?", amount),
			"locked_stock": gorm.Expr("locked_stock - ?", amount),
			"version":      gorm.Expr("version + 1"),
		}).Error
}

type DeductionRepository struct {
	db *gorm.DB
}

func (r *DeductionRepository) FindByRecordID(ctx context.Context, recordID string) (*model.OutInventory, error) {
	var record model.OutInventory
	if err := r.db.WithContext(ctx).Where("record_id = ?", recordID).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *DeductionRepository) Create(ctx context.Context, record *model.OutInventory) error {
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *DeductionRepository) MarkCommitted(ctx context.Context, recordID string, orderID uint64) (int64, error) {
	result := r.db.WithContext(ctx).
		Model(&model.OutInventory{}).
		Where("record_id = ? AND state = ?", recordID, model.DeductStatePending).
		Updates(map[string]any{
			"order_id": &orderID,
			"state":    model.DeductStateCommitted,
		})
	return result.RowsAffected, result.Error
}

func (r *DeductionRepository) MarkCanceled(ctx context.Context, recordID string) (int64, error) {
	result := r.db.WithContext(ctx).
		Model(&model.OutInventory{}).
		Where("record_id = ? AND state = ?", recordID, model.DeductStatePending).
		Update("state", model.DeductStateCanceled)
	return result.RowsAffected, result.Error
}

func (r *DeductionRepository) UpdateCommittedAt(ctx context.Context, recordID string) error {
	return r.db.WithContext(ctx).
		Model(&model.OutInventory{}).
		Where("record_id = ?", recordID).
		Update("committed_at", gorm.Expr("CURRENT_TIMESTAMP")).Error
}
