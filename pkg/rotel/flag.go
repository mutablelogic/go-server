package rotel

import "strings"

////////////////////////////////////////////////////////////////////////////////
// TYPES

// RotelFlag provides flags on state changes
type RotelFlag uint16

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	ROTEL_FLAG_POWER RotelFlag = (1 << iota)
	ROTEL_FLAG_VOLUME
	ROTEL_FLAG_MUTE
	ROTEL_FLAG_BASS
	ROTEL_FLAG_TREBLE
	ROTEL_FLAG_BALANCE
	ROTEL_FLAG_SOURCE
	ROTEL_FLAG_FREQ
	ROTEL_FLAG_BYPASS
	ROTEL_FLAG_SPEAKER
	ROTEL_FLAG_DIMMER
	ROTEL_FLAG_NONE RotelFlag = 0
	ROTEL_FLAG_MIN            = ROTEL_FLAG_POWER
	ROTEL_FLAG_MAX            = ROTEL_FLAG_DIMMER
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (f RotelFlag) String() string {
	if f == ROTEL_FLAG_NONE {
		return f.FlagString()
	}
	str := ""
	for v := ROTEL_FLAG_MIN; v <= ROTEL_FLAG_MAX; v <<= 1 {
		if v&f == v {
			str += "|" + v.FlagString()
		}
	}
	return strings.TrimPrefix(str, "|")
}

func (f RotelFlag) FlagString() string {
	switch f {
	case ROTEL_FLAG_NONE:
		return "ROTEL_FLAG_NONE"
	case ROTEL_FLAG_POWER:
		return "ROTEL_FLAG_POWER"
	case ROTEL_FLAG_VOLUME:
		return "ROTEL_FLAG_VOLUME"
	case ROTEL_FLAG_MUTE:
		return "ROTEL_FLAG_MUTE"
	case ROTEL_FLAG_BASS:
		return "ROTEL_FLAG_BASS"
	case ROTEL_FLAG_TREBLE:
		return "ROTEL_FLAG_TREBLE"
	case ROTEL_FLAG_BALANCE:
		return "ROTEL_FLAG_BALANCE"
	case ROTEL_FLAG_SOURCE:
		return "ROTEL_FLAG_SOURCE"
	case ROTEL_FLAG_FREQ:
		return "ROTEL_FLAG_FREQ"
	case ROTEL_FLAG_BYPASS:
		return "ROTEL_FLAG_BYPASS"
	case ROTEL_FLAG_SPEAKER:
		return "ROTEL_FLAG_SPEAKER"
	case ROTEL_FLAG_DIMMER:
		return "ROTEL_FLAG_DIMMER"
	default:
		return "[?? Invalid RotelFlag value]"
	}
}
