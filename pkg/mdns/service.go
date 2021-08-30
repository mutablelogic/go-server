package mdns

import (
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	// Modules
	dns "github.com/miekg/dns"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type service struct {
	service string
	zone    string
	name    string
	host    string
	port    uint16
	a       []net.IP
	aaaa    []net.IP
	txt     []string
	keys    []string
	ttl     time.Duration
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	reIsSubService  = regexp.MustCompile(`\._sub\.(.+)\.$`)
	reIsAddr4Lookup = regexp.MustCompile(`\.(in-addr\.arpa\.)$`)
	reIsAddr6Lookup = regexp.MustCompile(`\.(ip6\.arpa\.)$`)
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewService(zone string) *service {
	this := new(service)
	this.zone = zone
	return this
}

///////////////////////////////////////////////////////////////////////////////
// GET PROPERTIES

func (this *service) Instance() string {
	return strings.TrimSuffix(fqn(this.name), this.zone)
}

func (this *service) Service() string {
	service := strings.TrimSuffix(fqn(this.service), this.zone)
	if match := reIsSubService.FindStringSubmatch(service); match != nil {
		return match[1]
	} else if match := reIsAddr4Lookup.FindStringSubmatch(service); match != nil {
		return match[1]
	} else if match := reIsAddr6Lookup.FindStringSubmatch(service); match != nil {
		return match[1]
	} else {
		return service
	}
}

func (this *service) Name() string {
	name := this.name
	if srv := this.Service(); srv != fqn(ServicesQuery) && name != "" {
		name = strings.TrimSuffix(fqn(name), fqn(this.service))
		if name_, err := Unquote(name); err == nil {
			name = name_
		}
	}
	return unfqn(name)
}

func (this *service) Host() string {
	return this.host
}

func (this *service) Port() uint16 {
	return this.port
}

func (this *service) Zone() string {
	return fqn(this.zone)
}

func (this *service) Addrs() []net.IP {
	addrs := []net.IP{}
	addrs = append(addrs, this.a...)
	addrs = append(addrs, this.aaaa...)
	return addrs
}

func (this *service) Txt() []string {
	return this.txt
}

func (this *service) Keys() []string {
	if len(this.keys) == 0 {
		this.keys = make([]string, 0, len(this.txt))
		for _, value := range this.txt {
			if kv := strings.SplitN(value, "=", 2); len(kv) < 2 {
				continue
			} else if kv[0] == "" {
				continue
			} else {
				this.keys = append(this.keys, kv[0])
			}
		}
	}
	return this.keys
}

func (this *service) ValueForKey(key string) string {
	// Cache keys
	if len(this.keys) == 0 {
		this.Keys()
	}
	// Find key in TXT values
	key = key + "="
	for _, t := range this.txt {
		if strings.HasPrefix(t, key) {
			t = t[len(key):]
			if unquoted, err := Unquote(t); err != nil {
				return unquoted
			} else {
				return t
			}
		}
	}
	// Return empty string if not found
	return ""
}

///////////////////////////////////////////////////////////////////////////////
// SET PROPERTIES

func (this *service) SetPTR(ptr *dns.PTR) {
	this.service = ptr.Hdr.Name
	this.name = ptr.Ptr
	this.ttl = time.Duration(ptr.Hdr.Ttl) * time.Second
}

func (this *service) SetSRV(host string, port uint16, priority uint16) {
	this.host = host
	this.port = port
}

func (this *service) SetTXT(txt []string) {
	this.txt = txt
}

func (this *service) SetA(ip net.IP) {
	this.a = append(this.a, ip)
}

func (this *service) SetAAAA(ip net.IP) {
	this.aaaa = append(this.aaaa, ip)
}

///////////////////////////////////////////////////////////////////////////////
// CHECK FOR EQUALITY

func (this *service) Equals(other *service) bool {
	if this.Instance() != other.Instance() {
		fmt.Println("instance changed", this.Instance(), other.Instance())
		return false
	}
	if this.Host() != other.Host() {
		fmt.Println("host changed", this.Host(), other.Host())
		return false
	}
	if this.Port() != other.Port() {
		fmt.Println("port changed", this.Port(), other.Port())
		return false
	}
	if len(this.a) != len(other.a) {
		fmt.Println("A changed")
		return false
	}
	if len(this.aaaa) != len(other.aaaa) {
		fmt.Println("AAAA changed")
		return false
	}
	if len(this.txt) != len(other.txt) {
		fmt.Println("TXT changed")
		return false
	}
	// Now check a records
	for i := range this.a {
		if fmt.Sprint(this.a[i]) != fmt.Sprint(other.a[i]) {
			fmt.Println("A changed", this.a[i], other.a[i])
			return false
		}
	}
	for i := range this.aaaa {
		if fmt.Sprint(this.aaaa[i]) != fmt.Sprint(other.aaaa[i]) {
			fmt.Println("AAAA changed", this.aaaa[i], other.aaaa[i])
			return false
		}
	}
	for i := range this.txt {
		if this.txt[i] != other.txt[i] {
			fmt.Println("TXT changed", this.txt[i], other.txt[i])
			return false
		}
	}

	// Equals
	return true
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (s service) String() string {
	str := "<service"
	if instance := s.Instance(); instance != "" {
		str += fmt.Sprintf(" instance=%q", instance)
	}
	if service := s.Service(); service != "" {
		str += fmt.Sprintf(" service=%q", service)
	}
	if name := s.Name(); name != "" {
		str += fmt.Sprintf(" name=%q", name)
	}
	if zone := s.Zone(); zone != "" {
		str += fmt.Sprintf(" zone=%q", zone)
	}
	if host, port := s.Host(), s.Port(); host != "" {
		str += fmt.Sprintf(" host=%v", net.JoinHostPort(host, fmt.Sprint(port)))
	}
	if ips := s.Addrs(); len(ips) > 0 {
		str += fmt.Sprintf(" addrs=%v", ips)
	}
	if txt := s.Txt(); len(s.txt) > 0 {
		str += fmt.Sprintf(" txt=%q", txt)
	}
	if s.ttl != 0 {
		str += fmt.Sprintf(" ttl=%v", s.ttl)
	}
	return str + ">"
}
