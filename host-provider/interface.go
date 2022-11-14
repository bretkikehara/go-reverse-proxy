package hostprovider

type HostProvider interface {
	AddHost(host string) error
	RemoveHost(host string) error
	AddSubdomain(subdomain string) error
	RemoveSubdomain(subdomain string) error
	ListSubdomains() []string

	RemoveTLD(tld string) error
	GetTLD() string
}
