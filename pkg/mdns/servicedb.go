package mdns

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	// Modules
	. "github.com/djthorpe/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type servicedb struct {
	Title    string        `xml:"title"`
	Category string        `xml:"category"`
	Updated  string        `xml:"updated"`
	Records  []*namerecord `xml:"record"`

	// Lookup table
	lut map[string]*namerecord
}

type NameRecord interface {
	Name() string
	Service() string
	Protocol() string
	Description() string
	Note() string
}

type namerecord struct {
	Name_        string `xml:"name"`
	Description_ string `xml:"description"`
	Protocol_    string `xml:"protocol"`
	Note_        string `xml:"note"`
	Port_        string `xml:"number"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	DefaultServiceDatabase = "https://www.iana.org/assignments/service-names-port-numbers/service-names-port-numbers.xml"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewServiceDatabase() *servicedb {
	this := new(servicedb)
	this.lut = make(map[string]*namerecord)
	return this
}

func ReadServiceDatabase(v string) (*servicedb, error) {
	this := NewServiceDatabase()
	if url, err := url.Parse(v); err != nil {
		return nil, err
	} else if url.Scheme == "file" {
		r, err := os.Open(url.Path)
		if err != nil {
			return nil, err
		}
		defer r.Close()

		if err := this.read(r); err != nil {
			return nil, err
		}
	} else if url.Scheme == "http" || url.Scheme == "https" {
		r, err := http.Get(url.String())
		if err != nil {
			return nil, err
		}
		defer r.Body.Close()
		if r.StatusCode != http.StatusOK {
			return nil, ErrUnexpectedResponse.With(r.Status)
		}
		if err := this.read(r.Body); err != nil {
			return nil, err
		}
	} else {
		return nil, ErrBadParameter.With("Unsupported scheme: ", url.Scheme)
	}

	// Return success
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Lookup a record
func (this *servicedb) Lookup(service string) NameRecord {
	service = fqn(strings.ToLower(service))
	return this.lut[service]
}

// Name returns the service name
func (r *namerecord) Name() string {
	return r.Name_
}

func (r *namerecord) Description() string {
	return r.Description_
}

func (r *namerecord) Protocol() string {
	return r.Protocol_
}

func (r *namerecord) Note() string {
	return r.Note_
}

// Service returns the service name or empty string if not found
func (r *namerecord) Service() string {
	if r.Port_ == "" || r.Protocol_ == "" || r.Name_ == "" {
		return ""
	}
	return strings.ToLower("_" + r.Name_ + "._" + r.Protocol_ + ".")
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *servicedb) String() string {
	str := "<service-database"
	for k, v := range this.lut {
		str += fmt.Sprintf(" %v => %q", k, v.Description())
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *servicedb) read(r io.Reader) error {
	dec := xml.NewDecoder(r)
	if err := dec.Decode(this); err != nil {
		return err
	}

	// Add lookup table
	for _, record := range this.Records {
		if service := record.Service(); service != "" {
			this.lut[service] = record
		}
	}

	// Return success
	return nil
}
