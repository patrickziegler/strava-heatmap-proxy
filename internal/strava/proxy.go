package strava

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

type CookieClient interface {
	AddCookies(*http.Request)
	GetTarget() *url.URL
}

type StravaProxy struct {
	httputil.ReverseProxy
}

func NewStravaProxy(client CookieClient) *StravaProxy {
	target := client.GetTarget()
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
