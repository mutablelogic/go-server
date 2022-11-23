package plugin

import "net"

// DNSRegister provides a mechanism to register DNS records to a remote dynamic DNS service
type DNSRegister interface {
	GetExternalAddress() (net.IP, error)
	RegisterAddress(host, user, password string, addr net.IP, offline bool) error
}
