package rotel

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	// Modules
	multierror "github.com/hashicorp/go-multierror"
	. "github.com/mutablelogic/go-server"
	term "github.com/pkg/term"
)

// Ref: https://www.rotel.com/sites/default/files/product/rs232/A12-A14%20Protocol.pdf

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	TTY     string
	Baud    uint
	Timeout time.Duration
}

type Manager struct {
	State

	tty string
	fd  *term.Term // TTY file handle
	buf *strings.Builder
	c   chan<- Event
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	DEFAULT_TTY         = "/dev/ttyUSB0"
	DEFAULT_TTY_BAUD    = 115200
	DEFAULT_TTY_TIMEOUT = 100 * time.Millisecond
	deltaUpdate         = 500 * time.Millisecond
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New(cfg Config, c chan<- Event) (*Manager, error) {
	this := new(Manager)

	// Set tty from config
	if cfg.TTY == "" {
		cfg.TTY = DEFAULT_TTY
	}
	if cfg.Baud == 0 {
		cfg.Baud = DEFAULT_TTY_BAUD
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = DEFAULT_TTY_TIMEOUT
	}

	// Check parameters
	if _, err := os.Stat(cfg.TTY); os.IsNotExist(err) {
		return nil, ErrBadParameter.With("tty: ", strconv.Quote(cfg.TTY))
	} else if err != nil {
		return nil, err
	} else {
		this.tty = cfg.TTY
	}

	// Open term
	if fd, err := term.Open(cfg.TTY, term.Speed(int(cfg.Baud)), term.RawMode); err != nil {
		return nil, err
	} else {
		this.fd = fd
		this.buf = new(strings.Builder)
	}

	// Set term read timeout
	if err := this.fd.SetReadTimeout(cfg.Timeout); err != nil {
		defer this.fd.Close()
		return nil, err
	}

	// Return success
	return this, nil
}

func (this *Manager) Run(ctx context.Context) error {
	// Update rotel status every 100ms
	timer := time.NewTimer(100 * time.Millisecond)
	defer timer.Stop()

	// Loop handling messages until done
FOR_LOOP:
	for {
		select {
		case <-ctx.Done():
			break FOR_LOOP
		case <-timer.C:
			if cmd := this.State.Update(); cmd != "" {
				if err := this.writetty(cmd); err != nil {
					fmt.Print("writetty:  ", err)
				}
			}
			timer.Reset(time.Millisecond * 250)
		default:
			if err := this.readtty(); err != nil {
				fmt.Print("readtty: ", err)
			}
		}
	}

	// Close RS232 connection
	var result error
	if this.fd != nil {
		if err := this.fd.Close(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Clear resources
	this.fd = nil
	this.buf = nil

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *Manager) SetPower(state bool) error {
	if state {
		return this.writetty("power_on!")
	} else {
		return this.writetty("power_off!")
	}
}

func (this *Manager) SetSource(value string) error {
	// Cannot set value when power is off
	if this.Power() == false {
		return ErrOutOfOrder.With("SetSource")
	}

	// Check parameter and send command
	switch value {
	case "pc_usb":
		return this.writetty("pcusb!")
	case "cd", "coax1", "coax2", "opt1", "opt2", "aux1", "aux2", "tuner", "photo", "usb", "bluetooth":
		return this.writetty(value + "!")
	default:
		return ErrBadParameter.With("SetSource")
	}
}

func (this *Manager) SetVolume(value uint) error {
	// Cannot set value when power is off
	if this.Power() == false {
		return ErrOutOfOrder.With("SetVolume")
	}

	// Check parameter and send command
	if value < 1 || value > 96 {
		return ErrBadParameter.With("SetVolume")
	} else {
		return this.writetty(fmt.Sprintf("vol_%02d!", value))
	}
}

func (this *Manager) SetMute(state bool) error {
	// Cannot set value when power is off
	if this.Power() == false {
		return ErrOutOfOrder.With("SetMute")
	}

	// Check parameter and send command
	if state {
		return this.writetty("mute_on!")
	} else {
		return this.writetty("mute_off!")
	}
}

func (this *Manager) SetBypass(state bool) error {
	// Cannot set value when power is off
	if this.Power() == false {
		return ErrOutOfOrder.With("SetBypass")
	}

	// Check parameter and send command
	if state {
		return this.writetty("bypass_on!")
	} else {
		return this.writetty("bypass_off!")
	}
}

func (this *Manager) SetTreble(value int) error {
	// Cannot set value when power is off
	if this.Power() == false {
		return ErrOutOfOrder.With("SetTreble")
	}

	// Check parameter and send command
	if value < -10 || value > 10 {
		return ErrBadParameter.With("SetTreble")
	} else if value == 0 {
		return this.writetty("treble_000!")
	} else if value < 0 {
		return this.writetty(fmt.Sprint("treble_", value, "!"))
	} else {
		return this.writetty(fmt.Sprint("treble_+", value, "!"))
	}
}

func (this *Manager) SetBass(value int) error {
	// Cannot set value when power is off
	if this.Power() == false {
		return ErrOutOfOrder.With("SetBass")
	}

	// Check parameter and send command
	if value < -10 || value > 10 {
		return ErrBadParameter.With("SetBass")
	} else if value == 0 {
		return this.writetty("bass_000!")
	} else if value < 0 {
		return this.writetty(fmt.Sprint("bass_", value, "!"))
	} else {
		return this.writetty(fmt.Sprint("bass_+", value, "!"))
	}
}

func (this *Manager) SetBalance(loc string) error {
	// Cannot set value when power is off
	if this.Power() == false {
		return ErrOutOfOrder.With("SetBalance")
	}

	// Check parameter and send command
	switch loc {
	case "0":
		return this.writetty("balance_000!")
	case "L", "R":
		return this.writetty(fmt.Sprintf("balance_%v!", strings.ToLower(loc)))
	default:
		return ErrBadParameter.With("SetBalance")
	}
}

func (this *Manager) SetDimmer(value uint) error {
	// Cannot set value when power is off
	if this.Power() == false {
		return ErrOutOfOrder.With("SetDimmer")
	}

	// Check parameter and send command
	if value > 6 {
		return ErrBadParameter.With("SetDimmer")
	} else {
		return this.writetty(fmt.Sprint("dimmer_", value, "!"))
	}
}

func (this *Manager) Play() error {
	// Cannot perform action when power is off
	if this.Power() == false {
		return ErrOutOfOrder.With("Play")
	}

	// Send command
	return this.writetty("play!")
}

func (this *Manager) Stop() error {
	// Cannot perform action when power is off
	if this.Power() == false {
		return ErrOutOfOrder.With("Stop")
	}

	// Send command
	return this.writetty("stop!")
}

func (this *Manager) Pause() error {
	// Cannot perform action when power is off
	if this.Power() == false {
		return ErrOutOfOrder.With("Pause")
	}

	// Send command
	return this.writetty("pause!")
}

func (this *Manager) NextTrack() error {
	// Cannot perform action when power is off
	if this.Power() == false {
		return ErrOutOfOrder.With("NextTrack")
	}

	// Send command
	return this.writetty("trkf!")
}

func (this *Manager) PrevTrack() error {
	// Cannot perform action when power is off
	if this.Power() == false {
		return ErrOutOfOrder.With("PrevTrack")
	}

	// Send command
	return this.writetty("trkb!")
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Manager) String() string {
	str := "<rotel"
	if this.fd != nil {
		str += fmt.Sprintf(" tty=%q", this.tty)
	}
	str += fmt.Sprint(" ", this.State.String())
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHDOS

func (this *Manager) readtty() error {
	var result error
	var flags RotelFlag

	// Append data to the buffer and parse any parameters
	buf := make([]byte, 1024)
	if n, err := this.fd.Read(buf); err == io.EOF {
		return nil
	} else if err != nil {
		return err
	} else if _, err := this.buf.Write(buf[:n]); err != nil {
		return err
	} else if fields := strings.Split(this.buf.String(), "$"); len(fields) > 0 {
		// Parse each field and update state
		for _, param := range fields[0 : len(fields)-1] {
			if flag, err := this.State.Set(param); err != nil {
				result = multierror.Append(result, fmt.Errorf("%q: %w", param, err))
			} else {
				flags |= flag
			}
		}
		// Reset buffer with any remaining data not parsed
		this.buf.Reset()
		this.buf.WriteString(fields[len(fields)-1])
	}

	// If any flags set, then emit an event
	if flags != ROTEL_FLAG_NONE {
		this.c <- Event{this.State, flags}
	}

	// Return any errors
	return result
}

func (this *Manager) writetty(cmd string) error {
	_, err := this.fd.Write([]byte(cmd))
	return err
}
