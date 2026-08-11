# Build from workspace root (parent of backend/ and protos/):
#   docker build -f backend/gateway/Dockerfile -t marketing-digest-gateway .
FROM golang:1.25-alpine AS build
WORKDIR /src
RUN apk add --no-cache ca-certificates git

COPY protos ./protos
COPY backend/pkg ./backend/pkg
COPY backend/gateway ./backend/gateway

WORKDIR /src/backend/gateway
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/gateway ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/gateway /gateway
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/gateway"]
