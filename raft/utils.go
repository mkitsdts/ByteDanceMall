package raft

import (
	"math/rand"
	"time"
)

// 生成随机数时间间隔
func RandomDuration() time.Duration {
	return time.Duration(MIN_DURATION + rand.Intn(MAX_DURATION-MIN_DURATION))
}
