package ratelimit

import (
	"encoding/json"
	v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/common/ratelimit/v3"
)

type entries []*v3.RateLimitDescriptor_Entry

func (e entries) String() string {
	data, _ := json.Marshal(e)
	return string(data)
}