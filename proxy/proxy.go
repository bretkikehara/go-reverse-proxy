package proxy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	hostprovider "github.com/bretkikehara/go-reverse-proxy/host-provider"
)

type ReverseProxy struct {
	provider hostprovider.HostProvider
	targets  []*target
}

type TargetConfig struct {
	Subdomain string
	Host      string
	Port      int
	Secure    bool
}

func (c *TargetConfig) isSecure() bool {
	return c.Secure || c.Port == 443
}

func (c *TargetConfig) getURL() (*url.URL, error) {
	proto := "http"
	if c.isSecure() {
		proto = "https"
	}

	u := fmt.Sprintf("%s://%s", proto, c.Host)
	if c.Port != 0 && c.Port != 80 && c.Port != 443 {
		u += fmt.Sprintf(":%d", c.Port)
	}

	url, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	return url, nil
}

func (c *TargetConfig) createTarget() (*target, error) {
	url, err := c.getURL()
	if err != nil {
		return nil, err
	}
	t := target{
		subdomain: c.Subdomain,
		proxy:     httputil.NewSingleHostReverseProxy(url),
	}
	return &t, nil
}

type target struct {
	subdomain string
	proxy     *httputil.ReverseProxy
}

// New creates a reverse proxy
func New(provider hostprovider.HostProvider, targetConfigs []TargetConfig) *ReverseProxy {
	var targets []*target
	for i := range targetConfigs {
		cfg := targetConfigs[i]
		t, err := cfg.createTarget()
		if err != nil {
			panic(err)
		}
		targets = append(targets, t)
		provider.AddSubdomain(cfg.Subdomain)
	}
	return &ReverseProxy{
		provider: provider,
		targets:  targets,
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

func getSubdomain(h string) string {
	return strings.SplitN(h, ".", 2)[0]
}

func (pxy *ReverseProxy) handleReverseProxy(w http.ResponseWriter, r *http.Request) {
	subdomain := getSubdomain(r.Host)
	for _, t := range pxy.targets {
		if subdomain == t.subdomain {
			t.proxy.ServeHTTP(w, r)
			break
		}
	}
}
