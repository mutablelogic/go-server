package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"sync"
	"time"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	sync.Mutex
	*http.Client

	endpoint *url.URL
	ua       string
	rate     float32 // number of requests allowed per second
	strict   bool
	ts       time.Time
}

type ClientOpt func(*Client) error
type RequestOpt func(*http.Request) error

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	DefaultTimeout   = time.Second * 10
	DefaultUserAgent = "github.com/mutableloigc/go-server"
	PathSeparator    = string(os.PathSeparator)
	ContentTypeJson  = "application/json"
	ContentTypeText  = "text/plain"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New creates a new client with options. OptEndpoint is required as an option
// to set the endpoint for all requests.
func New(opts ...ClientOpt) (*Client, error) {
	this := new(Client)

	// Create a HTTP client
	this.Client = &http.Client{
		Timeout:   DefaultTimeout,
		Transport: http.DefaultTransport,
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(this); err != nil {
			return nil, err
		}
	}

	// If no endpoint, then return error
	if this.endpoint == nil {
		return nil, ErrBadParameter.With("missing endppint")
	}

	// Return success
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (client *Client) String() string {
	str := "<client"
	if client.endpoint != nil {
		str += fmt.Sprintf(" endpoint=%q", redactedUrl(client.endpoint))
	}
	if client.Client.Timeout > 0 {
		str += fmt.Sprint(" timeout=", client.Client.Timeout)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Do a JSON request with a payload, populate an object with the response
// and return any errors
func (client *Client) Do(in Payload, out any, opts ...RequestOpt) error {
	client.Mutex.Lock()
	defer client.Mutex.Unlock()

	// Check rate limit
	now := time.Now()
	if !client.ts.IsZero() && client.rate > 0.0 {
		next := client.ts.Add(time.Duration(float32(time.Second) / client.rate))
		if next.After(now) {
			time.Sleep(next.Sub(now))
		}
	}

	// Set timestamp at return
	defer func(now time.Time) {
		client.ts = now
	}(now)

	// Make a request
	var data []byte
	var err error
	if in != nil {
		if data, err = json.Marshal(in); err != nil {
			return err
		}
	}
	req, err := client.request(in.Method(), in.Accept(), bytes.NewReader(data))
	if err != nil {
		return err
	}

	if debug, ok := client.Client.Transport.(*logtransport); ok {
		debug.Payload(in)
	}

	return do(client.Client, req, in.Accept(), client.strict, out, opts...)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// request creates a request which can be used to return responses. The accept
// parameter is the accepted mime-type of the response. If the accept parameter is empty,
// then the default is application/json.
func (client *Client) request(method, accept string, body io.Reader) (*http.Request, error) {
	// Return error if no endpoint is set
	if client.endpoint == nil {
		return nil, ErrBadParameter.With("missing endpoint")
	}

	// Make a request
	r, err := http.NewRequest(method, client.endpoint.String(), body)
	if err != nil {
		return nil, err
	}

	// Set the credentials and user agent
	r.Header.Set("Content-Type", ContentTypeJson)
	if accept != "" {
		r.Header.Set("Accept", accept)
	}
	if client.ua != "" {
		r.Header.Set("User-Agent", client.ua)
	}

	// Return success
	return r, nil
}

// Do will make a JSON request, populate an object with the response and return any errors
func do(client *http.Client, req *http.Request, accept string, strict bool, out any, opts ...RequestOpt) error {
	// Apply request options
	for _, opt := range opts {
		if err := opt(req); err != nil {
			return err
		}
	}

	// Do the request
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// Decode body - this can be an array or an object, so we read the whole body
	// and choose happy and sad paths
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	// Check status code
	if response.StatusCode < 200 || response.StatusCode > 299 {
		// Read any information from the body
		var err string
		if err := decodeString(&err, string(data)); err != nil {
			return err
		}
		return ErrUnexpectedResponse.With(response.Status, ": ", err)
	}

	// When in strict mode, check content type returned is as expected
	if strict && accept != "" {
		contenttype := response.Header.Get("Content-Type")
		if mimetype, _, err := mime.ParseMediaType(contenttype); err != nil {
			return ErrUnexpectedResponse.With(contenttype)
		} else if mimetype != accept {
			return ErrUnexpectedResponse.With(contenttype)
		}
	}

	// Return success if out is nil
	if out == nil {
		return nil
	}

	// If JSON, then decode body
	if accept == ContentTypeJson {
		if err := json.NewDecoder(bytes.NewReader(data)).Decode(out); err == nil {
			return nil
		} else {
			return err
		}
	} else if accept == ContentTypeText {
		// Decode as text
		if err := decodeString(out, string(data)); err != nil {
			return err
		}
	}

	// Return success
	return nil
}

// Remove any usernames and passwords before printing out
func redactedUrl(url *url.URL) string {
	url_ := *url // make a copy
	url_.User = nil
	return url_.String()
}

// Set string from data
func decodeString(v interface{}, data string) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return ErrInternalAppError.With("DecodeString")
	} else {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.String {
		return ErrInternalAppError.With("DecodeString")
	}
	rv.Set(reflect.ValueOf(data))
	return nil
}
