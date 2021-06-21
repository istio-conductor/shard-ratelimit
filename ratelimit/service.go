package ratelimit

import (
	"context"
	pb "github.com/envoyproxy/go-control-plane/envoy/service/ratelimit/v3"
	"github.com/istio-conductor/shard-ratelimit/bucket"
	"github.com/istio-conductor/shard-ratelimit/config"
	"github.com/istio-conductor/shard-ratelimit/prom"
	log "github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
	"sync/atomic"
)

type Service struct {
	config       atomic.Value
	limiter      *bucket.Buckets
	mutex        sync.Mutex
	replicas     int32
	fileContents map[string][]byte
}

func (s *Service) OnReplicasUpdate(replicas int32) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.replicas = replicas
	s.reload()
}

func (s *Service) OnConfigUpdate(fileContents map[string][]byte) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.fileContents = fileContents
	s.reload()
}

func (s *Service) reload() {
	var files []config.File
	for name, bytes := range s.fileContents {
		files = append(files, config.File{Name: name, Content: bytes})
	}

	newConfig, err := config.New(s.replicas, files)
	if err != nil {
		prom.ConfigLoadError.Inc()
		log.Error().Err(err).Msg("load config failed")
		return
	}
	prom.ConfigLoadSuccess.Inc()
	s.config.Store(newConfig)
	limits := newConfig.KeyLimits()
	log.Info().Msgf("key limits: %v", limits)
	s.limiter.Update(limits)
}

var (
	ErrEmptyDomain      = status.Error(codes.InvalidArgument, "rate limit domain must not be empty")
	ErrEmptyDescriptors = status.Error(codes.InvalidArgument, "rate limit descriptor list must not be empty")
	ErrNoConfiguration  = status.Errorf(codes.Internal, "no rate limit configuration loaded")
)

func check(request *pb.RateLimitRequest) error {
	if request.Domain == "" {
		return ErrEmptyDomain
	}
	if len(request.Descriptors) == 0 {
		return ErrEmptyDescriptors
	}
	return nil
}

func (s *Service) shouldRateLimit(
	ctx context.Context, request *pb.RateLimitRequest) (*pb.RateLimitResponse, error) {
	if err := check(request); err != nil {
		return nil, err
	}
	conf := s.Config()
	if conf == nil {
		return nil, ErrNoConfiguration
	}

	limitsToCheck := make([]*config.RateLimit, len(request.Descriptors))

	for i, descriptor := range request.Descriptors {
		limit, err := conf.GetLimit(ctx, request.Domain, descriptor)
		if err != nil {
			return nil, err
		}
		log.Debug().Msgf("descriptor: %s", entries(descriptor.GetEntries()))
		limitsToCheck[i] = limit
		log.Debug().Msgf("limit: %s", (*config.DebugLimit)(limit))
	}

	statuses := s.limiter.DoLimit(ctx, request, limitsToCheck)

	response := &pb.RateLimitResponse{
		Statuses:    statuses,
		OverallCode: pb.RateLimitResponse_OK,
	}
	for _, s := range statuses {
		if s.Code == pb.RateLimitResponse_OVER_LIMIT {
			response.OverallCode = s.Code
		}
	}
	return response, nil
}

func (s *Service) ShouldRateLimit(
	ctx context.Context,
	request *pb.RateLimitRequest) (finalResponse *pb.RateLimitResponse, err error) {
	response, err := s.shouldRateLimit(ctx, request)
	log.Debug().Msg("returning normal response")
	return response, err
}

func (s *Service) Config() *config.Config {
	return s.config.Load().(*config.Config)
}

func New(limiter *bucket.Buckets) *Service {
	return &Service{
		limiter: limiter,
	}
}
