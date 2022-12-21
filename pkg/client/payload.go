package client

import (
	"net/http"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Payload interface {
	Method() string
	Accept() string
}

type payload struct {
	method string
	accept string
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewGetPayload(accept string) Payload {
	this := new(payload)
	this.method = http.MethodGet
	this.accept = accept
	return this
}

///////////////////////////////////////////////////////////////////////////////
// PAYLOAD METHODS

func (this *payload) Method() string {
	return this.method
}

func (this *payload) Accept() string {
	return this.accept
}
