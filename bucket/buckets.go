package bucket

import (
	pb "github.com/envoyproxy/go-control-plane/envoy/service/ratelimit/v3"
	"github.com/istio-conductor/shard-ratelimit/config"
	"golang.org/x/net/context"
	"golang.org/x/time/rate"
	"math"
)

type Buckets struct {
	limiter map[string]*rate.Limiter
}

func New() *Buckets {
	return &Buckets{map[string]*rate.Limiter{}}
}

var (
	OK = &pb.RateLimitResponse_DescriptorStatus{
		Code: pb.RateLimitResponse_OK,
	}
	FAIL = &pb.RateLimitResponse_DescriptorStatus{
		Code: pb.RateLimitResponse_OVER_LIMIT,
	}
	UNKNOWN = &pb.RateLimitResponse_DescriptorStatus{
		Code: pb.RateLimitResponse_UNKNOWN,
	}
)

func (b *Buckets) Update(limits map[string]float64) {
	m := make(map[string]*rate.Limiter, len(limits))
	for k, f := range limits {
		m[k] = rate.NewLimiter(rate.Limit(f), int(math.Ceil(f)))
	}
	b.limiter = m
}

func (b *Buckets) DoLimit(ctx context.Context, request *pb.RateLimitRequest, limits []*config.RateLimit) []*pb.RateLimitResponse_DescriptorStatus {
	resp := make([]*pb.RateLimitResponse_DescriptorStatus, 0, len(limits))
	for _, limit := range limits {
		if limit == nil {
			resp = append(resp, UNKNOWN)
			continue
		}
		l, ok := b.limiter[limit.FullKey]
		if !ok {
			resp = append(resp, UNKNOWN)
			continue
		}
		if l.Allow() {
			resp = append(resp, OK)
		} else {
			resp = append(resp, FAIL)
		}
	}
	return resp
}
