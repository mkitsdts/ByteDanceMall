package utils

import (
	"fmt"
	"sync"
	"time"
)

const (
	nodeBits  uint8  = 10                    // 机器ID的位数
	stepBits  uint8  = 12                    // 序列号的位数
	nodeMax   uint64 = -1 ^ (-1 << nodeBits) // 机器ID的最大值
	stepMax   uint64 = -1 ^ (-1 << stepBits) // 序列号的最大值
	timeShift uint8  = nodeBits + stepBits   // 时间戳左移位数
	nodeShift uint8  = stepBits              // 机器ID左移位数
)

var epoch uint64 = 1577836800000 // 2020-01-01 00:00:00 作为起始时间

// Snowflake 定义雪花算法结构
type Snowflake struct {
	mu        sync.Mutex
	timestamp uint64
	node      uint64
	step      uint64
}

// NewSnowflake 创建一个新的雪花算法实例
func NewSnowflake(node uint64) (*Snowflake, error) {
	if node > nodeMax {
		return nil, fmt.Errorf("node ID must be between 0 and %d", nodeMax)
	}
	return &Snowflake{
		timestamp: 0,
		node:      node,
		step:      0,
	}, nil
}

// GenerateID 生成唯一ID
func (s *Snowflake) GenerateID() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := uint64(time.Now().UnixNano() / 1000000) // 当前时间戳（毫秒）

	if s.timestamp == now {
		// 如果是同一时间生成的，则进行毫秒内序列
		s.step = (s.step + 1) & stepMax
		if s.step == 0 {
			// 序列号已经达到最大值，等待下一毫秒
			for now <= s.timestamp {
				now = uint64(time.Now().UnixNano() / 1000000)
			}
		}
	} else {
		// 时间戳改变，毫秒内序列重置
		s.step = 0
	}

	s.timestamp = now

	// 生成ID
	return ((now - epoch) << timeShift) | (s.node << nodeShift) | s.step
}

// 单例模式使用
var defaultSnowflake *Snowflake
var once sync.Once

var MachindID uint64

// GenerateId 保持原有API兼容性
func GenerateId() uint64 {
	once.Do(func() {
		flake, _ := NewSnowflake(MachindID) // 假设本机器ID为1
		defaultSnowflake = flake
	})
	return defaultSnowflake.GenerateID()
}
