package cache

import "errors"

var (
	ErrInventoryCacheBusy     = errors.New("inventory bucket is committing")
	ErrInsufficientCacheStock = errors.New("insufficient redis inventory")
)
