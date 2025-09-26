package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func FloatsToBytes(floats *[]float32) []byte {
	if floats == nil {
		fmt.Println("input floats is nil")
		return nil
	}

	// 创建一个 bytes.Buffer 来写入二进制数据
	buf := new(bytes.Buffer)

	// 使用 binary.Write 进行转换
	// binary.LittleEndian 是关键，它确保字节顺序是正确的
	err := binary.Write(buf, binary.LittleEndian, *floats)
	if err != nil {
		// 在实际应用中，这里应该记录错误日志
		fmt.Println("binary.Write failed:", err)
		return nil
	}

	return buf.Bytes()
}
