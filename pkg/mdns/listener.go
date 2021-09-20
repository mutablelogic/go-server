package mdns

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	// Packages
	multierror "github.com/hashicorp/go-multierror"
	dns "github.com/miekg/dns"
	ipv4 "golang.org/x/net/ipv4"
	ipv6 "golang.org/x/net/ipv6"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type listener struct {
	sync.WaitGroup

	// Domain for mDNS
	domain string

	// Interfaces for listener
	ifaces []net.Interface

	// Bound listeners
	ip4 *ipv4.PacketConn
	ip6 *ipv6.PacketConn

	// Channel to send messages to
	c chan<- message
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	emitRetryCount    = 3
	emitRetryDuration = 100 * time.Millisecond
	defaultDomain     = "local."
)

var (
	MULTICAST_ADDR_IPV4 = &net.UDPAddr{IP: net.ParseIP("224.0.0.251"), Port: 5353}
	MULTICAST_ADDR_IPV6 = &net.UDPAddr{IP: net.ParseIP("ff02::fb"), Port: 5353}
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewListener(domain, iface string, c chan<- message) (*listener, error) {
	this := new(listener)

	// Fully qualify domain (remove dots and add one to end)
	if domain = strings.Trim(domain, ".") + "."; domain == "." {
		this.domain = defaultDomain
	} else {
		this.domain = domain
	}

	// Obtain the interfaces for listening
	if iface, err := interfaceForName(iface); err != nil {
		return nil, err
	} else if ifaces, err := multicastInterfaces(iface); err != nil {
		return nil, err
	} else if len(ifaces) == 0 {
		return nil, ErrBadParameter.With("No interfaces defined for listening")
	} else {
		this.ifaces = ifaces
	}

	// Join IP4
	if ip4, err := bindUdp4(this.ifaces, MULTICAST_ADDR_IPV4); err != nil {
		return nil, err
	} else {
		this.ip4 = ip4
	}

	// Join IP6
	if ip6, err := bindUdp6(this.ifaces, MULTICAST_ADDR_IPV6); err != nil {
		return nil, err
	} else {
		this.ip6 = ip6
	}

	// Set channel
	this.c = c

	// Return success
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *listener) String() string {
	str := "<listener"
	str += fmt.Sprintf(" domain=%q", this.domain)
	str += " ifaces="
	for i, iface := range this.ifaces {
		if i > 0 {
			str += ","
		}
		str += iface.Name
	}

	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *listener) Run(ctx context.Context) error {
	var result error

	// Run4
	if this.ip4 != nil {
		this.WaitGroup.Add(1)
		go this.run4(ctx, this.ip4)
	}

	// Run6
	if this.ip6 != nil {
		this.WaitGroup.Add(1)
		go this.run6(ctx, this.ip6)
	}

	// Wait for cancels
	<-ctx.Done()

	// Close connections
	if this.ip4 != nil {
		if err := this.ip4.Close(); err != nil {
			result = multierror.Append(result, err)
		}
	}
	if this.ip6 != nil {
		if err := this.ip6.Close(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Wait until receive loops have completed
	this.WaitGroup.Wait()

	// Return any errors
	return result
}

// Send a single DNS message to a particular interface or all interfaces if 0
func (this *listener) Send(msg *dns.Msg, ifIndex int) error {
	var buf []byte
	var result error

	if buf_, err := msg.Pack(); err != nil {
		return err
	} else {
		buf = buf_
	}

	/*for i, q := range msg.Question {
		this.Debug("  ", i, " Send: ", q.Name, " type=", qTypeString(q.Qtype), " ifIndex=", ifIndex)
	}
	*/
	if this.ip4 != nil {
		var cm ipv4.ControlMessage
		if ifIndex != 0 {
			cm.IfIndex = ifIndex
			if _, err := this.ip4.WriteTo(buf, &cm, MULTICAST_ADDR_IPV4); err != nil {
				result = multierror.Append(result, err)
			}
		} else {
			for _, intf := range this.ifaces {
				cm.IfIndex = intf.Index
				if intf.Flags&net.FlagUp != 0 {
					this.ip4.WriteTo(buf, &cm, MULTICAST_ADDR_IPV4)
				}
			}
		}
	}

	if this.ip6 != nil {
		var cm ipv6.ControlMessage
		if ifIndex != 0 {
			cm.IfIndex = ifIndex
			if _, err := this.ip6.WriteTo(buf, &cm, MULTICAST_ADDR_IPV6); err != nil {
				result = multierror.Append(result, err)
			}
		} else {
			for _, intf := range this.ifaces {
				cm.IfIndex = intf.Index
				if intf.Flags&net.FlagUp != 0 {
					this.ip6.WriteTo(buf, &cm, MULTICAST_ADDR_IPV6)
				}
			}
		}
	}

	// Return any errors
	return result
}

// Query by sending multiple messages to interfaces
func (this *listener) Query(ctx context.Context, msg *dns.Msg, iface int) error {
	timer := time.NewTimer(1 * time.Nanosecond)
	defer timer.Stop()
	c := 0
	for {
		c++
		select {
		case <-timer.C:
			if err := this.Send(msg, iface); err != nil {
				return err
			} else if c >= emitRetryCount {
				return nil
			}
			timer.Reset(emitRetryDuration * time.Duration(c))
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *listener) run4(ctx context.Context, conn *ipv4.PacketConn) {
	defer this.WaitGroup.Done()

	buf := make([]byte, 65536)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if n, cm, from, err := conn.ReadFrom(buf); err != nil {
				continue
			} else if cm == nil {
				continue
			} else if msg, err := parseDnsPacket(buf[:n]); err != nil {
				this.c <- message{Err: err}
			} else {
				this.c <- message{msg, from, cm.IfIndex, this.domain, nil}
			}
		}
	}
}

func (this *listener) run6(ctx context.Context, conn *ipv6.PacketConn) {
	defer this.WaitGroup.Done()

	buf := make([]byte, 65536)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if n, cm, from, err := conn.ReadFrom(buf); err != nil {
				continue
			} else if cm == nil {
				continue
			} else if msg, err := parseDnsPacket(buf[:n]); err != nil {
				this.c <- message{Err: err}
			} else {
				this.c <- message{msg, from, cm.IfIndex, this.domain, nil}
			}
		}
	}
}
