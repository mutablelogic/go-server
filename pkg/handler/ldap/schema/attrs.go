package schema

import "fmt"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Attr func(*Object) error

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Set description
func OptDescription(v string) Attr {
	return func(o *Object) error {
		o.Set("description", v)
		return nil
	}
}

// Set Group ID
func OptGroupId(v int) Attr {
	return func(o *Object) error {
		if v < 0 {
			return fmt.Errorf("OptGroupId: invalid value")
		}
		o.Set("gidNumber", fmt.Sprint(v))
		return nil
	}
}

// Set User ID
func OptUserId(v int) Attr {
	return func(o *Object) error {
		if v < 0 {
			return fmt.Errorf("OptGroupId: invalid value")
		}
		o.Set("uidNumber", fmt.Sprint(v))
		return nil
	}
}

// Set Name
func OptName(givenName, surname string) Attr {
	return func(o *Object) error {
		if givenName != "" {
			o.Set("givenName", givenName)
		}
		if surname != "" {
			o.Set("sn", surname)
		}
		return nil
	}
}

// Set Mail
func OptMail(v string) Attr {
	return func(o *Object) error {
		o.Set("mail", v)
		return nil
	}
}

// Set Home Directory
func OptHomeDirectory(v string) Attr {
	return func(o *Object) error {
		o.Set("homeDirectory", v)
		return nil
	}
}

// Set Login Shell
func OptLoginShell(v string) Attr {
	return func(o *Object) error {
		o.Set("loginShell", v)
		return nil
	}
}
