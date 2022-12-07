package gateway_test

import (
	"context"
	"testing"

	// Packages
	"github.com/mutablelogic/go-server/pkg/event"
	"github.com/mutablelogic/go-server/pkg/nginx/gateway"
	"github.com/mutablelogic/go-server/pkg/types"
)

func Test_Task_000(t *testing.T) {
	task, err := gateway.NewWithPlugin(gateway.Plugin{
		Nginx_: types.Task{
			Task: NewTask(t),
		},
	})
	if err != nil {
		t.Skip("Skipping test:", err)
	}
	t.Log(task)
}

///////////////////////////////////////////////////////////////////////////////
// MOCK NGINX TASK

type nginx struct {
	event.Source
	*testing.T
}

func NewTask(t *testing.T) *nginx {
	return &nginx{T: t}
}

func (n *nginx) String() string {
	return "<nginx.mock>"
}

func (n *nginx) Run(ctx context.Context) error {
	n.Log(n.Name(), "Run()")
	<-ctx.Done()
	return nil
}

func (n *nginx) Test() error {
	n.Log(n.Name(), "Test()")
	return nil
}

func (n *nginx) Reopen() error {
	n.Log(n.Name(), "Reopen()")
	return nil
}

func (n *nginx) Reload() error {
	n.Log(n.Name(), "Reload()")
	return nil
}
