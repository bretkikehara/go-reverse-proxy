package proxy

import (
	"net/http"
	"strings"
)

func GetHost(r *http.Request) string {
	h := getHostInner(r)
	sp := strings.SplitN(h, ":", 2)
	return sp[0]
}

func getHostInner(r *http.Request) string {
	if xfh := r.Header.Get("X-Forwarded-Host"); xfh != "" {
		return xfh
	}
	return r.Host
}
