# syntax=docker/dockerfile:1.7

# --- Build stage ---
FROM golang:1.22 AS build
WORKDIR /src
# Build the proxy binary from the module ROOT (no cmd/… suffix)
RUN CGO_ENABLED=0 go install github.com/patrickziegler/strava-heatmap-proxy@latest

# --- Runtime stage ---
FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app
COPY --from=build /go/bin/strava-heatmap-proxy /usr/local/bin/strava-heatmap-proxy
EXPOSE 8080
USER nonroot:nonroot
# Looks for cookies at /config/strava-cookies.json by default (mount it read-only)
ENTRYPOINT ["/usr/local/bin/strava-heatmap-proxy"]
CMD ["--port","8080","--cookies","/config/strava-cookies.json"]