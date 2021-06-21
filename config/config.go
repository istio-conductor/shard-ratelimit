package config

import (
	"context"
	"errors"
	pb_struct "github.com/envoyproxy/go-control-plane/envoy/extensions/common/ratelimit/v3"
	pb "github.com/envoyproxy/go-control-plane/envoy/service/ratelimit/v3"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
	"strconv"
)

// RateLimit is a wrapper for an individual rate limit config entry which includes the defined limit and metrics.
type RateLimit struct {
	FullKey string
	Metrics Metrics
	Limit   *pb.RateLimitResponse_RateLimit
}

type DebugLimit RateLimit

func (l *DebugLimit) String() string {
	if l == nil {
		return ""
	}
	return strconv.FormatInt(int64(l.Limit.RequestsPerUnit), 10) + "/" + l.Limit.Unit.String()
}

type File struct {
	Name    string
	Content []byte
}

var (
	ErrNoDomain                     = errors.New("no domain in config file")
	ErrDuplicate                    = errors.New("duplicate domain in config file")
	ErrEmptyDescriptor              = errors.New("descriptor has empty key")
	ErrDuplicateDescriptor          = errors.New("duplicate descriptor")
	ErrInvalidUnit                  = errors.New("invalid unit")
	ErrUnsupportedRateLimitOverride = errors.New("unsupported ratelimit override")
)

type Config struct {
	domains map[string]*Domain
}

// NewRateLimit Create a new rate limit config entry.
func NewRateLimit(
	requestsPerUnit uint32, unit pb.RateLimitResponse_RateLimit_Unit, key string) *RateLimit {
	return &RateLimit{FullKey: key, Metrics: NewMetrics(key), Limit: &pb.RateLimitResponse_RateLimit{RequestsPerUnit: requestsPerUnit, Unit: unit}}
}

type Descriptor struct {
	Key         string
	FullKey     string
	Descriptors map[string]*Descriptor
	Limit       *RateLimit
}

func (d *Descriptor) KeyLimits(keys map[string]float64) {
	if d.Limit != nil {
		keys[d.Limit.FullKey] = float64(int(d.Limit.Limit.RequestsPerUnit))
	}
	for _, child := range d.Descriptors {
		child.KeyLimits(keys)
	}
}

type Domain struct {
	Descriptor
}

// Load a set of config descriptors from the YAML file and check the input.
func (d *Descriptor) loadDescriptors(descriptors []yamlDescriptor) error {
	for _, conf := range descriptors {
		descriptor, err := conf.ToDescriptor(d)
		if err != nil {
			return err
		}
		d.Descriptors[descriptor.Key] = descriptor
	}
	return nil
}

func (c *Config) loadConfig(config File) error {
	var root YamlFile
	err := yaml.UnmarshalStrict(config.Content, &root)
	if err != nil {
		return err
	}

	if root.Domain == "" {
		return ErrNoDomain
	}

	if _, present := c.domains[root.Domain]; present {
		return ErrDuplicate
	}

	log.Debug().Msgf("loading domain: %s", root.Domain)
	domain := &Domain{Descriptor{FullKey: root.Domain, Descriptors: map[string]*Descriptor{}}}
	err = domain.loadDescriptors(root.Descriptors)
	if err != nil {
		return err
	}
	c.domains[root.Domain] = domain
	return nil
}

func (c *Config) KeyLimits() map[string]float64 {
	m := map[string]float64{}
	for _, domain := range c.domains {
		for _, descriptor := range domain.Descriptors {
			descriptor.KeyLimits(m)
		}
	}
	return m
}

func (c *Config) GetLimit(
	_ context.Context, domain string, descriptor *pb_struct.RateLimitDescriptor) (rateLimit *RateLimit, err error) {
	domainLimits := c.domains[domain]
	if domainLimits == nil {
		log.Debug().Msgf("unknown domain '%s'", domain)
		return
	}

	if descriptor.GetLimit() != nil {
		return nil, ErrUnsupportedRateLimitOverride
	}

	descriptors := domainLimits.Descriptors
	for i, entry := range descriptor.Entries {
		key := entry.Key + "_" + entry.Value
		next := descriptors[key]
		if next == nil {
			key = entry.Key
			next = descriptors[key]
		}
		if next == nil {
			return
		}
		if next.Limit != nil && i == len(descriptor.Entries)-1 {
			log.Debug().Msgf("found rate limit: %s", key)
			return next.Limit, nil
		}
		if len(next.Descriptors) == 0 {
			return
		}
		descriptors = next.Descriptors
	}
	return
}

// New create rate limit config from a list of input YAML files.
func New(replicas int32, configs []File) (*Config, error) {
	c := &Config{domains: map[string]*Domain{}}
	for _, config := range configs {
		err := c.loadConfig(config)
		if err != nil {
			log.Error().Err(err).Msgf("load config failed: %s", config.Name)
			continue
		}
	}
	if replicas == 1 || replicas == 0 {
		return c, nil
	}
	divideBy(c, uint32(replicas))
	log.Info().Msgf("request unit is divide by replicas %d", replicas)
	return c, nil
}

func divideBy(c *Config, replicas uint32) {
	for _, rc := range c.domains {
		divideRPBy(&rc.Descriptor, replicas)
	}
}

func divideRPBy(r *Descriptor, replicas uint32) {
	if r.Limit != nil {
		r.Limit.Limit.RequestsPerUnit /= replicas
	}
	for _, des := range r.Descriptors {
		divideRPBy(des, replicas)
	}
}
