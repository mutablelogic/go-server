package dnsregister

import (
	"errors"
	"fmt"

	// Packages
	client "github.com/mutablelogic/go-server/pkg/client"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	*client.Client
	Prefix string
}

type NomadOpt func(*Client) error

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultPrefix = "v1"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewNomadClient(client *client.Client, opts ...NomadOpt) (*Client, error) {
	this := new(Client)
	if client == nil {
		return nil, ErrBadParameter
	} else {
		this.Client = client
		this.Prefix = defaultPrefix
	}

	// Apply options
	var result error
	for _, opt := range opts {
		if err := opt(this); err != nil {
			result = errors.Join(result, err)
		}
	}

	// Return any errors
	return this, result
}

///////////////////////////////////////////////////////////////////////////////
// METHODS - Options

func OptPrefix(prefix string) NomadOpt {
	return func(c *Client) error {
		c.Prefix = prefix
		return nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// METHODS - Services

type NomadServicesResponse struct {
	Namespace string         `json:"Namespace"`
	Services  []NomadService `json:"Services"`
}

type NomadService struct {
	Name        string   `json:"ServiceName"`
	Namespace   string   `json:"Namespace"`
	Datacenter  string   `json:"Datacenter"`
	Address     string   `json:"Address,omitempty"`
	Port        uint64   `json:"Port,omitempty"`
	Id          string   `json:"ID,omitempty"`
	AllocId     string   `json:"AllocID,omitempty"`
	NodeId      string   `json:"NodeID,omitempty"`
	JobId       string   `json:"JobID,omitempty"`
	CreateIndex uint64   `json:"CreateIndex,omitempty"`
	ModifyIndex uint64   `json:"ModifyIndex,omitempty"`
	Tags        []string `json:"Tags,omitempty"`
}

// Output the response as a string
func (services NomadServicesResponse) String() string {
	str := "<nomad.services"
	if namespace := services.Namespace; namespace != "" {
		str += fmt.Sprintf(" namespace=%q", namespace)
	}
	if services := services.Services; len(services) > 0 {
		str += fmt.Sprintf(" services=%v", services)
	}
	return str + ">"
}

// Output the service as a string
func (service NomadService) String() string {
	str := "<nomad.service"
	if name := service.Name; name != "" {
		str += fmt.Sprintf(" name=%q", name)
	}
	if namespace := service.Namespace; namespace != "" {
		str += fmt.Sprintf(" namespace=%q", namespace)
	}
	if datacenter := service.Datacenter; datacenter != "" {
		str += fmt.Sprintf(" datacenter=%q", datacenter)
	}
	if address := service.Address; address != "" {
		str += fmt.Sprintf(" address=%q", address)
	}
	if port := service.Port; port > 0 {
		str += fmt.Sprintf(" port=%v", port)
	}
	if id := service.Id; id != "" {
		str += fmt.Sprintf(" id=%q", id)
	}
	if allocid := service.AllocId; allocid != "" {
		str += fmt.Sprintf(" allocId=%q", allocid)
	}
	if nodeid := service.NodeId; nodeid != "" {
		str += fmt.Sprintf(" nodeId=%q", nodeid)
	}
	if jobid := service.JobId; jobid != "" {
		str += fmt.Sprintf(" jobId=%q", jobid)
	}
	if createindex := service.CreateIndex; createindex > 0 {
		str += fmt.Sprintf(" createIndex=%v", createindex)
	}
	if modifyindex := service.ModifyIndex; modifyindex > 0 {
		str += fmt.Sprintf(" modifyIndex=%v", modifyindex)
	}
	if tags := service.Tags; len(tags) > 0 {
		str += fmt.Sprintf(" tags=%q", tags)
	}
	return str + ">"
}

// Return the version and uptime for nginx
func (c *Client) Services() ([]NomadServicesResponse, error) {
	var response []NomadServicesResponse
	if err := c.Do(nil, &response, client.OptPath(c.Prefix, "services")); err != nil {
		return response, err
	} else {
		return response, nil
	}
}

// Return a service based on the name
func (c *Client) Service(name string) ([]NomadService, error) {
	var response []NomadService
	if err := c.Do(nil, &response, client.OptPath(c.Prefix, "service", name)); err != nil {
		return response, err
	} else {
		return response, nil
	}
}
