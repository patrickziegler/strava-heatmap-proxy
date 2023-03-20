build:
	go build -o build/strava-heatmap-auth cmd/strava-heatmap-auth/strava-heatmap-auth.go
	go build -o build/strava-heatmap-proxy cmd/strava-heatmap-proxy/strava-heatmap-proxy.go

install:
	GOPATH=~/.local go install cmd/strava-heatmap-auth/strava-heatmap-auth.go
	GOPATH=~/.local go install cmd/strava-heatmap-proxy/strava-heatmap-proxy.go

clean:
	rm -rf build
