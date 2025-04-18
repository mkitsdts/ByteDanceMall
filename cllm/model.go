package cllm

import (
	"net/http"
	"sync"
	"time"
)

const (
	MAX_CONNECTIONS = 10
	MAX_IDLE_TIME   = 60 * time.Second
)

type ModelService struct {
	Client []*http.Client
	mux    sync.Mutex
	Host   string
	Name   string
}

func (s *ModelService) GetHttpClient() *http.Client {
	s.mux.Lock()
	defer s.mux.Unlock()
	var client *http.Client
	if len(s.Client) != 0 {
		client = s.Client[len(s.Client)-1]
		s.Client = s.Client[:len(s.Client)-1]
		return client
	}
	for {
		if len(s.Client) > 0 {
			client = s.Client[len(s.Client)-1]
			s.Client = s.Client[:len(s.Client)-1]
			return client
		}
	}
}

func (s *ModelService) ReturnHttpClient(c *http.Client) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.Client = append(s.Client, c)
}
