package tokenjar

import (
	"context"
	"path/filepath"
	"time"

	// Packages
	provider "github.com/mutablelogic/go-server/pkg/provider"
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
			if jar.Modified() {
				if err := jar.Write(); err != nil {
					logger.Print(ctx, err)
				} else {
					logger.Printf(ctx, "Sync %q", filepath.Base(jar.filename))
				}
			}
		case <-ctx.Done():
			return jar.Write()
		}
	}
}
