package tokenjar

import (
	"context"
	"time"

	"github.com/mutablelogic/go-server/pkg/provider"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (jar *tokenjar) Label() string {
	// TODO
	return defaultName
}

func (jar *tokenjar) Run(ctx context.Context) error {
	// Get logger
	logger := provider.Logger(ctx)

	// Ticker for writing to disk
	ticker := time.NewTicker(jar.writeInterval)
	defer ticker.Stop()

	// Loop until cancelled
	for {
		select {
		case <-ticker.C:
			if err := jar.write(); err != nil {
				logger.Print(ctx, err)
			}
		case <-ctx.Done():
			return jar.write()
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (jar *tokenjar) write() error {
	if !jar.Modified() {
		return nil
	} else {
		return jar.Write()
	}
}
