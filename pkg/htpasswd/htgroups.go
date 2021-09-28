package htpasswd

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strings"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Htgroups struct {
	groups map[string][]string
	users  map[string][]string
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	groupSeparator = ":"
	GroupPrefix    = "@"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewGroups() *Htgroups {
	this := new(Htgroups)
	this.groups = make(map[string][]string)
	this.users = make(map[string][]string)
	return this
}

func ReadGroups(r io.Reader) (*Htgroups, error) {
	this := NewGroups()

	// Scan the file
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); line == "" {
			continue
		} else if strings.HasPrefix(line, "#") {
			continue
		} else if group := strings.SplitN(line, groupSeparator, 2); len(group) != 2 {
			continue
		} else if !rePOSIXName.MatchString(group[0]) {
			return nil, ErrBadParameter.With("group", group[0])
		} else {
			users := strings.FieldsFunc(group[1], func(r rune) bool {
				return r == ',' || r == ' ' || r == ':'
			})
			for _, user := range users {
				if rePOSIXName.MatchString(user) {
					this.groups[group[0]] = append(this.groups[group[0]], user)
					this.users[user] = append(this.users[user], group[0])
				}
			}
		}
	}

	// Return any errors
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Return success
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Htgroups) String() string {
	str := "<htgroups"
	for group, users := range this.groups {
		str += fmt.Sprintf(" %s%v=%q", GroupPrefix, group, users)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return all groups
func (this *Htgroups) Groups() []string {
	result := make([]string, 0, len(this.groups))
	for group := range this.groups {
		result = append(result, group)
	}
	sort.Strings(result)
	return result
}

// Return groups for a user
func (this *Htgroups) GroupsForUser(user string) []string {
	groups, ok := this.users[user]
	if !ok {
		return nil
	} else {
		return groups
	}
}

// Return users for a group
func (this *Htgroups) UsersForGroup(group string) []string {
	users, ok := this.groups[group]
	if !ok {
		return nil
	}
	sort.Strings(users)
	return users
}

// Return true if a user is a member of a group
func (this *Htgroups) UserInGroup(user, group string) bool {
	groups, ok := this.users[user]
	if !ok {
		return false
	}
	for _, g := range groups {
		if g == group {
			return true
		}
	}
	// Not found in groups
	return false
}

// Add user to a group
func (this *Htgroups) AddUserToGroup(user, group string) error {
	// Return if user already in group
	if this.UserInGroup(user, group) {
		return nil
	}

	// Check user and group names
	if !rePOSIXName.MatchString(group) {
		return ErrBadParameter.Withf("group %q", group)
	}
	if !rePOSIXName.MatchString(user) {
		return ErrBadParameter.Withf("user %q", user)
	}

	// Append user to group and group to user and sort
	this.groups[group] = append(this.groups[group], user)
	this.users[user] = append(this.users[user], group)

	// Return success
	return nil
}

// Remove user from a group
func (this *Htgroups) RemoveUserFromGroup(user, group string) error {
	// Return if user not in group
	if !this.UserInGroup(user, group) {
		return nil
	}

	// Check user and group names
	if !rePOSIXName.MatchString(group) {
		return ErrBadParameter.Withf("group %q", group)
	}
	if !rePOSIXName.MatchString(user) {
		return ErrBadParameter.Withf("user %q", user)
	}

	// Remove user from group, etc
	this.groups[group] = removeString(this.groups[group], user)
	this.users[user] = removeString(this.users[user], group)

	// Remove user or group if empty
	if len(this.groups[group]) == 0 {
		delete(this.groups, group)
	}
	if len(this.users[user]) == 0 {
		delete(this.users, user)
	}

	// Return success
	return nil
}

// Write out group file
func (this *Htgroups) Write(w io.Writer) error {
	for group, users := range this.groups {
		if _, err := fmt.Fprintln(w, group+groupSeparator+strings.Join(users, " ")); err != nil {
			return err
		}
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func removeString(arr []string, elem string) []string {
	for i, v := range arr {
		if v == elem {
			return append(arr[:i], arr[i+1:]...)
		}
	}
	return arr
}
