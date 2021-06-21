High Performance RateLimit Server
---
A high performance version of envoy's [ratelimit](https://github.com/envoyproxy/ratelimit/) server. 

## Quick start
```bash
git clone https://github.com/istio-conductor/shard-ratelimit.git
helm upgrade -i -n istio-system prod ./helm/shard-ratelimit
```

## Performance