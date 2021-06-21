package config

import (
	pb "github.com/envoyproxy/go-control-plane/envoy/service/ratelimit/v3"
	"github.com/rs/zerolog/log"
	"strings"
)

type yamlRateLimit struct {
	RequestsPerUnit uint32 `yaml:"requests_per_unit"`
	Unit            string
}

func (y *yamlRateLimit) ToRateLimit(key string) (*RateLimit, error) {
	if y == nil {
		return nil, nil
	}
	unit :=
		pb.RateLimitResponse_RateLimit_Unit_value[strings.ToUpper(y.Unit)]
	if unit == int32(pb.RateLimitResponse_RateLimit_UNKNOWN) {
		return nil, ErrInvalidUnit
	}
	return NewRateLimit(
		y.RequestsPerUnit, pb.RateLimitResponse_RateLimit_Unit(unit), key), nil
}

type yamlDescriptor struct {
	Key         string
	Value       string
	RateLimit   *yamlRateLimit `yaml:"rate_limit"`
	Descriptors []yamlDescriptor
}

func (conf *yamlDescriptor) ToDescriptor(parent *Descriptor) (*Descriptor, error) {
	if conf.Key == "" {
		return nil, ErrEmptyDescriptor
	}

	// Value is optional, so the final key for the map is either the key only or key_value.
	key := conf.Key
	if conf.Value != "" {
		key += "_" + conf.Value
	}
	if _, present := parent.Descriptors[key]; present {
		return nil, ErrDuplicateDescriptor
	}

	finalKey := parent.FullKey + "." + key
	rateLimit, err := conf.RateLimit.ToRateLimit(finalKey)
	if err != nil {
		return nil, err
	}
	log.Debug().Msgf(
		"loading descriptor: key=%s %s", finalKey, (*DebugLimit)(rateLimit))

	descriptor := &Descriptor{Descriptors: map[string]*Descriptor{}, Key: key, FullKey: finalKey, Limit: rateLimit}
	err = descriptor.loadDescriptors(conf.Descriptors)
	if err != nil {
		return nil, err
	}
	return descriptor, nil
}

type YamlFile struct {
	Domain      string
	Descriptors []yamlDescriptor
}
