package cmd

import (
	"context"
	"fmt"
	"math/big"
	"time"

	server "github.com/mutablelogic/go-server"
	"github.com/mutablelogic/go-server/pkg/cert/schema"
	client "github.com/mutablelogic/go-server/plugin/certmanager/client"
)

// Packages

///////////////////////////////////////////////////////////////////////////////
// TYPES

type CertCommands struct {
	Cert       CertGetCommand    `cmd:"" group:"CERTIFICATES" help:"Get a certificate"`
	CreateCert CertCreateCommand `cmd:"" group:"CERTIFICATES" help:"Create a new signed certificate"`
	DeleteCert CertDeleteCommand `cmd:"" group:"CERTIFICATES" help:"Delete a certificate"`
}

// Certificate Metadata for creating a new certificate
type CertCreateCommand struct {
	Name         string        `arg:"" name:"name" help:"Certificate name"`
	Domain       string        `arg:"" name:"domain" help:"Domain the certificate is for"`
	Signer       string        `name:"signer" help:"Signer name, if certificate is to be signed by a CA"`
	Subject      string        `name:"subject" help:"Subject name"`
	SerialNumber *big.Int      `name:"serial" help:"Serial number"`
	Expiry       time.Duration `name:"expiry" help:"Expiry duration"`
	IsCA         bool          `name:"ca" help:"Certificate is a CA"`
	KeyType      string        `name:"type" help:"Key type"`
	Hosts        []string      `name:"host" help:"Comma-separated list of host names"`
	IP           []string      `name:"ip" help:"Comma-separated list of IP addresses"`
}

type CertGetCommand struct {
	Name string `arg:"" name:"name" help:"Certificate Name"`
}

type CertDeleteCommand struct {
	CertGetCommand
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd CertCreateCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		cert, err := provider.CreateCert(ctx, schema.CertCreateMeta{
			Name:         cmd.Name,
			CommonName:   cmd.Domain,
			Signer:       cmd.Signer,
			Subject:      cmd.Subject,
			SerialNumber: cmd.SerialNumber,
			Expiry:       cmd.Expiry,
			IsCA:         cmd.IsCA,
			KeyType:      cmd.KeyType,
		})
		if err != nil {
			return err
		}

		// Print cert
		fmt.Println(cert)
		return nil
	})
}

func (cmd CertGetCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		name, err := provider.GetCert(ctx, cmd.Name)
		if err != nil {
			return err
		}

		// Print name
		fmt.Println(name)
		return nil
	})
}

func (cmd CertDeleteCommand) Run(ctx server.Cmd) error {
	return run(ctx, func(ctx context.Context, provider *client.Client) error {
		return provider.DeleteCert(ctx, cmd.Name)
	})
}
