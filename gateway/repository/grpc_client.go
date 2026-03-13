package repository

import (
	"fmt"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type grpcClient struct {
	target string
	mu     sync.Mutex
	conn   *grpc.ClientConn
}

func newGRPCClient(host string, port int) *grpcClient {
	return &grpcClient{
		target: fmt.Sprintf("%s:%d", host, port),
	}
}

func (c *grpcClient) Conn() (*grpc.ClientConn, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return c.conn, nil
	}

	conn, err := grpc.NewClient(c.target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	c.conn = conn
	return c.conn, nil
}
