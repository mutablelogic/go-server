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
	types "github.com/mutablelogic/go-server/pkg/types"
	ipv4 "golang.org/x/net/ipv4"
	ipv6 "golang.org/x/net/ipv6"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type mdns struct {
	task.Task
	Services

	ifaces map[int]net.Interface // Interface parameters
	label  string
	zone   string        // mDNS zone (usually "local")
	count  int           // Send parameters
	delta  time.Duration // Send parameters
	send   chan Message  // Send channel
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new logger task with provider of other tasks
func NewWithPlugin(p Plugin, label string) (*mdns, error) {
	mdns := new(mdns)
	mdns.count = sendRetryCount
	mdns.delta = sendRetryDelta
	mdns.zone = p.Domain()
	mdns.label = label

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
	if mdns.label != "" {
		str += fmt.Sprintf(" label=%q", mdns.label)
	}
	if mdns.zone != "" {
		str += fmt.Sprintf(" zone=%q", mdns.zone)
	}
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

	// Channel for receiving messages
	ch := make(chan Message, 1)
	defer close(ch)

	// Channel for sending messages
	mdns.send = make(chan Message)
	defer close(mdns.send)

	// IP4
	if ip4, result = bindUdp4(ifaces, multicastAddrIp4); result != nil {
		return result
	} else {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mdns.run4(ctx, ip4, ch)
		}()
	}

	// IP6
	if ip6, result = bindUdp6(ifaces, multicastAddrIp6); result != nil {
		return result
	} else {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mdns.run6(ctx, ip6, ch)
		}()
	}

	// Timer for discovery
	discoverTimer := time.NewTimer(10 * time.Second)
	defer discoverTimer.Stop()

	// Timer for expiring instances
	expireTimer := time.NewTimer(30 * time.Second)
	defer expireTimer.Stop()

	// Timer for resolving service names to records
	resolveTimer := time.NewTimer(20 * time.Second)
	defer resolveTimer.Stop()

	// Receive messages until done
FOR_LOOP:
	for {
		select {
		case <-ctx.Done():
			// Loop should quit
			break FOR_LOOP
		case <-discoverTimer.C:
			// Discover service names in background. It is done every ten minutes.
			// We run this in the background so it doesn't break the run loop
			go func() {
				child, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()
				if _, err := mdns.Discover(child); err != nil {
					mdns.Emit(event.Error(ctx, err))
				}
			}()
			// Repeat again in 10 minutes
			discoverTimer.Reset(discoverDelta)
		case <-expireTimer.C:
			// Expire instances which have reached their TTL
			for _, instance := range mdns.ExpireInstances() {
				mdns.Emit(event.New(ctx, Expired, instance))
			}
			// Repeat again every minute
			expireTimer.Reset(expireDelta)
		case <-resolveTimer.C:
			// Resolve service instances in background. This comes in two parts:
			// Browse will return the PTR, A and AAAA records for the service
			// and Lookup will return the SRV and TXT records. In terms of timing,
			// Discover is run no more than once every 10 minutes per service name
			// to find new instances.
			// We run this in the background so it doesn't break the run loop
			go func() {
				child, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()
				if service := mdns.NextBrowseName(discoverDelta); service != "" {
					if err := mdns.Browse(child, service); err != nil {
						mdns.Emit(event.Error(ctx, err))
					}
				} else if instance := mdns.NextLookupInstance(); instance != nil {
					if _, err := mdns.Lookup(child, types.DomainFqn(instance.Name, instance.Service), instance.IfIndex); err != nil {
						mdns.Emit(event.Error(ctx, err))
					}
				}
			}()
			// Repeat again in 10 seconds
			resolveTimer.Reset(resolveDelta)
		case msg := <-ch:
			// When a message is received it is parsed and the appropriate action is taken
			// in the background
			go mdns.parse(ctx, msg)
		case msg := <-mdns.send:
			// Message should be sent
			if msg != nil {
				if err := mdns.tx(ip4, ip6, msg); err != nil {
					mdns.Emit(event.Error(ctx, err))
				} else {
					mdns.Emit(event.New(ctx, Sent, msg))
				}
			}
		}
	}

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

