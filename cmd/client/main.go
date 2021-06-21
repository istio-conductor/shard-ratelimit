package main

import (
	"flag"
	"fmt"
	pb_struct "github.com/envoyproxy/go-control-plane/envoy/extensions/common/ratelimit/v3"
	pb "github.com/envoyproxy/go-control-plane/envoy/service/ratelimit/v3"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var host string
var domain string
var kv string
var nConcurrency int
var nReqPerConn int
var timeout int

func constructRequest() *pb.RateLimitRequest {
	desc := make([]*pb_struct.RateLimitDescriptor, 1)
	entries := make([]*pb_struct.RateLimitDescriptor_Entry, 1)
	keyAndValue := strings.Split(kv, ":")
	entries[0] = &pb_struct.RateLimitDescriptor_Entry{Key: keyAndValue[0], Value: keyAndValue[1]}
	desc[0] = &pb_struct.RateLimitDescriptor{
		Entries: entries,
	}
	ratelimitReq := &pb.RateLimitRequest{
		Domain:      domain,
		Descriptors: desc,
		HitsAddend:  1,
	}
	return ratelimitReq
}

func main() {
	flag.StringVar(&host, "h", "", "")
	flag.StringVar(&domain, "d", "", "")
	flag.StringVar(&kv, "k", "", "")
	flag.IntVar(&nConcurrency, "c", 5, "")
	flag.IntVar(&nReqPerConn, "n", 100, "")
	flag.IntVar(&timeout, "t", 30, "")

	flag.Parse()

	fmt.Println("host:", host)
	fmt.Println("domain:", domain)
	fmt.Println("kv:", kv)
	fmt.Println("concurrency:", nConcurrency)
	fmt.Println("request per conn:", nReqPerConn)
	fmt.Println("timeout:", timeout, "ms")

	conns := make([]*grpc.ClientConn, 0)
	for index := 0; index < nConcurrency; index++ {
		conn, err := grpc.Dial(host, grpc.WithInsecure())
		if err != nil {
			fmt.Printf("error connecting: %s\n", err.Error())
			os.Exit(1)
		}
		conns = append(conns, conn)
	}

	latency := make([]time.Duration, nConcurrency)
	successNum := make([]int, nConcurrency)

	timeCostPerReq := make([][]int64, nConcurrency)
	for index := 0; index < nConcurrency; index++ {
		timeCostPerReq[index] = make([]int64, nReqPerConn)
	}

	wg := sync.WaitGroup{}
	wg.Add(nConcurrency)
	ratelimitReq := constructRequest()
	startTime := time.Now()
	var okSize int64
	for index := 0; index < nConcurrency; index++ {
		go func(index int) {
			conn := conns[index]
			c := pb.NewRateLimitServiceClient(conn)
			for i := 0; i < nReqPerConn; i++ {
				ctx, _ := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(timeout))
				t := time.Now()
				resp, err := c.ShouldRateLimit(ctx, ratelimitReq)
				if err == nil {
					successNum[index]++
					if resp.OverallCode == pb.RateLimitResponse_OK {
						atomic.AddInt64(&okSize, 1)
					}
				}

				cost := time.Since(t)
				latency[index] += cost
				timeCostPerReq[index][i] = cost.Microseconds()
			}
			wg.Done()
		}(index)
	}
	wg.Wait()
	timeCost := time.Since(startTime)
	for index := 0; index < nConcurrency; index++ {
		_ = conns[index].Close()
	}

	var latencyTotal int64
	var successNumTotal int
	totalNum := nReqPerConn * nConcurrency
	tmp := make([]int, 0)
	for index := 0; index < nConcurrency; index++ {
		latencyTotal += latency[index].Microseconds()
		successNumTotal += successNum[index]
		for i := 0; i < nReqPerConn; i++ {
			tmp = append(tmp, int(timeCostPerReq[index][i]))
		}
	}
	sort.Ints(tmp)

	fmt.Println("Time taken for tests: ", timeCost.Seconds(), "s")
	fmt.Println("Success num: ", successNumTotal)
	fmt.Println("Fail num: ", totalNum-successNumTotal)
	fmt.Println("Fail ratio:", float64(totalNum-successNumTotal)/float64(totalNum)*100, "%")
	fmt.Println("OK per second: ", float64(okSize)/timeCost.Seconds())
	fmt.Println("Requests per second: ", float64(totalNum)/timeCost.Seconds())
	fmt.Println("Latency per request: ", latencyTotal/int64(totalNum), "us")
	fmt.Println("P50: ", tmp[totalNum/2], "us")
	fmt.Println("P99: ", tmp[int(float64(totalNum*99)/100.0)], "us")
	fmt.Println("P999: ", tmp[int(float64(totalNum*999)/1000.0)], "us")
	fmt.Println("MAX: ", tmp[len(tmp)-1])
}
