FROM golang:1.16.5 AS build
WORKDIR /ratelimit
RUN go env && sed -i 's/_StackMin = 2048/_StackMin = 8192/g'  /usr/local/go/src/runtime/stack.go && cat /usr/local/go/src/runtime/stack.go
ARG GOPROXY=https://goproxy.io,direct
ARG GOARCH=amd64
COPY go.mod go.sum /ratelimit/
COPY third_party /ratelimit/third_party
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /go/bin/ratelimit  -v github.com/istio-conductor/shard-ratelimit

FROM alpine:3.11 AS final
COPY --from=build /go/bin/ratelimit /bin/ratelimit
ENTRYPOINT /bin/ratelimit