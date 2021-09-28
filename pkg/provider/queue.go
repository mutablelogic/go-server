package provider

import (
	"context"

	// Modules
	"github.com/hashicorp/go-multierror"

	// Namespace imports
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// EVENT QUEUE IMPLEMENTATION

// Post events to one or more event queues
func (this *provider) Post(ctx context.Context, evt Event) {
	for _, bus := range this.queue {
		bus.Post(ctx, evt)
	}
}

// Subscribe to get events from one or more event queues
func (this *provider) Subscribe(ctx context.Context, c chan<- Event) error {
	var result error
	if len(this.queue) == 0 {
		this.Print(ctx, "Subscribe: No event queues to subscribe to")
		return nil
	}
	for _, bus := range this.queue {
		if err := bus.Subscribe(ctx, c); err != nil {
			result = multierror.Append(result, err)
		}
	}
	return result
}
