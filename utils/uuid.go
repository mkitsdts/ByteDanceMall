package utils

import (
	"sync"
	"time"
)

// Snowflake constants
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

	// Get current timestamp in milliseconds
	timestamp := time.Now().UnixNano() / 1e6

	// If clock moved backwards, wait until it catches up
	if timestamp < lastTimestamp {
		time.Sleep(time.Duration(lastTimestamp-timestamp) * time.Millisecond)
		timestamp = time.Now().UnixNano() / 1e6
	}

	// If same timestamp, increment sequence
	if timestamp == lastTimestamp {
		sequence = (sequence + 1) & sequenceMask
		// If sequence exhausted, wait for next millisecond
		if sequence == 0 {
			for timestamp <= lastTimestamp {
				timestamp = time.Now().UnixNano() / 1e6
			}
		}
	} else {
		// Reset sequence for new timestamp
		sequence = 0
	}

	lastTimestamp = timestamp

	// Extract a portion of the userID to fit within the allocated bits
	userIDPortion := int64(userID) & userIDMask

	// Build the ID: timestamp | userID | sequence
	return uint64((timestamp-epoch)<<timestampShift |
		userIDPortion<<userIDShift |
		sequence)
}
