package log

import (
	"context"
	"log"
	"strings"

	// Package imports
	iface "github.com/mutablelogic/go-server"
	task "github.com/mutablelogic/go-server/pkg/task"
	types "github.com/mutablelogic/go-server/pkg/types"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Plugin struct {
	task.Plugin
	Prefix_ string   `json:"prefix" hcl:"prefix"`
	Flags_  []string `json:"flags,omitempty" hcl:"flags,optional"`
	flags   int
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultName = "log"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new logger task with provider of other tasks
func (p Plugin) New(_ context.Context, _ iface.Provider) (iface.Task, error) {
	if !types.IsIdentifier(p.Name()) {
		return nil, ErrBadParameter.With("Invalid plugin name: %q", p.Name())
	}
	if !types.IsIdentifier(p.Label()) {
		return nil, ErrBadParameter.With("Invalid plugin label: %q", p.Label())
	}
	if flags, err := flagsForSlice(p.Flags_); err != nil {
		return nil, err
	} else {
		p.flags = flags
	}

	return NewWithPlugin(p)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (p Plugin) Name() string {
	if name := p.Plugin.Name(); name != "" {
		return name
	} else {
		return defaultName
	}
}

func (p Plugin) Prefix() string {
	if p.Prefix_ != "" {
		return p.Prefix_
	} else {
		return p.Label()
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - MIDDLEWARE

func flagsForSlice(flags []string) (int, error) {
	var result int
	for _, flag := range flags {
		flag = strings.ToLower(flag)
		switch flag {
		case "default", "standard", "std":
			result |= log.LstdFlags
		case "date":
			result |= log.Ldate
		case "time":
			result |= log.Ltime
		case "microseconds", "ms":
			result |= log.Lmicroseconds
		case "utc":
			result |= log.LUTC
		case "msgprefix", "prefix":
			result |= log.Lmsgprefix
		default:
			return 0, ErrBadParameter.Withf("flag: %q", flag)
		}
	}
	// Return success
	return result, nil
}
