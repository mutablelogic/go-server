package cert

import (
	"crypto/x509/pkix"
	"net"
	"strings"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Opt func(*opts) error

type opts struct {
	Name                *pkix.Name
	Serial              int64
	KeyType             keyType
	Years, Months, Days int
	IPAddresses         []net.IP
	DNSNames            []string
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func OptX509Name(v pkix.Name) Opt {
	return func(o *opts) error {
		o.Name = &v
		return nil
	}
}

func OptSerial(serial int64) Opt {
	return func(o *opts) error {
		o.Serial = serial
		return nil
	}
}

func OptKeyType(v string) Opt {
	return func(o *opts) error {
		switch strings.ToUpper(v) {
		case "ED25519":
			o.KeyType = ED25519
		case "RSA2048":
			o.KeyType = RSA2048
		case "P224":
			o.KeyType = P224
		case "P256":
			o.KeyType = P256
		case "P384":
			o.KeyType = P384
		case "P521":
			o.KeyType = P521
		default:
			return ErrBadParameter.Withf("invalid key type %q", v)
		}
		return nil
	}
}

func OptExpiry(years, months, days int) Opt {
	return func(o *opts) error {
		if years < 0 || months < 0 || days < 0 {
			return ErrBadParameter.Withf("OptExpiry")
		}
		// Maximum expiry is 4 years, 12 months, 31 days
		if years > 4 || months > 12 || days > 31 {
			return ErrBadParameter.Withf("OptExpiry")
		}
		o.Years = years
		o.Months = months
		o.Days = days
		return nil
	}
}

func OptHosts(v ...string) Opt {
	return func(o *opts) error {
		for _, v := range v {
			if ip := net.ParseIP(v); ip != nil {
				o.IPAddresses = append(o.IPAddresses, ip)
			} else {
				o.DNSNames = append(o.DNSNames, v)
			}
		}
		return nil
	}
}