// Multicast send a message
func (mdns *mdns) Send(ctx context.Context, msg Message) error {
	timer := time.NewTimer(time.Millisecond)
	defer timer.Stop()

	i := 0
	for {
		i++
		select {
		case <-timer.C:
			mdns.send <- msg
			if i > mdns.count {
				return nil
			}
			timer.Reset(mdns.delta * time.Duration(i))
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// Discover services
func (mdns *mdns) Discover(ctx context.Context) ([]string, error) {
	// Create a message to send - without defined interface, so it broadcasts on all interfaces
	if message, err := NewDiscoverQuestion(types.DomainFqn(servicesQuery, mdns.zone), net.Interface{}); err != nil {
		return nil, err
	} else if err := mdns.Send(ctx, message); err != nil {
		return nil, err
	}

	// Return all service names
	return mdns.GetNames(), nil
}

// Browse for instances of a service
func (mdns *mdns) Browse(ctx context.Context, service string) error {
	// Create a message to send - without defined interface, so it broadcasts on all interfaces
	if message, err := NewBrowseQuestion(types.DomainFqn(service, mdns.zone), net.Interface{}); err != nil {
		return err
	} else if err := mdns.Send(ctx, message); err != nil {
		return err
	}

	// Return all service names
	return nil
}

// Lookup a specific service by its name and type, and return the service instances
func (mdns *mdns) Lookup(ctx context.Context, name string, iface int) ([]Service, error) {
	// Create a message to send
	if message, err := NewLookupQuestion(types.DomainFqn(name, mdns.zone), mdns.interfaceWithIndex(iface)); err != nil {
		return nil, err
	} else if err := mdns.Send(ctx, message); err != nil {
		return nil, err
	}

	// Return success
	return nil, nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (mdns *mdns) parse(ctx context.Context, msg Message) {
	// parse each PTR record
	for _, ptr := range msg.PTR() {
		service := types.DomainInZone(ptr.Service(), mdns.zone)
		switch service {
		case "":
			// Not a service we are interested in
			continue
		case servicesQuery:
			// Add into the database
			if name := types.DomainInZone(ptr.Name(), mdns.zone); name != "" {
				if mdns.Registered(name, ptr.TTL()) {
					// Registration of a service name
					mdns.Emit(event.New(ctx, Registered, ptr))
				}
			}
		default:
			key := ptr.Name()
			if ptr.TTL() == 0 {
				// Remove from the database
				if mdns.Expired(key) {
					// Expiry of a service
					mdns.Emit(event.New(ctx, Expired, ptr))
				}
			} else {
				// Add into the database
				if mdns.AddPTR(key, service, ptr, msg.IfIndex()) {
					// Resolution of a service
					mdns.Emit(event.New(ctx, Resolved, ptr))
				}
				// Add A and AAAA records
				mdns.AddA(key, msg.A())
			}
		}
	}
	// Add TXT records
	if key, txt := msg.TXT(); key != "" {
		mdns.AddTXT(key, txt)
	}
	// Add SRV records
	if key, srv := msg.SRV(); key != "" {
		mdns.AddSRV(key, srv)
	}
}

func (mdns *mdns) interfaceWithIndex(ifIndex int) net.Interface {
	if iface, exists := mdns.ifaces[ifIndex]; exists {
		return iface
	} else {
		return net.Interface{}
	}
}

func (mdns *mdns) run4(ctx context.Context, conn *ipv4.PacketConn, ch chan<- Message) {
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
			} else if msg.IsAnswer() {
				mdns.Emit(event.New(ctx, Answer, msg))
				ch <- msg
			} else {
				mdns.Emit(event.New(ctx, Question, msg))
			}
		}
	}
}

func (mdns *mdns) run6(ctx context.Context, conn *ipv6.PacketConn, ch chan<- Message) {
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
			} else if msg.IsAnswer() {
				mdns.Emit(event.New(ctx, Answer, msg))
				ch <- msg
			} else {
				mdns.Emit(event.New(ctx, Question, msg))
			}
		}
	}
}

// Transmit a single DNS message to a particular interface or all interfaces if message.IfIndex is 0
func (mdns *mdns) tx(conn4 *ipv4.PacketConn, conn6 *ipv6.PacketConn, message Message) error {
	var result error

	// Pack the message
	ifIndex := message.IfIndex()
	data, err := message.Bytes()
	if err != nil {
		return err
	}

	// IP4 send
	if conn4 != nil {
		var cm ipv4.ControlMessage
		if ifIndex != 0 {
			cm.IfIndex = ifIndex
			if _, err := conn4.WriteTo(data, &cm, multicastAddrIp4); err != nil {
				result = multierror.Append(result, err)
			}
		} else {
			for _, intf := range mdns.ifaces {
				cm.IfIndex = intf.Index
				if _, err := conn4.WriteTo(data, &cm, multicastAddrIp4); err != nil {
					result = multierror.Append(result, err)
				}
			}
		}
	}

	// IP6 send
	if conn6 != nil {
		var cm ipv6.ControlMessage
		if ifIndex != 0 {
			cm.IfIndex = ifIndex
			if _, err := conn6.WriteTo(data, &cm, multicastAddrIp6); err != nil {
				result = multierror.Append(result, err)
			}
		} else {
			for _, intf := range mdns.ifaces {
				cm.IfIndex = intf.Index
				if _, err := conn6.WriteTo(data, &cm, multicastAddrIp6); err != nil {
					result = multierror.Append(result, err)
				}
			}
		}
	}

	// Return any errors
	return result
}
