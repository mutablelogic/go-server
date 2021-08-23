package provider

import (
	"context"
	"log"
)

///////////////////////////////////////////////////////////////////////////////
// LOGGER IMPLEMENTATION

func (this *provider) Print(ctx context.Context, args ...interface{}) {
	for _, logger := range this.loggers {
		logger.Print(ctx, args...)
	}
	if len(this.loggers) == 0 {
		log.Print(args...)
	}
}

func (this *provider) Printf(ctx context.Context, f string, args ...interface{}) {
	for _, logger := range this.loggers {
		logger.Printf(ctx, f, args...)
	}
	if len(this.loggers) == 0 {
		log.Printf(f, args...)
	}
}
