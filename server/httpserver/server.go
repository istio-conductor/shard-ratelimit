package httpserver

import (
	"context"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/atomic"
	"net/http"
	_ "net/http/pprof"
	"strconv"
)

var health = atomic.NewBool(false)

func Health() {
	health.Store(true)
}

func NotHealth() {
	health.Store(false)
}

func Run(ctx context.Context, HTTPPort int) error {
	server := &http.Server{Addr: ":" + strconv.Itoa(HTTPPort)}
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/check_health", func(writer http.ResponseWriter, request *http.Request) {
		if health.Load() {
			writer.WriteHeader(200)
			_, _ = writer.Write([]byte("ok"))
			return
		}
		writer.WriteHeader(500)
		_, _ = writer.Write([]byte("not ok"))
		return
	})
	go func() {
		<-ctx.Done()
		_ = server.Shutdown(context.TODO())
	}()

	return server.ListenAndServe()
}
