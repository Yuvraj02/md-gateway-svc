# Build from this service repo root:
#   docker build -t marketing-digest-gateway .

FROM golang:1.25-alpine AS build
WORKDIR /src
RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
COPY pkg ./pkg
COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/gateway ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/gateway /gateway
USER nobody
EXPOSE 8080
ENTRYPOINT ["/gateway"]
