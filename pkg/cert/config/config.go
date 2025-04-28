package config

import (
	"context"
	"time"

	// Packages
	server "github.com/mutablelogic/go-server"
	cert "github.com/mutablelogic/go-server/pkg/cert"
	certhandler "github.com/mutablelogic/go-server/pkg/cert/handler"
	schema "github.com/mutablelogic/go-server/pkg/cert/schema"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	pgqueue "github.com/mutablelogic/go-server/pkg/pgqueue/schema"
	"github.com/mutablelogic/go-server/pkg/ref"
	"github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Pool   server.PG         `kong:"-"`                                    // Connection pool
	Router server.HTTPRouter `kong:"-"`                                    // Which HTTP router to use
	Prefix string            `default:"${CERT_PREFIX}" help:"Path prefix"` // HTTP Path Prefix
	Root   struct {
		Organization  string `name:"org" help:"Organization name"`
		Unit          string `name:"unit" help:"Organization unit"`
		Country       string `name:"country" help:"Country name"`
		City          string `name:"city" help:"City name"`
		State         string `name:"state" help:"State name"`
		StreetAddress string `name:"street" help:"Street address"`
		PostalCode    string `name:"postal" help:"Postal code"`
	} `embed:"" prefix:"root."`
	Queue server.PGQueue `kong:"-"` // Connection to queue
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	rootCertName    = "root"
	rootCertKeyBits = 4096
	rootCertExpiry  = 16 * 365 * 24 * time.Hour
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (c Config) New(ctx context.Context) (server.Task, error) {
	if c.Pool == nil {
		return nil, httpresponse.ErrBadRequest.With("missing connection pool")
	}

	// Create a new certmanager
	certmanager, err := cert.NewCertManager(ctx, c.Pool.Conn(),
		cert.WithCommonName(rootCertName),
		cert.WithOrganization(c.Root.Organization, c.Root.Unit),
		cert.WithCountry(c.Root.Country, c.Root.State, c.Root.City),
		cert.WithAddress(c.Root.StreetAddress, c.Root.PostalCode),
		cert.WithRSAKey(rootCertKeyBits),
		cert.WithExpiry(rootCertExpiry),
		cert.WithCA(),
	)
	if err != nil {
		return nil, err
	}

	// Register HTTP handlers
	if c.Router != nil {
		certhandler.RegisterName(ctx, c.Router, c.Prefix, certmanager)
		certhandler.RegisterCert(ctx, c.Router, c.Prefix, certmanager)
	}

	// Queue ticker and queue to check for SSL expiry
	if c.Queue != nil {
		if _, err := c.Queue.RegisterTicker(ctx, pgqueue.TickerMeta{
			Ticker:   c.Name(),
			Interval: types.DurationPtr(time.Minute),
		}, func(ctx context.Context, _ any) error {
			ref.Log(ctx).Print(ctx, "Checking for SSL expiry...")
			return nil
		}); err != nil {
			return nil, err
		}
		if _, err := c.Queue.RegisterQueue(ctx, pgqueue.QueueMeta{
			Queue: c.Name(),
			TTL:   types.DurationPtr(time.Minute),
		}, func(ctx context.Context, _ any) error {
			ref.Log(ctx).Print(ctx, "Checking for SSL expiry...")
			return nil
		}); err != nil {
			return nil, err
		}
	}

	// Return the task
	return newTaskWith(certmanager), nil
}

////////////////////////////////////////////////////////////////////////////////
// MODULE

func (c Config) Name() string {
	return schema.SchemaName
}

func (c Config) Description() string {
	return "Certificate manager"
}
