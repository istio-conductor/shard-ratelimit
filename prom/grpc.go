package prom

import (
	"context"
	"google.golang.org/grpc"
	"time"
)

func MiddleWare(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	Requests.Inc()
	resp, err := handler(ctx, req)
	ResponseTime.Observe(float64(time.Since(start).Milliseconds()))
	return resp, err
}
