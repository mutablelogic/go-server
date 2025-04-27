package config

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"

	// Packages
	pg "github.com/djthorpe/go-pg"
	server "github.com/mutablelogic/go-server"
	pgqueue "github.com/mutablelogic/go-server/pkg/pgqueue"
	schema "github.com/mutablelogic/go-server/pkg/pgqueue/schema"
	ref "github.com/mutablelogic/go-server/pkg/ref"
	"github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type task struct {
	manager *pgqueue.Manager
	wg      sync.WaitGroup
}

var _ server.Task = (*task)(nil)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (task *task) Run(parent context.Context) error {
	var errs error

	// Create a cancelable context
	ctx, cancel := context.WithCancel(context.Background())
	ctx = ref.WithPath(ref.WithLog(ctx, ref.Log(parent)), ref.Path(parent)...)

	// Ticker loop
	tickerch := make(chan *schema.Ticker)
	defer close(tickerch)
	task.wg.Add(1)
	go func() {
		defer task.wg.Done()
		if err := task.manager.RunTickerLoop(ctx, tickerch); err != nil {
			errs = errors.Join(errs, err)
		}

	}()

	// Notification loop
	notifych := make(chan *pg.Notification)
	defer close(notifych)
	task.wg.Add(1)
	go func() {
		defer task.wg.Done()
		if err := task.manager.RunNotificationLoop(ctx, notifych); err != nil {
			errs = errors.Join(errs, err)
		}
	}()

	// Task loop
	taskch := make(chan *schema.Task)
	defer close(taskch)
	task.wg.Add(1)
	go func() {
		defer task.wg.Done()
		if err := task.manager.RunTaskLoop(ctx, taskch); err != nil {
			errs = errors.Join(errs, err)
		}
	}()

FOR_LOOP:
	for {
		select {
		case <-parent.Done():
			cancel()
			break FOR_LOOP
		case <-notifych:
			// Handle notification - try and retain a tasks
			for {
				evt, err := task.manager.NextTask(ctx)
				if err != nil {
					ref.Log(ctx).Print(parent, err)
					break
				}
				if evt == nil {
					break
				}
				task.tryTask(ctx, evt)
			}
		case evt := <-tickerch:
			task.tryTicker(ctx, evt)
		case evt := <-taskch:
			task.tryTask(ctx, evt)
		}
	}

	// Wait for all goroutines to finish
	task.wg.Wait()

	// Return any errors
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (t *task) tryTicker(ctx context.Context, ticker *schema.Ticker) error {
	return pgqueue.RunTicker(ctx, ticker, tickerFunc)
}

func (t *task) tryTask(ctx context.Context, task *schema.Task) (*schema.Task, error) {
	err := pgqueue.RunTask(ctx, task, taskFunc, task.Payload)

	if err != nil {
		// Fail the task
		if task, err := t.manager.ReleaseTask(ctx, task.Id, false, err.Error(), nil); err != nil {
			return nil, err
		} else if types.PtrUint64(task.Retries) == 0 {
			return task, errors.New("task failed, will not retry")
		} else {
			return task, nil
		}
	} else {
		// Succeed the task
		return t.manager.ReleaseTask(ctx, task.Id, true, nil, nil)
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST CALLBACKS

func taskFunc(ctx context.Context, payload any) error {
	var err error
	if rand.Intn(2) == 0 {
		err = errors.New("random error")
	}

	ref.Log(ctx).Debug(ctx, "Running task ", ref.Task(ctx), " with payload ", payload)
	select {
	case <-ctx.Done():
		ref.Log(ctx).Debug(ctx, "  Task cancelled")
		return ctx.Err()
	case <-time.After(time.Second * time.Duration(rand.Intn(10))):
		if err != nil {
			ref.Log(ctx).Debug(ctx, "  Task failed: ", err)
		} else {
			ref.Log(ctx).Debug(ctx, "  Task succeeded")
		}
	}
	return err
}

func tickerFunc(ctx context.Context, payload any) error {
	ref.Log(ctx).Debug(ctx, "Running ticker ", ref.Ticker(ctx))
	select {
	case <-ctx.Done():
		ref.Log(ctx).Debug(ctx, "  Ticker cancelled")
		return ctx.Err()
	case <-time.After(time.Second * time.Duration(rand.Intn(20))):
		ref.Log(ctx).Debug(ctx, "  Ticker done")
	}
	return nil
}
