package mdns

import (
	"fmt"
	"net"
	"time"

	// Modules
	dns "github.com/miekg/dns"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type message struct {
	msg    dns.Msg
	sender net.Addr
	iface  net.Interface
}

type ptr struct {
	service string
	name    string
	ttl     time.Duration
}

type srv struct {
	Host_     string `json:"host,omitempty"`
	Port_     uint16 `json:"port,omitempty"`
	Priority_ uint16 `json:"priority,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// MESSAGE INTERFACE

// A sent or received message
type Message interface {
	PTR() []Ptr              // Return PTR records
	A() []net.IP             // Return A and AAAA records
	TXT() (string, []string) // Return TXT records
	SRV() (string, []Srv)    // Return SRV records
	Bytes() ([]byte, error)  // Return the packed message
	IsAnswer() bool          // Return true when the message is an answer
	IfIndex() int            // Interface to send and receive on, or nil
}

type Ptr interface {
	Service() string    // Return service
	Name() string       // Return name
	TTL() time.Duration // Return time-to-live for record
}

type Srv interface {
	Host() string     // Return host
	Port() uint16     // Return port
	Priority() uint16 // Return priority
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a message from a received packet, with a sender and the interface that
// the message was received on
func MessageFromPacket(packet []byte, sender net.Addr, iface net.Interface) (Message, error) {
	message := new(message)
	message.sender = sender
	message.iface = iface

	if packet == nil {
		return nil, ErrBadParameter.With("packet")
	}
	if err := message.msg.Unpack(packet); err != nil {
		return nil, ErrBadParameter.Withf("%v: %v", sender, err)
	}
	if message.msg.Opcode != dns.OpcodeQuery {
		return nil, ErrUnexpectedResponse.Withf("Query with invalid Opcode %v (expected %v)", message.msg.Opcode, dns.OpcodeQuery)
	}
	if message.msg.Rcode != 0 {
		return nil, ErrUnexpectedResponse.Withf("Query with non-zero Rcode %v", message.msg.Rcode)
	}
	if message.msg.Truncated {
		return nil, ErrUnexpectedResponse.Withf("Support for DNS requests with high truncated bit not implemented")
	}

	// Return success
	return message, nil
}

// Create a message from a question, to be sent on a specific interface or
// on all interfaces if the index is zero
func NewDiscoverQuestion(question string, iface net.Interface) (Message, error) {
	message := new(message)
	message.iface = iface
	message.msg.SetQuestion(question, dns.TypePTR)
	message.msg.RecursionDesired = false
	return message, nil
}

// Create a message from a question, to be sent on a specific interface or
// on all interfaces if the index is zero
func NewBrowseQuestion(question string, iface net.Interface) (Message, error) {
	message := new(message)
	message.iface = iface
	message.msg.Question = []dns.Question{
		{Name: question, Qtype: dns.TypePTR, Qclass: dns.ClassINET},
	}
	message.msg.RecursionDesired = false
	return message, nil
}

// Create a message from a question to retrieve TXT and SRV records, to be sent on a specific interface or
// on all interfaces if the index is zero
func NewLookupQuestion(question string, iface net.Interface) (Message, error) {
	message := new(message)
	message.iface = iface
	message.msg.Question = []dns.Question{
		{Name: question, Qtype: dns.TypeSRV, Qclass: dns.ClassINET},
		{Name: question, Qtype: dns.TypeTXT, Qclass: dns.ClassINET},
	}
	message.msg.RecursionDesired = false
	return message, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (m *message) String() string {
	str := "<message"
	if id := m.msg.Id; id != 0 {
		str += fmt.Sprintf(" id=%v", id)
	}
	for _, qs := range m.msg.Question {
		str += fmt.Sprintf(" q=%q", qs.Name)
	}
	if ptr := m.PTR(); len(ptr) > 0 {
		str += fmt.Sprintf(" ptr=%v", ptr)
	}
	if a := m.A(); len(a) > 0 {
		str += fmt.Sprintf(" a=%q", a)
	}
	if _, srv := m.SRV(); len(srv) > 0 {
		str += fmt.Sprintf(" srv=%v", srv)
	}
	if _, txt := m.TXT(); len(txt) > 0 {
		str += fmt.Sprintf(" txt=%q", txt)
	}
	if m.sender != nil {
		str += fmt.Sprintf(" sender=%v", m.sender)
	}
	if m.iface.Index != 0 {
		str += fmt.Sprintf(" iface=%q", m.iface.Name)
	}
	return str + ">"
}

func (p ptr) String() string {
	str := "<ptr"
	if p.service != "" {
		str += fmt.Sprintf(" service=%q", p.service)
	}
	if p.name != "" {
		str += fmt.Sprintf(" name=%q", p.name)
	}
	if p.ttl > 0 {
		str += fmt.Sprintf(" ttl=%v", p.ttl)
	} else if p.ttl == 0 {
		str += " expired"
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PTR: PUBLIC METHODS

func (p ptr) Name() string {
	return p.name
}

func (p ptr) Service() string {
	return p.service
}

func (p ptr) TTL() time.Duration {
	return p.ttl
}

///////////////////////////////////////////////////////////////////////////////
// SRV: PUBLIC METHODS

func (s srv) Host() string {
	return s.Host_
}

func (s srv) Port() uint16 {
	return s.Port_
}

func (s srv) Priority() uint16 {
	return s.Priority_
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Bytes returns the packed message
func (m *message) Bytes() ([]byte, error) {
	return m.msg.Pack()
}

// IsAnswer returns true when the message is an answer
func (m *message) IsAnswer() bool {
	return len(m.msg.Question) == 0 && len(m.msg.Answer) > 0
}

// IfIndex returns the index of the interface for receive or send
func (m *message) IfIndex() int {
	return m.iface.Index
}

// Return PTR records
func (m *message) PTR() []Ptr {
	var result []Ptr
	for _, rec := range m.msg.Answer {
		switch rr := rec.(type) {
		case *dns.PTR:
			result = append(result, ptr{rr.Hdr.Name, rr.Ptr, time.Duration(rr.Hdr.Ttl) * time.Second})
		}
	}
	return result
}

// Return A and AAAA records
func (m *message) A() []net.IP {
	var result []net.IP
	for _, rec := range m.msg.Answer {
		switch rr := rec.(type) {
		case *dns.A:
			result = append(result, rr.A)
		case *dns.AAAA:
			result = append(result, rr.AAAA)
		}
	}
	for _, rec := range m.msg.Extra {
		switch rr := rec.(type) {
		case *dns.A:
			result = append(result, rr.A)
		case *dns.AAAA:
			result = append(result, rr.AAAA)
		}
	}
	return result
}

// Return TXT records
func (m *message) TXT() (string, []string) {
	var key string
	var result []string
	for _, rec := range m.msg.Answer {
		switch rr := rec.(type) {
		case *dns.TXT:
			key = rr.Hdr.Name
			for _, txt := range rr.Txt {
				if txt != "" {
					result = append(result, txt)
				}
			}
		}
	}
	for _, rec := range m.msg.Extra {
		switch rr := rec.(type) {
		case *dns.TXT:
			key = rr.Hdr.Name
			for _, txt := range rr.Txt {
				if txt != "" {
					result = append(result, txt)
				}
			}
		}
	}
	return key, result
}

// Return SRV records
func (m *message) SRV() (string, []Srv) {
	var key string
	var result []Srv
	for _, rec := range m.msg.Answer {
		switch rr := rec.(type) {
		case *dns.SRV:
			key = rr.Hdr.Name
			result = append(result, srv{rr.Target, rr.Port, rr.Priority})
		}
	}
	return key, result
}
