package nginx_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/mutablelogic/go-server/pkg/nginx"
)

const (
	NginxConfig = `../../etc/test/nginx/nginx.conf`
)

func Test_Task_000(t *testing.T) {
	task, err := nginx.NewWithPlugin(nginx.Plugin{
		Path_:   "nginx",
		Config_: NginxConfig,
	})
	if err != nil {
		t.Skip("Skipping test: ", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := task.Run(ctx); err != nil {
		t.Error(err)
	}
	t.Log(task)
}

func Test_Task_001(t *testing.T) {
	type N interface {
		Reload() error
	}

	task, err := nginx.NewWithPlugin(nginx.Plugin{
		Path_:   "nginx",
		Config_: NginxConfig,
	})
	if err != nil {
		t.Skip("Skipping test: ", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := task.Run(ctx); err != nil {
			t.Error(err)
		}
	}()
	go func() {
		defer wg.Done()
		time.Sleep(2 * time.Second)
		if err := N(task).Reload(); err != nil {
			t.Error(err)
		}
	}()
	wg.Wait()
}

func Test_Task_002(t *testing.T) {
	type N interface {
		Test() error
	}

	task, err := nginx.NewWithPlugin(nginx.Plugin{
		Path_:   "nginx",
		Config_: NginxConfig,
	})
	if err != nil {
		t.Skip("Skipping test: ", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := task.Run(ctx); err != nil {
			t.Error(err)
		}
	}()
	go func() {
		defer wg.Done()
		time.Sleep(2 * time.Second)
		if err := N(task).Test(); err != nil {
			t.Error(err)
		}
	}()
	wg.Wait()
}
