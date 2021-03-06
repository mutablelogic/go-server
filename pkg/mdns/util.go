package mdns

import (
	"fmt"
	"net"
	"os"
	"strings"
	"syscall"

	// Packages
	multierror "github.com/hashicorp/go-multierror"
	dns "github.com/miekg/dns"
	ipv4 "golang.org/x/net/ipv4"
	ipv6 "golang.org/x/net/ipv6"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// interfaceForName returns a net.Interface or error
func interfaceForName(name string) (net.Interface, error) {
	if name == "" {
		return net.Interface{}, nil
	}
	ifaces, err := net.Interfaces()
	if err != nil {
		return net.Interface{}, err
	}
	for _, iface := range ifaces {
		if iface.Name == name {
			return iface, nil
		}
	}
	return net.Interface{}, ErrBadParameter.With(name)
}

// multicastInterfaces returns one or more interfaces which should be bound
// for listening
func multicastInterfaces(iface net.Interface) ([]net.Interface, error) {
	if iface.Name != "" {
		if (iface.Flags&net.FlagUp) > 0 && (iface.Flags&net.FlagMulticast) > 0 {
			return []net.Interface{iface}, nil
		} else {
			return nil, fmt.Errorf("Interface %v is not up and/or multicast-enabled", iface.Name)
		}
	}
	if ifaces, err := net.Interfaces(); err != nil {
		return nil, err
	} else {
		interfaces := make([]net.Interface, 0, len(ifaces))
		for _, ifi := range ifaces {
			if (ifi.Flags & net.FlagUp) == 0 {
				continue
			}
			if (ifi.Flags & net.FlagMulticast) == 0 {
				continue
			}
			if (ifi.Flags & net.FlagPointToPoint) != 0 {
				continue
			}
			if addrs, err := ifi.MulticastAddrs(); err != nil || len(addrs) == 0 {
				continue
			}
			interfaces = append(interfaces, ifi)
		}
		if len(interfaces) > 0 {
			return interfaces, nil
		} else {
			return nil, fmt.Errorf("No multicast-enabled interface found")
		}
	}
}

// bindUdp4 binds to listen on a particular address for IPv4
func bindUdp4(ifaces []net.Interface, addr *net.UDPAddr) (*ipv4.PacketConn, error) {
	var result error

	if len(ifaces) == 0 {
		return nil, ErrBadParameter
	} else if conn, err := net.ListenUDP("udp4", addr); err != nil {
		return nil, err
	} else if packet_conn := ipv4.NewPacketConn(conn); packet_conn == nil {
		return nil, conn.Close()
	} else {
		packet_conn.SetControlMessage(ipv4.FlagInterface, true)
		for _, iface := range ifaces {
			if err := packet_conn.JoinGroup(&iface, &net.UDPAddr{IP: addr.IP}); err != nil {
				if err_, ok := err.(*os.SyscallError); ok && err_.Err == syscall.EAFNOSUPPORT {
					continue
				} else {
					result = multierror.Append(result, fmt.Errorf("%v: %w", iface.Name, err))
				}
			}
		}
		if result != nil {
			if err := conn.Close(); err != nil {
				result = multierror.Append(result, err)
			}
		}
		return packet_conn, result
	}
}

// bindUdp6 binds to listen on a particular address for IPv6
func bindUdp6(ifaces []net.Interface, addr *net.UDPAddr) (*ipv6.PacketConn, error) {
	var result error

	if len(ifaces) == 0 {
		return nil, ErrBadParameter
	} else if conn, err := net.ListenUDP("udp6", addr); err != nil {
		return nil, err
	} else if packet_conn := ipv6.NewPacketConn(conn); packet_conn == nil {
		return nil, conn.Close()
	} else {
		packet_conn.SetControlMessage(ipv6.FlagInterface, true)
		for _, iface := range ifaces {
			if err := packet_conn.JoinGroup(&iface, &net.UDPAddr{IP: addr.IP}); err != nil {
				if err_, ok := err.(*os.SyscallError); ok && err_.Err == syscall.EAFNOSUPPORT {
					continue
				} else {
					result = multierror.Append(result, fmt.Errorf("%v: %w", iface.Name, err))
				}
			}
		}
		if result != nil {
			if err := conn.Close(); err != nil {
				result = multierror.Append(result, err)
			}
		}
		return packet_conn, result
	}
}

// parseDnsPacket parses packets into DNS message
func parseDnsPacket(packet []byte) (*dns.Msg, error) {
	var msg dns.Msg
	if err := msg.Unpack(packet); err != nil {
		return nil, err
	}
	if msg.Opcode != dns.OpcodeQuery {
		return nil, fmt.Errorf("Query with invalid Opcode %v (expected %v)", msg.Opcode, dns.OpcodeQuery)
	}
	if msg.Rcode != 0 {
		return nil, fmt.Errorf("Query with non-zero Rcode %v", msg.Rcode)
	}
	if msg.Truncated {
		return nil, fmt.Errorf("Support for DNS requests with high truncated bit not implemented")
	}
	return &msg, nil
}

// Return fully-qualified value
func fqn(value string) string {
	return strings.Trim(value, ".") + "."
}

// Transform from fully-qualified value
func unfqn(value string) string {
	return strings.TrimSuffix(value, ".")
}

// Return a query message looking up all services
func msgQueryServices(zone string) *dns.Msg {
	msg := new(dns.Msg)
	msg.SetQuestion(fqn(ServicesQuery)+fqn(zone), dns.TypePTR)
	msg.RecursionDesired = false
	return msg
}

// Return a query message looking up a specific service record
func msgQueryInstances(srv, zone string) *dns.Msg {
	msg := new(dns.Msg)
	msg.SetQuestion(fqn(srv)+fqn(zone), dns.TypePTR)
	msg.RecursionDesired = false
	return msg
}
