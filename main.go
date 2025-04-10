package main

import (
	"bytedancemall/seckill/service"
	"encoding/json"
	"fmt"
	"net/http"
)

func main() {
	s := service.NewSeckillService()
	fmt.Println("Seckill service started...")
	// 启动 http 服务
	http.HandleFunc("/seckill", func(w http.ResponseWriter, r *http.Request) {
		type Request struct {
			ProductId   uint32  `json:"product_id"`
			Quantity    uint32 `json:"quantity"`
			ReleaseTime string `json:"release_time"`
		}
		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		s.AddItemHandler(r.Context(), req.ProductId, req.Quantity,req.ReleaseTime)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	})
	http.HandleFunc("/seckill/tryseckill", func(w http.ResponseWriter, r *http.Request) {
		type Request struct {
			ProductId uint32 `json:"product_id"`
			UserId    uint32 `json:"user_id"`
		}
		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			fmt.Println("Error decoding request:", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := s.TrySecKillItemHandler(r.Context(), req.ProductId, req.UserId); err != nil {
			fmt.Println("Error processing request:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	})
	http.ListenAndServe(":8080", nil)
}