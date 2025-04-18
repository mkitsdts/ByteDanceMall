package utils

import (
	pb "bytedancemall/llm/proto"
	"fmt"
	"strings"
)

func ProductInfoToString(ProductInfo []*pb.Product) string {
	res := ""
	idx := 1
	for range ProductInfo {
		res += "商品" + string(idx) + ":\n"
		res += "商品名称: " + ProductInfo[idx].Name + "\n"
		res += "商品描述: " + ProductInfo[idx].Describe + "\n"
		res += "商品数量: " + string(ProductInfo[idx].Quantity) + "\n"
		res += "商品id: " + string(ProductInfo[idx].Id) + "\n"
		res += "\n"
		idx++
	}
	return res
}

func StringParseToUID(str string) []uint32 {
	res := make([]uint32, 0)
	// string一行一个uint32整数
	lines := strings.Split(str, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		var uid uint32
		_, err := fmt.Sscanf(line, "%d", &uid)
		if err != nil {
			continue
		}
		res = append(res, uid)
	}
	return res
}
