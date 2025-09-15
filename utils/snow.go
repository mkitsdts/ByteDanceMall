package utils

import (
	"sync"
	"time"
)

const (
	epoch        int64 = 1609459200000
	userIDBits   uint8 = 10
	sequenceBits uint8 = 12

	userIDShift    = sequenceBits
	timestampShift = userIDBits + sequenceBits

	userIDMask   = (1 << userIDBits) - 1
	sequenceMask = (1 << sequenceBits) - 1
)

var (
	lastTimestamp int64 = -1
	sequence      int64 = 0
	mutex         sync.Mutex
)

func GenerateOrderID(userID uint64) uint64 {
	mutex.Lock()
	defer mutex.Unlock()

	// 获取时间戳
	timestamp := time.Now().UnixNano() / 1e6

	// 如果时间回拨则等待
	if timestamp < lastTimestamp {
		time.Sleep(time.Duration(lastTimestamp-timestamp) * time.Millisecond)
		timestamp = time.Now().UnixNano() / 1e6
	}

	// 如果是同一时间戳，递增序列号
	if timestamp == lastTimestamp {
		sequence = (sequence + 1) & sequenceMask
		// 如果序列号耗尽，等待下一毫秒
		if sequence == 0 {
			for timestamp <= lastTimestamp {
				timestamp = time.Now().UnixNano() / 1e6
			}
		}
	} else {
		// 新时间戳，重置序列号
		sequence = 0
	}

	lastTimestamp = timestamp

	// 限制 userID 部分
	userIDPortion := int64(userID) & userIDMask

	return uint64((timestamp-epoch)<<timestampShift |
		userIDPortion<<userIDShift |
		sequence)
}
