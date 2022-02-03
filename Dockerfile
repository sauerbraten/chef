FROM golang:1-alpine as builder

COPY . /src/

RUN cd /src && \
    CGO_ENABLED=0 go build ./cmd/chef


FROM gcr.io/distroless/base

COPY                ./config.fly.json /config.json
COPY                ./migrations      /migrations
COPY --from=builder /src/chef         /

CMD ["./chef"]
