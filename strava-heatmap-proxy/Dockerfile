# syntax=docker/dockerfile:1.7

# --- Build stage ---
FROM golang:1.24 AS build
COPY ./ /src
WORKDIR /src
RUN CGO_ENABLED=0 go install -ldflags "-X main.version=$(cat version)" ./cmd/strava-heatmap-proxy

# --- Runtime stage ---
FROM gcr.io/distroless/base-debian13:nonroot
WORKDIR /app
COPY --from=build /go/bin/strava-heatmap-proxy ./
EXPOSE 8080
USER nonroot:nonroot

ENTRYPOINT ["/app/strava-heatmap-proxy"]
