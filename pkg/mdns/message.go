package mdns

import (
	"net"

	// Modules
	dns "github.com/miekg/dns"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type message struct {
	*dns.Msg
	Addr  net.Addr
	Index int
	Zone  string
	Err   error
}
