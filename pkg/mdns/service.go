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

type Service struct {
	service string
	zone    string
	name    string
	host    string
	port    uint16
	a       []net.IP
	aaaa    []net.IP
	txt     []string
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

func NewService(zone string) *Service {
	this := new(Service)
	this.zone = zone
	return this
}

///////////////////////////////////////////////////////////////////////////////
// GET PROPERTIES

func (this *Service) Instance() string {
	return strings.TrimSuffix(fqn(this.name), this.zone)
}

func (this *Service) Service() string {
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

func (this *Service) Name() string {
	name := this.name
	if srv := this.Service(); srv != fqn(ServicesQuery) && name != "" {
		name = strings.TrimSuffix(fqn(name), fqn(this.service))
		if name_, err := Unquote(name); err == nil {
			name = name_
		}
	}
	return unfqn(name)
}

func (this *Service) Host() string {
	return this.host
}

func (this *Service) Port() uint16 {
	return this.port
}

func (this *Service) Zone() string {
	return fqn(this.zone)
}

func (this *Service) Addrs() []net.IP {
	addrs := []net.IP{}
	addrs = append(addrs, this.a...)
	addrs = append(addrs, this.aaaa...)
	return addrs
}

func (this *Service) Txt() []string {
	return this.txt
}

///////////////////////////////////////////////////////////////////////////////
// SET PROPERTIES

func (this *Service) SetPTR(ptr *dns.PTR) {
	this.service = ptr.Hdr.Name
	this.name = ptr.Ptr
	this.ttl = time.Duration(ptr.Hdr.Ttl) * time.Second
}

func (this *Service) SetSRV(host string, port uint16, priority uint16) {
	this.host = host
	this.port = port
}

func (this *Service) SetTXT(txt []string) {
	this.txt = txt
}

func (this *Service) SetA(ip net.IP) {
	this.a = append(this.a, ip)
}

func (this *Service) SetAAAA(ip net.IP) {
	this.aaaa = append(this.aaaa, ip)
}

///////////////////////////////////////////////////////////////////////////////
// CHECK FOR EQUALITY

func (this *Service) Equals(other *Service) bool {
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

func (s Service) String() string {
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
