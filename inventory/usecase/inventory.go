package usecase

import (
	"bytedancemall/inventory/cache"
	"bytedancemall/inventory/model"
	"bytedancemall/inventory/repository"
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

var (
	ErrInvalidRequest         = errors.New("invalid request")
	ErrInsufficientDBStock    = errors.New("insufficient database inventory")
	ErrDeductionAlreadyExists = errors.New("deduction record already exists")
)

type InventoryUsecase struct {
	repos       *repository.Repositories
	cache       *cache.InventoryCache
	queryFlight flightGroup
}

func New(repos *repository.Repositories, cacheStore *cache.InventoryCache) *InventoryUsecase {
	return &InventoryUsecase{
		repos: repos,
		cache: cacheStore,
	}
}

func (u *InventoryUsecase) Deduct(ctx context.Context, productID uint64, recordID string, amount uint64) (*model.OutInventory, error) {
	if productID == 0 || recordID == "" || amount == 0 {
		return nil, ErrInvalidRequest
	}

	record, err := u.repos.Deduction.FindByRecordID(ctx, recordID)
	if err == nil {
		return record, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	bucket, _, err := u.cache.Reserve(ctx, productID, recordID, amount)
	if err != nil {
		return nil, err
	}

	var created *model.OutInventory
	err = u.repos.Transaction(ctx, func(txRepos *repository.Repositories) error {
		inventory, err := txRepos.Inventory.FindByProductIDForUpdate(ctx, productID)
		if err != nil {
			return err
		}
		if inventory.TotalStock < inventory.LockedStock || inventory.TotalStock-inventory.LockedStock < amount {
			return ErrInsufficientDBStock
		}

		record := &model.OutInventory{
			RecordID:  recordID,
			ProductID: productID,
			Amount:    amount,
			Bucket:    bucket,
			State:     model.DeductStatePending,
		}
		if err := txRepos.Deduction.Create(ctx, record); err != nil {
			if isDuplicateKeyError(err) {
				return ErrDeductionAlreadyExists
			}
			return err
		}
		if err := txRepos.Inventory.IncreaseLockedStock(ctx, productID, amount); err != nil {
			return err
		}
		created = record
		return nil
	})
	if err != nil {
		_ = u.cache.RestoreReservation(ctx, productID, bucket, amount)
		if errors.Is(err, ErrDeductionAlreadyExists) {
			return u.repos.Deduction.FindByRecordID(ctx, recordID)
		}
		return nil, err
	}

	_ = u.cache.DeleteAvailableStock(ctx, productID)

	return created, nil
}

func (u *InventoryUsecase) CommitByRecordID(ctx context.Context, recordID string, orderID uint64) error {
	record, err := u.repos.Deduction.FindByRecordID(ctx, recordID)
	if err != nil {
		return err
	}
	if record.State == model.DeductStateCommitted {
		return nil
	}
	if record.State != model.DeductStatePending {
		return gorm.ErrInvalidData
	}

	if err := u.cache.MarkBucketCommitting(ctx, record.ProductID, record.Bucket); err != nil {
		return err
	}
	defer func() {
		_ = u.cache.DeleteBucket(ctx, record.ProductID, record.Bucket)
	}()

	err = u.repos.Transaction(ctx, func(txRepos *repository.Repositories) error {
		inventory, err := txRepos.Inventory.FindByProductIDForUpdate(ctx, record.ProductID)
		if err != nil {
			return err
		}
		if inventory.TotalStock < record.Amount || inventory.LockedStock < record.Amount {
			return gorm.ErrInvalidData
		}
		if err := txRepos.Inventory.CommitDeduction(ctx, record.ProductID, record.Amount); err != nil {
			return err
		}
		rows, err := txRepos.Deduction.MarkCommitted(ctx, record.RecordID, orderID)
		if err != nil || rows == 0 {
			return err
		}
		if err := txRepos.Deduction.UpdateCommittedAt(ctx, record.RecordID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return u.cache.DeleteAvailableStock(ctx, record.ProductID)
}

func (u *InventoryUsecase) ReleaseByRecordID(ctx context.Context, recordID string) error {
	record, err := u.repos.Deduction.FindByRecordID(ctx, recordID)
	if err != nil {
		return err
	}
	if record.State == model.DeductStateCanceled {
		return nil
	}
	if record.State != model.DeductStatePending {
		return gorm.ErrInvalidData
	}

	err = u.repos.Transaction(ctx, func(txRepos *repository.Repositories) error {
		rows, err := txRepos.Deduction.MarkCanceled(ctx, record.RecordID)
		if err != nil || rows == 0 {
			return err
		}
		return txRepos.Inventory.DecreaseLockedStock(ctx, record.ProductID, record.Amount)
	})
	if err != nil {
		return err
	}

	_ = u.cache.DeleteAvailableStock(ctx, record.ProductID)
	return u.cache.DeleteBucket(ctx, record.ProductID, record.Bucket)
}

func (u *InventoryUsecase) Create(ctx context.Context, productID, amount uint64) error {
	if productID == 0 || amount == 0 {
		return ErrInvalidRequest
	}
	inventory := &model.Inventory{
		ProductID:   productID,
		TotalStock:  amount,
		LockedStock: 0,
		State:       model.StateOnSale,
	}
	if err := u.repos.Inventory.Create(ctx, inventory); err != nil {
		return err
	}
	_ = u.cache.DeleteAvailableStock(ctx, productID)
	return u.cache.SeedProductBuckets(ctx, productID, amount)
}

func (u *InventoryUsecase) Delete(ctx context.Context, productID uint64) error {
	if err := u.cache.DeleteProductBuckets(ctx, productID); err != nil {
		return err
	}
	if err := u.repos.Inventory.DeleteByProductID(ctx, productID); err != nil {
		return err
	}
	_ = u.cache.DeleteAvailableStock(ctx, productID)
	return nil
}

func (u *InventoryUsecase) Query(ctx context.Context, productID uint64) (*model.Inventory, error) {
	if productID == 0 {
		return nil, ErrInvalidRequest
	}

	stock, hit, err := u.cache.GetAvailableStock(ctx, productID)
	if err == nil && hit {
		return &model.Inventory{
			ProductID:   productID,
			TotalStock:  stock,
			LockedStock: 0,
		}, nil
	}
	if err != nil {
		return nil, err
	}

	result, err, _ := u.queryFlight.Do(fmt.Sprintf("query:%d", productID), func() (any, error) {
		cachedStock, cachedHit, cacheErr := u.cache.GetAvailableStock(ctx, productID)
		if cacheErr == nil && cachedHit {
			return &model.Inventory{
				ProductID:   productID,
				TotalStock:  cachedStock,
				LockedStock: 0,
			}, nil
		}
		if cacheErr != nil {
			return nil, cacheErr
		}

		inventory, dbErr := u.repos.Inventory.FindByProductID(ctx, productID)
		if dbErr != nil {
			return nil, dbErr
		}

		available := inventory.TotalStock
		if inventory.TotalStock >= inventory.LockedStock {
			available = inventory.TotalStock - inventory.LockedStock
		} else {
			available = 0
		}
		if err := u.cache.SetAvailableStock(ctx, productID, available); err != nil {
			return nil, err
		}

		return inventory, nil
	})
	if err != nil {
		return nil, err
	}
	return result.(*model.Inventory), nil
}

func (u *InventoryUsecase) Preheat(ctx context.Context, productIDs []uint64) error {
	for _, productID := range productIDs {
		inventory, err := u.repos.Inventory.FindByProductID(ctx, productID)
		if err != nil {
			return err
		}
		if err := u.cache.SeedProductBuckets(ctx, inventory.ProductID, inventory.TotalStock); err != nil {
			return err
		}
	}
	return nil
}

func isDuplicateKeyError(err error) bool {
	return err != nil && (errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "duplicated key"))
}
