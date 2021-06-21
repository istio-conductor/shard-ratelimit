package server

import (
	"context"
	v3 "github.com/envoyproxy/go-control-plane/envoy/service/ratelimit/v3"
	"github.com/istio-conductor/shard-ratelimit/bucket"
	"github.com/istio-conductor/shard-ratelimit/prom"
	"github.com/istio-conductor/shard-ratelimit/ratelimit"
	"github.com/istio-conductor/shard-ratelimit/reloader"
	"github.com/istio-conductor/shard-ratelimit/reloader/configmap"
	"github.com/istio-conductor/shard-ratelimit/replicas"
	"github.com/istio-conductor/shard-ratelimit/server/httpserver"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"net"
	"strconv"
)

type Server struct {
	Replicas  int
	Namespace string
	Service   string
	Port      int
	HTTPPort  int
	Dir       string
	ConfigMap string
}

func New(port int, httpPort int, dir string, ns, svc string, cm string, replicas int) *Server {
	return &Server{Port: port, HTTPPort: httpPort, Dir: dir, Namespace: ns, Service: svc, ConfigMap: cm, Replicas: replicas}
}

func (s *Server) Run(ctx context.Context) error {

	group, ctx := errgroup.WithContext(ctx)

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(s.Port))
	if err != nil {
		return err
	}

	server := grpc.NewServer(grpc.ChainUnaryInterceptor(prom.MiddleWare))

	buckets := bucket.New()

	service := ratelimit.New(buckets)
	if s.Replicas == 0 {
		r, err := replicas.New(s.Namespace, s.Service, service.OnReplicasUpdate)
		if err != nil {
			return err
		}
		group.Go(func() error {
			return r.Run(ctx)
		})
	} else {
		service.OnReplicasUpdate(int32(s.Replicas))
	}

	if s.ConfigMap != "" {
		cm, err := configmap.New(s.Namespace, s.ConfigMap, service.OnConfigUpdate)
		if err != nil {
			return err
		}
		group.Go(func() error {
			return cm.Run(ctx)
		})
	} else {
		loader, err := reloader.New(s.Dir, service.OnConfigUpdate)
		if err != nil {
			return err
		}
		loader.LoadOnce()
		group.Go(func() error {
			return loader.Watch(ctx)
		})
	}

	v3.RegisterRateLimitServiceServer(server, service)

	group.Go(func() error {
		<-ctx.Done()
		httpserver.NotHealth()
		server.GracefulStop()
		return ctx.Err()
	})

	group.Go(func() error {
		httpserver.Health()
		return server.Serve(listener)
	})

	group.Go(func() error {
		return httpserver.Run(ctx, s.HTTPPort)
	})

	return group.Wait()
}
