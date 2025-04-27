package auth

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Opt is a function which applies options
type Opt func(*opts) error

type opts struct {
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func apply(opt ...Opt) (*opts, error) {
	// Create new options
	o := new(opts)

	// Apply options
	for _, fn := range opt {
		if err := fn(o); err != nil {
			return nil, err
		}
	}

	// Return success
	return o, nil
}
