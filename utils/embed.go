package utils

import (
	"encoding/binary"
	"math"
)

func FloatsToBytes(fs *[]float32) []byte {
	buf := make([]byte, len(*fs)*4)

	for i, f := range *fs {
		u := math.Float32bits(f)
		binary.NativeEndian.PutUint32(buf[i*4:], u)
	}

	return buf
}
