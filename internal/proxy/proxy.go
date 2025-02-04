package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

type CookieClient interface {
	GetCookies(url *url.URL) []*http.Cookie
	GetTarget() *url.URL
}

func NewReverseProxy(client CookieClient) *httputil.ReverseProxy {
	target := client.GetTarget()
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host
		for _, c := range client.GetCookies(req.URL) {
			req.AddCookie(c)
		}
	}
	return &httputil.ReverseProxy{Director: director}
}
