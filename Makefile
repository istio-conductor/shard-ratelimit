.PHONY: build
platform=--platform=linux/amd64
build:
	go build -o ./bin/ratelimit
docker:
	docker build platform -t istioconductor/ratelimit:latest .