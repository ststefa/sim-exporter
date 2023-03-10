# syntax=docker/dockerfile:1.3

# Do not name this file "Dockerfile.something" because the ALM process picks up
# any Dockerfile*

# alpine lacks upx package for aarch64 (Apple Silicon)
FROM golang:1.18-alpine AS build
RUN apk --no-cache add git make curl
#FROM golang:1.18-bullseye AS build
#RUN apt-get update ; apt-get install -y --no-install-recommends git make curl ; rm -fr /var/lib/apt/lists/*
WORKDIR /app
# Copy just go mod and sum files first
COPY go.mod go.sum ./
# Download all dependencies so that the layer will be cached unless go.{mod,sum}
# change. A netrc needs to be bind-mounted because go uses git which requires
# auth info when downloading private modules.
RUN --mount=type=secret,id=netrc,target=/root/.netrc go mod download
# Copy and build
COPY . .
RUN --mount=type=secret,id=netrc,target=/root/.netrc make build

# The remaining part should be similar to the regular `Dockerfile`
FROM alpine
WORKDIR /
COPY --from=build /app/build/sim-exporter .
COPY --from=build /app/build/examples examples
ENTRYPOINT ["/sim-exporter"]
CMD ["help"]
