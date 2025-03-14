package utils

import (
	"sync"
)

var id int64 = 0

func GenerateId() int64 {
	// 递增生成id
	var mu sync.Mutex
	mu.Lock()
	id++
	mu.Unlock()
	return id
}