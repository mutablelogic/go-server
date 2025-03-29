package schema

import (
	"net"
	"time"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Certificate Metadata
type CertMeta struct {
	Name         string    `json:"name"`
	Signer       *string   `json:"signer,omitempty"`
	SerialNumber string    `json:"serial_number,omitempty"`
	Subject      string    `json:"subject,omitempty"`
	NotBefore    time.Time `json:"not_before,omitempty"`
	NotAfter     time.Time `json:"not_after,omitempty"`
	IPs          []net.IP  `json:"ips,omitempty"`
	Hosts        []string  `json:"hosts,omitempty"`
	IsCA         bool      `json:"is_ca,omitempty"`
	KeyType      string    `json:"key_type,omitempty"`
	KeyBits      string    `json:"key_subtype,omitempty"`
}
