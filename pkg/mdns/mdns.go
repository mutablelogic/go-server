package mdns

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	// Package imports
	multierror "github.com/hashicorp/go-multierror"
	event "github.com/mutablelogic/go-server/pkg/event"
	task "github.com/mutablelogic/go-server/pkg/task"
	ipv4 "golang.org/x/net/ipv4"
	ipv6 "golang.org/x/net/ipv6"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type mdns struct {
	task.Task

	ifaces map[int]net.Interface // Interface parameters
	count  int                   // Send parameters
	delta  time.Duration         // Send parameters
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new logger task with provider of other tasks
func NewWithPlugin(p Plugin) (*mdns, error) {
	mdns := new(mdns)
	mdns.count = sendRetryCount
	mdns.delta = sendRetryDelta

	// Enumerate interfaces to bind to
	interfaces, err := p.Interfaces()
	if err != nil {
		return nil, err
	}

	// Make index->interface map
	mdns.ifaces = make(map[int]net.Interface, len(interfaces))
	for _, iface := range interfaces {
		mdns.ifaces[iface.Index] = iface
	}

	return mdns, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (mdns *mdns) String() string {
	str := "<mdns"
	if ifaces := mdns.ifaces; len(ifaces) > 0 {
		var result []string
		for _, iface := range ifaces {
			result = append(result, iface.Name)
		}
		str += fmt.Sprintf(" interfaces=%q", result)
	}
	return str + ">"
}

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (mdns *mdns) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	var result error
	var ip4 *ipv4.PacketConn
	var ip6 *ipv6.PacketConn

	// Bind interfaces to sockets
	var ifaces []net.Interface
	for _, iface := range mdns.ifaces {
		ifaces = append(ifaces, iface)
	}

	// IP4
	if ip4, result = bindUdp4(ifaces, multicastAddrIp4); result != nil {
		return result
	} else {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mdns.run4(ctx, ip4)
		}()
	}

	// IP6
	if ip6, result = bindUdp6(ifaces, multicastAddrIp6); result != nil {
		return result
	} else {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mdns.run6(ctx, ip6)
		}()
	}

	// TOOD: Send and receive messages until done
	<-ctx.Done()

	// Close connections - these will quit tuen run4/run6 goroutines
	if err := ip4.Close(); err != nil {
		result = multierror.Append(result, err)
	}
	if err := ip6.Close(); err != nil {
		result = multierror.Append(result, err)
	}

	// Wait until receive loops have completed
	wg.Wait()

	// Return any errors
	return result
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (mdns *mdns) interfaceWithIndex(ifIndex int) net.Interface {
	if iface, exists := mdns.ifaces[ifIndex]; exists {
		return iface
	} else {
		return net.Interface{}
	}
}

func (mdns *mdns) run4(ctx context.Context, conn *ipv4.PacketConn) {
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
			} else if msg, err := MessageFromPacket(buf[:n], from, mdns.interfaceWithIndex(cm.IfIndex)); err != nil {
				mdns.Emit(event.Error(ctx, err))
			} else if msg.IsAnswer() && len(msg.PTR()) > 0 {
				mdns.Emit(event.New(ctx, Receive, msg))
			}
		}
	}
}

func (mdns *mdns) run6(ctx context.Context, conn *ipv6.PacketConn) {
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
			} else if msg, err := MessageFromPacket(buf[:n], from, mdns.interfaceWithIndex(cm.IfIndex)); err != nil {
				mdns.Emit(event.Error(ctx, err))
			} else if msg.IsAnswer() && len(msg.PTR()) > 0 {
				mdns.Emit(event.New(ctx, Receive, msg))
			}
		}
	}
}
