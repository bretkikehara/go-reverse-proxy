package etchosts

import (
	"errors"
	"strings"
	"sync"

	hostprovider "github.com/bretkikehara/go-reverse-proxy/host-provider"
	"github.com/goodhosts/hostsfile"
)

type etchostProvider struct {
	mu    sync.Mutex
	hfile *hostsfile.Hosts
	tld   string
}

func New(tld string) (hostprovider.HostProvider, error) {
	hfile, err := hostsfile.NewHosts()
	if err != nil {
		return nil, err
	}
	p := etchostProvider{
		hfile: hfile,
		tld:   tld,
	}
	if !p.hfile.HasHostname(p.tld) {
		if err := p.addHostRaw(p.tld); err != nil {
			return nil, err
		}
	}
	p.hfile.RemoveDuplicateHosts()
	if err := p.hfile.Flush(); err != nil {
		return nil, err
	}
	return &p, nil
}

func (p *etchostProvider) GetTLD() string {
	return p.tld
}

func (p *etchostProvider) formatSubdomain(subdomain string) string {
	out := subdomain
	if !strings.HasSuffix(subdomain, ".") {
		out += "."
	}
	return out + p.tld
}

func (p *etchostProvider) AddSubdomain(subdomin string) error {
	return p.AddHost(p.formatSubdomain(subdomin))
}

func (p *etchostProvider) AddHost(fullhost string) error {
	if !strings.HasSuffix(fullhost, p.tld) {
		return errors.New("provider host mismatch")
	}
	return p.addHostRaw(fullhost)
}

func (p *etchostProvider) addHostRaw(fullhost string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.hfile.Load(); err != nil {
		return err
	}
	p.hfile.Add("127.0.0.1", fullhost)
	return p.hfile.Flush()
}

func (p *etchostProvider) RemoveSubdomain(subdomin string) error {
	return p.RemoveHost(p.formatSubdomain(subdomin))
}

func (p *etchostProvider) RemoveTLD(tld string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.hfile.Load(); err != nil {
		return err
	}
	p.hfile.Remove("127.0.0.1", tld)
	for _, l := range p.hfile.Lines {
		for _, h := range l.Hosts {
			if strings.HasSuffix(h, tld) {
				p.hfile.Remove("127.0.0.1", h)
			}
		}
	}
	return p.hfile.Flush()
}

func (p *etchostProvider) RemoveHost(fullhost string) error {
	if !strings.HasSuffix(fullhost, p.tld) {
		return errors.New("provider host mismatch")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.hfile.Load(); err != nil {
		return err
	}
	p.hfile.Remove("127.0.0.1", fullhost)
	return p.hfile.Flush()
}

func (p *etchostProvider) ListSubdomains() []string {
	mp := make(map[string]bool)
	for _, l := range p.hfile.Lines {
		for _, h := range l.Hosts {
			if strings.HasSuffix(h, p.tld) {
				mp[h] = true
			}
		}
	}
	var hosts []string
	for h := range mp {
		hosts = append(hosts, h)
	}
	return hosts
}
