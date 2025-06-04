package main

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	// 配置参数
	productId := 1

	// 添加商品到秒杀系统
	addSeckillProduct(productId, 1000, "2023-5-17 10:00:00")
	time.Sleep(3 * time.Second) // 等待商品添加完成

	// 启动压测
	stressTest()
}

func addSeckillProduct(productId int, quantity int, releaseTime string) bool {
	body := fmt.Sprintf(`{"product_id": %d, "quantity": %d, "release_time": "%s"}`, productId, quantity, releaseTime)

	req, err := http.NewRequest("POST", "http://localhost:8080/seckill/add", bytes.NewReader([]byte(body)))
	if err != nil {
		fmt.Printf("创建请求失败: %v\n", err)
		return false
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("发送添加商品请求失败: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func sendSeckillRequest(productId, userId int) bool {
	reqBody := fmt.Sprintf(`{"product_id": %d, "user_id": %d}`, productId, userId)

	req, err := http.NewRequest("POST", "http://localhost:8080/seckill/tryseckill", bytes.NewReader([]byte(reqBody)))
	if err != nil {
		fmt.Printf("创建请求失败: %v\n", err)
		return false
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("发送秒杀请求失败: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func stressTest() {
	// 配置参数
	productId := 1
	concurrency := 100          // 并发数
	requestsPerGoroutine := 100 // 每个协程发送的请求数

	// 统计信息
	var successCount int64
	var failCount int64
	var totalLatency int64

	// 等待组
	var wg sync.WaitGroup
	wg.Add(concurrency)

	// 限制启动速度，避免瞬时峰值
	startTime := time.Now()

	// 启动多个goroutine并发请求
	for i := range concurrency {
		go func(workerID int) {
			defer wg.Done()

			for j := range requestsPerGoroutine {
				userId := workerID*requestsPerGoroutine + j

				// 记录请求开始时间
				requestStart := time.Now()

				// 发送秒杀请求
				success := sendSeckillRequest(productId, userId)

				// 计算延迟
				latency := time.Since(requestStart).Milliseconds()
				atomic.AddInt64(&totalLatency, latency)

				// 更新统计
				if success {
					atomic.AddInt64(&successCount, 1)
				} else {
					atomic.AddInt64(&failCount, 1)
				}

				// 打印进度
				if userId%100 == 0 {
					fmt.Printf("已完成 %d 个请求\n", userId)
				}
			}
		}(i)

		// 控制启动速率
		if i%10 == 0 {
			time.Sleep(50 * time.Millisecond)
		}
	}

	// 等待所有请求完成
	wg.Wait()
	duration := time.Since(startTime)

	// 输出结果
	totalRequests := successCount + failCount
	fmt.Printf("\n========= 压测结果 =========\n")
	fmt.Printf("总请求数: %d\n", totalRequests)
	fmt.Printf("成功请求: %d (%.2f%%)\n", successCount, float64(successCount*100)/float64(totalRequests))
	fmt.Printf("失败请求: %d (%.2f%%)\n", failCount, float64(failCount*100)/float64(totalRequests))
	fmt.Printf("平均延迟: %.2f ms\n", float64(totalLatency)/float64(totalRequests))
	fmt.Printf("总用时: %.2f 秒\n", duration.Seconds())
	fmt.Printf("QPS: %.2f 请求/秒\n", float64(totalRequests)/duration.Seconds())
	fmt.Printf("============================\n")
}
