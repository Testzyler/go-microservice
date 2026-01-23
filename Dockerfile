# syntax=docker/dockerfile:1

# syntax=docker/dockerfile:1
ARG GO_VERSION=1.25

FROM golang:${GO_VERSION} AS build

ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the single binary CLI; commands are selected via args on run (serve-internal-gateway, serve-auth, serve-auth-worker, serve-auth-scheduler, serve-all, etc.)
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
  go build -buildvcs=false -trimpath -ldflags="-s -w" -o /out/app .

FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /
COPY --from=build /out/app /app
# Embed migrations for the migrate job; place away from /app binary.
COPY --from=build /src/services/auth/migrations /migrations/auth
USER nonroot

# Override with e.g. `docker run <img> serve-auth` or other subcommands.
ENTRYPOINT ["/app"]
CMD ["serve-all"]
