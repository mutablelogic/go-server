package provider

import (
	"context"
	"sync"

	// Namespace imports
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Print logging message
func (self *provider) Print(ctx context.Context, v ...any) {
	var wg sync.WaitGroup
	for _, logger := range self.logger {
		wg.Add(1)
		go func(l Logger) {
			defer wg.Done()
			l.Print(ctx, v...)
		}(logger)
	}
	wg.Wait()
}

// Print logging message with format
func (self *provider) Printf(ctx context.Context, f string, v ...any) {
	var wg sync.WaitGroup
	for _, logger := range self.logger {
		wg.Add(1)
		go func(l Logger) {
			defer wg.Done()
			l.Printf(ctx, f, v...)
		}(logger)
	}
	wg.Wait()
}
