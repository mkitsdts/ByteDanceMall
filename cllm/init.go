package cllm

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func (s *ModelService) try_get_model_info() bool {
	resp, err := s.Client[0].Get(s.Host + "/v1/models/" + s.Name)
	if err != nil {
		return false
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return false
	}
	fmt.Println(string(body))
	return resp.StatusCode == http.StatusOK
}

func (s *ModelService) Init(name string, host string) {
	s.Name = name
	s.Host = host
	s.Client = make([]*http.Client, MAX_CONNECTIONS)
	for i := range MAX_CONNECTIONS {
		s.Client[i] = &http.Client{}
		s.Client[i].Timeout = MAX_IDLE_TIME
		s.Client[i].Transport = &http.Transport{
			MaxIdleConns:          10,
			IdleConnTimeout:       30 * time.Second,
			DisableCompression:    true,
			ForceAttemptHTTP2:     true,
			ExpectContinueTimeout: 10 * time.Second,
		}
	}
	if !s.try_get_model_info() {
		panic("Failed to connect to model server")
	} else {
		fmt.Println("Model server connected successfully")
	}
}
