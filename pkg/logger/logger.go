package logger

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	// Packages
	hcl "github.com/mutablelogic/go-hcl/pkg/block"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
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

	name string
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

func (c Config) New(context.Context) (hcl.Resource, error) {
	self := new(logger)
	self.name = c.Name()

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
	str := "<block "
	str += fmt.Sprintf("name=%q", c.Name())
	if len(c.Flags) > 0 {
		str += fmt.Sprintf(" flags=%q", c.Flags)
	}
	return str + ">"
}

func (self *logger) String() string {
	str := "<" + self.name
	return str + ">"
}

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (self *logger) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (self *logger) Print(ctx context.Context, v ...any) {
	self.Mutex.Lock()
	defer self.Mutex.Unlock()
	self.Logger.Print(v...)
}

func (self *logger) Printf(ctx context.Context, f string, v ...any) {
	self.Mutex.Lock()
	defer self.Mutex.Unlock()
	self.Logger.Printf(f, v...)
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
