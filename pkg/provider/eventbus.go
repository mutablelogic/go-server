package provider

import (
	"context"

	. "github.com/djthorpe/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// EVENT BUS IMPLEMENTATION

// Post events to one or more event busses
func (this *provider) Post(ctx context.Context, evt Event) {
	for _, bus := range this.eventbus {
		bus.Post(ctx, evt)
	}
}

// Subscribe to get events from one or more event buses
func (this *provider) Subscribe(ctx context.Context, c chan<- Event) {
	for _, bus := range this.eventbus {
		bus.Subscribe(ctx, c)
	}
}
