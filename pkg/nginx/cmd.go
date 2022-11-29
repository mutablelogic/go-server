package nginx

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"

	// Packages
	types "github.com/mutablelogic/go-server/pkg/types"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Cmd struct {
	*exec.Cmd
	Out, Err CallbackFn
}

// Callback output from the command. Newlines are embedded
// within the string
type CallbackFn func(*Cmd, []byte)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new logger task with provider of other tasks
func NewWithCommand(cmd string, args ...string) (*Cmd, error) {
	this := new(Cmd)
	if !filepath.IsAbs(cmd) {
		if cmd_, err := exec.LookPath(cmd); err != nil {
			return nil, err
		} else {
			cmd = cmd_
		}
	}
	if stat, err := os.Stat(cmd); err != nil {
		return nil, err
	} else if !IsExecAny(stat.Mode()) {
		return nil, ErrBadParameter.Withf("Command is not executable: %q", cmd)
	} else {
		this.Cmd = exec.Command(cmd, args...)
	}

	// Return success
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t *Cmd) String() string {
	str := "<cmd"
	if t.Cmd != nil {
		str += fmt.Sprintf(" exec=%q", t.Cmd.Path)
		if len(t.Cmd.Args) > 1 {
			str += fmt.Sprintf(" args=%q", t.Cmd.Args[1:])
		}
		for _, v := range t.Cmd.Env {
			str += fmt.Sprint(v)
		}
		if t.Cmd.Process != nil {
			if pid := t.Cmd.Process.Pid; pid > 0 {
				str += fmt.Sprintf(" pid=%d", pid)
			}
			if t.Cmd.ProcessState.Exited() {
				str += fmt.Sprintf(" exit_code=%d", t.Cmd.ProcessState.ExitCode())
			}
		}
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Env appends the environment variables for the command
func (c *Cmd) Env(env map[string]string) error {
	for k, v := range env {
		if !types.IsIdentifier(k) {
			return ErrBadParameter.Withf("Invalid environment variable name: %q", k)
		}
		c.Cmd.Env = append(c.Cmd.Env, fmt.Sprintf("%s=%q", k, v))
	}
	// return success
	return nil
}

// Run the command in the foreground, and return any errors
func (c *Cmd) Run() error {
	var wg sync.WaitGroup

	// Check to see if command has alreay been run
	if c.Cmd.Process != nil {
		return ErrOutOfOrder.With("Command has already been run")
	}

	// Pipes for reading stdout and stderr
	if c.Out != nil {
		stdout, err := c.StdoutPipe()
		if err != nil {
			return err
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.read(stdout, c.Out)
		}()
	}
	if c.Err != nil {
		stderr, err := c.StderrPipe()
		if err != nil {
			return err
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.read(stderr, c.Err)
		}()
	}

	// Start command
	if err := c.Start(); err != nil {
		return err
	}

	// Wait for stdout and stderr to be closed
	wg.Wait()

	// Wait for command to exit, return any errors
	return c.Wait()
}

// Return whether command has exited
func (c *Cmd) Exited() bool {
	if c.Cmd.ProcessState == nil {
		return false
	} else {
		return c.Cmd.ProcessState.Exited()
	}
}

// Return the pid of the process or 0
func (c *Cmd) Pid() int {
	if c.Cmd.Process == nil {
		return 0
	} else {
		return c.Cmd.Process.Pid
	}
}

// Send signal to the process
func (c *Cmd) Signal(s os.Signal) error {
	pid := c.Pid()
	if pid == 0 || c.Exited() {
		return ErrOutOfOrder.With("Cannot signal exited process")
	} else if err := syscall.Kill(pid, s.(syscall.Signal)); err != nil {
		return err
	} else {
		return nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (t *Cmd) read(r io.ReadCloser, fn CallbackFn) {
	buf := bufio.NewReader(r)
	for {
		if line, err := buf.ReadBytes('\n'); err == io.EOF {
			return
		} else if err != nil && t.Err != nil {
			t.Err(t, []byte(err.Error()))
		} else {
			fn(t, line)
		}
	}
}
