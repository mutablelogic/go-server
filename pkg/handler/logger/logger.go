package logger

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
	"sync"

	// Namespace imports
	. "github.com/djthorpe/go-errors"

	// Packages
	"github.com/mutablelogic/go-server"
	"github.com/mutablelogic/go-server/pkg/provider"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// logging configuration
type Config struct {
	Flags []string `hcl:"flags,optional" description:"zero or more formatting flags (std, date, time, ms, utc, prefix)"`
}

// logging instance
type logger struct {
	sync.Mutex
	*log.Logger
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultName = "logger"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (c Config) Name() string {
	return defaultName
}

func (c Config) Description() string {
	return "logs messages to stdout"
}

func (c Config) New() (server.Task, error) {
	self := new(logger)

	if flags, err := flagsForSlice(c.Flags); err != nil {
		return nil, err
	} else {
		self.Logger = log.New(os.Stderr, "", flags)
	}

	// Return success
	return self, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (c Config) String() string {
	data, _ := json.MarshalIndent(c, "", "  ")
	return string(data)
}

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (*logger) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (l *logger) Label() string {
	// TODO
	return defaultName
}

func (l *logger) Print(ctx context.Context, v ...any) {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()
	if label := provider.Label(ctx); label != "" {
		l.SetPrefix("[" + label + "] ")
	} else {
		l.SetPrefix("")
	}
	l.Logger.Print(v...)
}

func (l *logger) Printf(ctx context.Context, f string, v ...any) {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()
	if label := provider.Label(ctx); label != "" {
		l.SetPrefix("[" + label + "] ")
	} else {
		l.SetPrefix("")
	}
	l.Logger.Printf(f, v...)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

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
