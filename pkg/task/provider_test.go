package task_test

import (
	"context"
	"strings"
	"testing"
	"time"

	// Package imports
	log "github.com/mutablelogic/go-server/pkg/log"
	task "github.com/mutablelogic/go-server/pkg/task"
	types "github.com/mutablelogic/go-server/pkg/types"
	plugin "github.com/mutablelogic/go-server/plugin"
	assert "github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-server"
)

func Test_provider_000(t *testing.T) {
	assert := assert.New(t)
	provider, error := task.NewProvider(context.Background(), log.WithLabel(t.Name()))
	assert.NoError(error)
	assert.NotNil(provider)
	log, ok := provider.Get("log").(plugin.Log)
	assert.True(ok)
	assert.NotNil(log)
	log2, ok := provider.Get("log", t.Name()).(plugin.Log)
	assert.True(ok)
	assert.NotNil(log2)
}

func Test_provider_001(t *testing.T) {
	assert := assert.New(t)
	plugins := []Plugin{
		log.WithLabel(t.Name()),
		MockTaskPlugin{}.WithNameLabel("a", t.Name()).WithRef1("log", t.Name()).WithRef2("b", t.Name()),
		MockTaskPlugin{}.WithNameLabel("b", t.Name()).WithRef1("log", t.Name()),
	}
	provider, error := task.NewProvider(context.Background(), plugins...)
	assert.NoError(error)
	assert.NotNil(provider)
	// TODO: Order should be log, b, a
}

func Test_provider_002(t *testing.T) {
	assert := assert.New(t)
	plugins := []Plugin{
		log.WithLabel(t.Name()),
		MockTaskPlugin{}.WithNameLabel("a", t.Name()).WithRef1("log", t.Name()).WithRef2("b", t.Name()),
		MockTaskPlugin{}.WithNameLabel("b", t.Name()).WithRef1("log", t.Name()),
		MockTaskPlugin{}.WithNameLabel("c", t.Name()).WithRef1("a", t.Name()).WithRef2("b", t.Name()),
	}
	provider, error := task.NewProvider(context.Background(), plugins...)
	assert.NoError(error)
	assert.NotNil(provider)
	// TODO: Order should be log, b, a, c
}

func Test_provider_003(t *testing.T) {
	assert := assert.New(t)
	plugins := []Plugin{
		log.WithLabel(t.Name()),
		MockTaskPlugin{}.WithNameLabel("a", t.Name()).WithRef1("log", t.Name()).WithRef2("b", t.Name()),
		MockTaskPlugin{}.WithNameLabel("b", t.Name()).WithRef1("log", t.Name()).WithRef2("c", t.Name()),
		MockTaskPlugin{}.WithNameLabel("c", t.Name()).WithRef1("a", t.Name()),
	}
	_, error := task.NewProvider(context.Background(), plugins...)
	assert.Error(ErrOutOfOrder, error)
	// Circular reference error
}

func Test_provider_004(t *testing.T) {
	assert := assert.New(t)
	plugins := []Plugin{
		log.WithLabel(t.Name()),
		MockTaskPlugin{}.WithNameLabel("a", t.Name()).WithRef1("log", t.Name()).WithRef2("b", t.Name()),
		MockTaskPlugin{}.WithNameLabel("b", t.Name()).WithRef1("log", t.Name()),
		MockTaskPlugin{}.WithNameLabel("c", t.Name()).WithRef1("a", t.Name()).WithRef2("b", t.Name()),
	}
	provider, error := task.NewProvider(context.Background(), plugins...)
	assert.NoError(error)
	assert.NotNil(provider)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := provider.Run(ctx)
	assert.NoError(err)
}

////////

type MockTaskPlugin struct {
	task.Plugin
	Ref1 types.Task
	Ref2 types.Task
}

func (p MockTaskPlugin) WithNameLabel(name, label string) MockTaskPlugin {
	p.Plugin.Name_ = types.String(name)
	p.Plugin.Label_ = types.String(label)
	return p
}

func (p MockTaskPlugin) WithRef1(label ...string) MockTaskPlugin {
	p.Ref1.Ref = strings.Join(label, ".")
	return p
}

func (p MockTaskPlugin) WithRef2(label ...string) MockTaskPlugin {
	p.Ref2.Ref = strings.Join(label, ".")
	return p
}

func (p MockTaskPlugin) New(_ context.Context, _ Provider) (Task, error) {
	return &task.Task{}, nil
}

func (p MockTaskPlugin) Run(ctx context.Context) error {
	if p.Ref1.Ref != "" {
		if p.Ref1.Task == nil {
			return ErrInternalAppError.With("Ref1 not found", p.Ref1.Ref)
		}
	}
	if p.Ref2.Ref != "" {
		if p.Ref2.Task == nil {
			return ErrInternalAppError.With("Ref2 not found", p.Ref2.Ref)
		}
	}
	<-ctx.Done()
	return nil
}
