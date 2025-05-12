package middleware

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/binary"
	"sync"
)

// 使用3个哈希函数
type BloomFilter struct {
	read  sync.RWMutex
	write sync.Mutex
	data  []byte
}

func NewBloomFilter(size int) *BloomFilter {
	sz := (size + 7) / 8
	d := make([]byte, sz)
	return &BloomFilter{
		data: d,
	}
}

func (f *BloomFilter) hash1(key string) int {
	// md5算法
	value := md5.Sum([]byte(key))
	hash := binary.BigEndian.Uint32(value[:])
	return int(hash % uint32(len(f.data)))
}

func (f *BloomFilter) hash2(key string) int {
	// sha1算法
	value := sha1.Sum([]byte(key))
	hash := binary.BigEndian.Uint32(value[:])
	return int(hash % uint32(len(f.data)))
}

func (f *BloomFilter) Set(key string) {
	f.write.Lock()
	defer f.write.Unlock()

	index1 := f.hash1(key)
	index2 := f.hash2(key)

	f.data[index1/8] |= 1 << (index1 % 8)
	f.data[index2/8] |= 1 << (index2 % 8)
}

func (f *BloomFilter) Check(key string) bool {
	f.read.RLock()
	defer f.read.RUnlock()

	index1 := f.hash1(key)
	index2 := f.hash2(key)

	return (f.data[index1/8]&(1<<(index1%8)) != 0) && (f.data[index2/8]&(1<<(index2%8)) != 0)
}

func (f *BloomFilter) Clear() {
	f.write.Lock()
	defer f.write.Unlock()

	for i := range f.data {
		f.data[i] = 0
	}
}
