package strava

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

type CookieClient interface {
	AddCookies(*http.Request)
}

type StravaProxy struct {
	httputil.ReverseProxy
}

func NewStravaProxy(client CookieClient) *StravaProxy {
	target, _ := url.Parse("https://heatmap-external-a.strava.com/")
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host
		client.AddCookies(req)
	}
	return &StravaProxy{
		httputil.ReverseProxy{Director: director},
	}
}
