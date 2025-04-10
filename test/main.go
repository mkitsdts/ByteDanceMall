package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

func main() {
	productId := 1
	// quantity := 10
	// releaseTime := "2025-4-10 9:35:00"
	// body := fmt.Sprintf(`{"product_id": %d, "quantity": %d, "release_time": %s}`, productId, quantity, releaseTime)
	// req, err := http.NewRequest("POST", "http://192.168.2.169:8080/seckill", strings.NewReader(body))
	// if err != nil {
	// 	panic(err)
	// }
	// req.Header.Set("Content-Type", "application/json")
	// client := &http.Client{}
	// resp, err := client.Do(req)
	// if err != nil {
	// 	panic(err)
	// }
	// defer resp.Body.Close()
	// fmt.Printf("Error: %s\n", resp.Status)
	// time.Sleep(2 * time.Second)
	// 发送请求
	for range 100 {
		go func() {
			for j := 1; j <= 10; j++ {
				// 发送 POST 请求
				// 这里可以使用 http.NewRequest 来设置请求头等
				// 也可以使用 http.Post 来简化请求
				reqBody := fmt.Sprintf(`{"product_id": %d, "user_id": %d}`, productId, j)
				reqSec, err := http.NewRequest("POST", "http://192.168.2.169:8080/seckill/tryseckill", strings.NewReader(reqBody))
				if err != nil {
					panic(err)
				}
				reqSec.Header.Set("Content-Type", "application/json")
				client := &http.Client{}
				resp, err := client.Do(reqSec)
				if err != nil {
					panic(err)
				}
				defer resp.Body.Close()
				fmt.Printf("Result: %s\n", resp.Status)
				time.Sleep(100 * time.Millisecond)
			}
		}()
	}
	// 等待所有请求完成
	time.Sleep(100 * time.Second)
}