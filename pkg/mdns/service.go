package mdns

import (
	"encoding/json"
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

func NewService(zone string) *Service {
	this := new(Service)
	this.zone = zone
	return this
}

///////////////////////////////////////////////////////////////////////////////
// GET PROPERTIES

// Instance returns the full instance name of the service
func (this *Service) Instance() string {
	return strings.TrimSuffix(fqn(this.name), this.zone)
}

// Service returns the short service identifier
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

// Name returns the name of the service
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

// Host returns the host of the service
func (this *Service) Host() string {
	return this.host
}

// Port returns the port of the service
func (this *Service) Port() uint16 {
	return this.port
}

// Zone returns the fully qualified zone (usually local.) of the service
func (this *Service) Zone() string {
	return fqn(this.zone)
}

// Addrs returns associated addresses for the service
func (this *Service) Addrs() []net.IP {
	addrs := []net.IP{}
	addrs = append(addrs, this.a...)
	addrs = append(addrs, this.aaaa...)
	return addrs
}

// Txt returns the txt records for the service
func (this *Service) Txt() []string {
	if len(this.txt) == 0 || (len(this.txt) == 1 && this.txt[0] == "") {
		return []string{}
	} else {
		return this.txt
	}
}

// Keys returns the keys in the TXT record
func (this *Service) Keys() []string {
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

// ValueForKey return value for a key in the TXT record, or
// empty string if not found
func (this *Service) ValueForKey(key string) string {
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
// PRIVATE METHODS

func (this *Service) setPTR(ptr *dns.PTR) {
	this.service = ptr.Hdr.Name
	this.name = ptr.Ptr
	this.ttl = time.Duration(ptr.Hdr.Ttl) * time.Second
}

func (this *Service) setSRV(host string, port uint16, priority uint16) {
	this.host = host
	this.port = port
}

func (this *Service) setTXT(txt []string) {
	this.txt = txt
}

func (this *Service) setA(ip net.IP) {
	this.a = append(this.a, ip)
}

func (this *Service) setAAAA(ip net.IP) {
	this.aaaa = append(this.aaaa, ip)
}

func (this *Service) txtMap() map[string]string {
	result := make(map[string]string, len(this.txt))
	for _, k := range this.Keys() {
		if v := this.ValueForKey(k); v != "" {
			result[k] = v
		}
	}
	return result
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (s *Service) MarshalJSON() ([]byte, error) {
	if s.Service() == fqn(ServicesQuery) {
		return json.Marshal(struct {
			Service string `json:"service,omitempty"`
			Zone    string `json:"zone,omitempty"`
		}{
			Service: s.Instance(),
			Zone:    s.Zone(),
		})
	} else {
		return json.Marshal(struct {
			Instance string            `json:"instance,omitempty"`
			Service  string            `json:"service,omitempty"`
			Name     string            `json:"name,omitempty"`
			Zone     string            `json:"zone,omitempty"`
			Host     string            `json:"host,omitempty"`
			Port     uint16            `json:"port,omitempty"`
			Addrs    []net.IP          `json:"addr,omitempty"`
			Txt      []string          `json:"txt,omitempty"`
			Map      map[string]string `json:"map,omitempty"`
		}{
			Instance: s.Instance(),
			Service:  s.Service(),
			Name:     s.Name(),
			Zone:     s.Zone(),
			Host:     s.Host(),
			Port:     s.Port(),
			Addrs:    s.Addrs(),
			Txt:      s.Txt(),
			Map:      s.txtMap(),
		})
	}
}
