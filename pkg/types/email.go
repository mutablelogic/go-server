package types

import "net/mail"

// Parse an email address into a name and address parts, and return true if
// the email address is valid
func IsEmail(addr string, name, email *string) bool {
	parsed, err := mail.ParseAddress(addr)
	if err != nil {
		return false
	}
	if name != nil {
		*name = parsed.Name
	}
	if email != nil {
		*email = parsed.Address
	}
	return true
}
