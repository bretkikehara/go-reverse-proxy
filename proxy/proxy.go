package proxy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	hostprovider "github.com/bretkikehara/go-reverse-proxy/host-provider"
)

type ReverseProxy struct {
	provider hostprovider.HostProvider
}

// New creates a reverse proxy
func New(provider hostprovider.HostProvider) *ReverseProxy {
	return &ReverseProxy{
		provider: provider,
	}
}

func (pxy *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if GetHost(r) == pxy.provider.GetTLD() {
		pxy.handleAdmin(w, r)
		return
	}
	pxy.handleReverseProxy(w, r)
}

func (pxy *ReverseProxy) handleAdmin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if r.URL.Path == "/add" {
			if err := pxy.handleAdminAdd(w, r); err != nil {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("bad request: " + err.Error()))
				return
			}
			return
		}
		if r.URL.Path == "/remove" {
			if err := pxy.handleAdminRemove(w, r); err != nil {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("bad request: " + err.Error()))
				return
			}
			return
		}
	} else if r.Method == http.MethodGet {
		if r.URL.Path == "/list" {
			if err := pxy.handleAdminList(w, r); err != nil {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("bad request: " + err.Error()))
				return
			}
			return
		}
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("not found"))
}

type adminAddRequest struct {
	Subdomain string
}

func (pxy *ReverseProxy) handleAdminAdd(w http.ResponseWriter, r *http.Request) error {
	defer func() {
		if r.Body != nil {
			r.Body.Close()
		}
	}()
	var ask adminAddRequest
	if err := json.NewDecoder(r.Body).Decode(&ask); err != nil {
		return err
	}
	return pxy.provider.AddSubdomain(ask.Subdomain)
}

type admimRemoveRequest struct {
	Subdomain string
}

func (pxy *ReverseProxy) handleAdminRemove(w http.ResponseWriter, r *http.Request) error {
	defer func() {
		if r.Body != nil {
			r.Body.Close()
		}
	}()
	var ask admimRemoveRequest
	if err := json.NewDecoder(r.Body).Decode(&ask); err != nil {
		return err
	}
	return pxy.provider.RemoveSubdomain(ask.Subdomain)
}

func (pxy *ReverseProxy) handleAdminList(w http.ResponseWriter, r *http.Request) error {
	for _, a := range pxy.provider.ListSubdomains() {
		if _, err := fmt.Fprintf(w, "%s\n", a); err != nil {
			return err
		}
	}
	return nil
}

func (pxy *ReverseProxy) handleReverseProxy(w http.ResponseWriter, r *http.Request) {
	url, err := url.Parse("http://localhost:8000/")
	if err != nil {
		panic(err)
	}
	httputil.NewSingleHostReverseProxy(url).ServeHTTP(w, r)
}
