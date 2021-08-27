package chromecast

import (
	"net"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type device struct {
	id, fn string
	md, rs string
	st     uint
	host   string
	ips    []net.IP
	port   uint16
	//vol *Volume
	//app *App
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (this *device) connect() error {
	return nil
}

func (this *device) disconnect() error {
	return nil
}
