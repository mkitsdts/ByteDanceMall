package service

import (
	"bytes"
	"runtime"
	"strconv"
)

func getGid() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	i := bytes.IndexByte(b, ' ')
	if i < 0 {
		return 0
	}
	id, _ := strconv.ParseUint(string(b[:i]), 10, 64)
	return id
}
