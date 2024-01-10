FROM golang:1.21-alpine AS builder

WORKDIR $GOPATH/src/github.com/wakurobotics/beatflip

RUN apk add --no-cache git

COPY ./ ./

# Set the environment variables for the go command:
# * CGO_ENABLED=0 to build a statically-linked executable
# * GOFLAGS=-mod=vendor to force `go build` to look into the `/vendor` folder.
ENV CGO_ENABLED=0 GOFLAGS=-mod=vendor

RUN VERSION=$(git describe --tags --abbrev=0) && \
  go build -v \
  -installsuffix 'static' \
  -ldflags="-X github.com/wakurobotics/beatflip/cmd.version=${VERSION}" \
  -o /app/beatflip .

# ----
# 
FROM alpine:latest AS final

RUN apk add --no-cache ca-certificates tzdata && \
  apk upgrade --no-cache libcrypto3 libssl3

WORKDIR /app

COPY --from=builder /app /app

# opt-out of root
RUN addgroup -S wakurobotics && adduser -S wakurobotics -G wakurobotics && \
  chown -R wakurobotics:wakurobotics /app
USER wakurobotics

ENTRYPOINT ["./beatflip"]