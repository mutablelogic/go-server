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
	reIsSubService = regexp.MustCompile(`\._sub\.(.+)\.$`)
	reIsAddrLookup = regexp.MustCompile(`\.?(in-addr.arpa.)$`)
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
	} else if match := reIsAddrLookup.FindStringSubmatch(service); match != nil {
		return match[1]
	} else {
		return service
	}
}

func (this *Service) Name() string {
	name := strings.TrimSuffix(this.name, this.zone)
	if this.Service() != fqn(ServicesQuery) {
		name = strings.TrimSuffix(this.name, this.service)
		if name_, err := Unquote(unfqn(name)); err != nil {
		} else {
			name = name_
		}
	}
	return name
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

func (this *Service) String() string {
	str := "<service"
	if instance := this.Instance(); instance != "" {
		str += fmt.Sprintf(" instance=%q", instance)
	}
	if service := this.Service(); service != "" {
		str += fmt.Sprintf(" service=%q", service)
	}
	if name := this.Name(); name != "" {
		str += fmt.Sprintf(" name=%q", name)
	}
	if zone := this.Zone(); zone != "" {
		str += fmt.Sprintf(" zone=%q", zone)
	}
	if host, port := this.Host(), this.Port(); host != "" {
		str += fmt.Sprintf(" host=%v", net.JoinHostPort(host, fmt.Sprint(port)))
	}
	if ips := this.Addrs(); len(ips) > 0 {
		str += fmt.Sprintf(" addrs=%v", ips)
	}
	if txt := this.Txt(); len(this.txt) > 0 {
		str += fmt.Sprintf(" txt=%q", txt)
	}
	if this.ttl != 0 {
		str += fmt.Sprintf(" ttl=%v", this.ttl)
	}
	return str + ">"
}
