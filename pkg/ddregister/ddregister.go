package ddregister

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Configuration values
type Config struct {
	Timeout time.Duration `yaml:"timeout"`
}

// Concrete implementation
type Registration struct {
	client        *http.Client
	apify, google time.Time
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	userAgent        = "github.com/mutablelogic/go-server/pkg/ddregister"
	apifyUrl         = "https://api.ipify.org"
	googleDomainsUrl = "https://domains.google.com/nic/update"
	defaultTimeout   = 10 * time.Second
	ratePeriod       = 30 * time.Second // Maximum of two requests per minute
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New() *Registration {
	return NewWithConfig(Config{Timeout: defaultTimeout})
}

func NewWithConfig(cfg Config) *Registration {
	dd := new(Registration)

	// Create HTTP client
	dd.client = &http.Client{
		Timeout: cfg.Timeout,
	}

	// Return success
	return dd
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (dd *Registration) String() string {
	str := "<ddregister"
	if dd.client.Timeout > 0 {
		str += fmt.Sprint(" timeout=", dd.client.Timeout)
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// GetExternalAddress returns the current external IP address as reported
// by https://api.ipify.org
func (dd *Registration) GetExternalAddress() (net.IP, error) {
	// Check rate limit
	if !dd.apify.IsZero() && time.Since(dd.apify) < ratePeriod {
		return nil, ErrOutOfOrder.With("Request rate limit exceeded")
	}

	// Make request
	req, err := req("GET", apifyUrl)
	if err != nil {
		return nil, err
	}
	body, content_type, err := dd.do(req)
	dd.apify = time.Now()
	if err != nil {
		return nil, err
	}

	// Check response
	if content_type != "text/plain" {
		return nil, ErrUnexpectedResponse.Withf("Unexpected content type: %q", content_type)
	} else if ip := net.ParseIP(string(body)); ip == nil {
		return nil, ErrUnexpectedResponse.Withf("%q", string(body))
	} else {
		return ip, nil
	}
}

// RegisterAddress registers address with Google DNS and returns the response
// https://support.google.com/domains/answer/6147083?hl=en
func (dd *Registration) RegisterAddress(host, user, password string, addr net.IP, offline bool) error {
	// Check rate limit
	if !dd.google.IsZero() && time.Since(dd.google) < ratePeriod {
		return ErrOutOfOrder.With("Request rate limit exceeded")
	}

	// Make a request object
	req, err := req("GET", googleDomainsUrl)
	if err != nil {
		return err
	}

	// Form a query
	values := req.URL.Query()
	values.Set("hostname", strings.Trim(host, "."))
	values.Set("myip", addr.String())
	values.Set("offline", boolToString(offline))
	req.URL.RawQuery = values.Encode()
	if user != "" {
		req.URL.User = url.UserPassword(user, password)
	}

	// Make a request
	body, _, err := dd.do(req)
	dd.google = time.Now()
	if err != nil {
		return err
	}

	// Decode the response - return nil, ErrNotModified or ErrUnexpectedResponse
	status := strings.TrimSpace(string(body))
	switch {
	case strings.HasPrefix(status, "good"):
		return nil
	case strings.HasPrefix(status, "nochg"):
		return ErrNotModified.With(status)
	default:
		return ErrUnexpectedResponse.With(status)
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func req(method, url string) (*http.Request, error) {
	if req, err := http.NewRequest(method, url, nil); err != nil {
		return nil, err
	} else {
		req.Header.Add("User-Agent", userAgent)
		return req, nil
	}
}

func boolToString(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

func (dd *Registration) do(req *http.Request) ([]byte, string, error) {
	if resp, err := dd.client.Do(req); err != nil {
		return nil, "", err
	} else {
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return nil, "", ErrUnexpectedResponse.Withf("%q", resp.Status)
		} else if body, err := ioutil.ReadAll(resp.Body); err != nil {
			return nil, "", err
		} else {
			return body, resp.Header.Get("Content-Type"), nil
		}
	}
}
