package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	// Packages
	"github.com/mutablelogic/go-server/pkg/htpasswd"
	"golang.org/x/crypto/ssh/terminal"
)

/////////////////////////////////////////////////////////////////////
// TYPES

type Command struct {
	Name        string
	Description string
	Command     func(w io.Writer, args []string) error
}

type App struct {
	Dirty    bool
	Passwd   *htpasswd.Htpasswd
	Group    *htpasswd.Htgroups
	Commands []Command
}

/////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	flagPasswd = "passwd"
	flagGroup  = "group"
	flagCreate = "create"
)

/////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewApp() *App {
	app := new(App)
	app.Passwd = htpasswd.New()
	app.Group = htpasswd.NewGroups()

	// Return success
	return app
}

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (app *App) ListUsers(w io.Writer) {
	users := app.Passwd.Users()
	if len(users) == 0 {
		fmt.Fprintln(w, "No users, use \"htpasswd -create -passwd <file> add <user>\" to create a new password file")
	} else {
		fmt.Fprintf(w, "users=%q\n", users)
	}
}

func (app *App) ListGroups(w io.Writer) {
	for _, group := range app.Group.Groups() {
		fmt.Fprintf(w, "%v=%q\n", "@"+group, app.Group.UsersForGroup(group))
	}
}

func (app *App) Run(w io.Writer, args []string) error {
	for _, cmd := range app.Commands {
		if cmd.Name == args[0] && cmd.Command != nil {
			return cmd.Command(w, args[1:])
		}
	}
	return fmt.Errorf("unknown command %q", args[0])
}

func (app *App) Add(w io.Writer, args []string) error {
	if len(args) == 0 {
		return errors.New("add requires a user or @group")
	}

	// Requires the user does not exist
	if app.Passwd.Exists(args[0]) {
		return fmt.Errorf("user %q already exists", args[0])
	}

	// Set password
	fmt.Fprint(w, "New password: ")
	passwd, err := Password()
	fmt.Fprintln(w, "")
	if err != nil {
		return err
	} else if passwd == "" {
		return fmt.Errorf("empty password")
	}

	fmt.Fprint(w, "Repeat password: ")
	passwd2, err := Password()
	fmt.Fprintln(w, "")
	if err != nil {
		return err
	} else if passwd2 != passwd {
		return fmt.Errorf("passwords do not match")
	}

	if err := app.Passwd.Set(args[0], passwd, htpasswd.BCrypt); err != nil {
		return err
	}

	app.Dirty = true
	app.ListUsers(w)
	return nil
}

func (app *App) Update(w io.Writer, args []string) error {
	if len(args) == 0 {
		return errors.New("add requires a user")
	}

	// Requires the user does not exist
	if !app.Passwd.Exists(args[0]) {
		return fmt.Errorf("user %q does not exist", args[0])
	}

	// Set password
	fmt.Fprint(w, "New password: ")
	passwd, err := Password()
	fmt.Fprintln(w, "")
	if err != nil {
		return err
	} else if passwd == "" {
		return fmt.Errorf("empty password")
	}

	fmt.Fprint(w, "Repeat password: ")
	passwd2, err := Password()
	fmt.Fprintln(w, "")
	if err != nil {
		return err
	} else if passwd2 != passwd {
		return fmt.Errorf("passwords do not match")
	}

	if err := app.Passwd.Set(args[0], passwd, htpasswd.BCrypt); err != nil {
		return err
	}

	fmt.Fprintln(w, "password updated")
	app.Dirty = true
	return nil
}

func (app *App) Verify(w io.Writer, args []string) error {
	if len(args) == 0 {
		return errors.New("verify requires a user")
	}

	// Requires the user does not exist
	if !app.Passwd.Exists(args[0]) {
		return fmt.Errorf("user %q does not exist", args[0])
	}

	// Get password
	fmt.Fprint(w, "Verify password: ")
	passwd, err := Password()
	fmt.Fprintln(w, "")
	if err != nil {
		return err
	} else if passwd == "" {
		return fmt.Errorf("empty password")
	}

	if ok := app.Passwd.Verify(args[0], passwd); !ok {
		fmt.Fprintln(w, "password does not match")
	} else {
		fmt.Fprintln(w, "password matches")
	}

	return nil
}

func (app *App) Remove(w io.Writer, args []string) error {
	if len(args) == 0 {
		return errors.New("verify requires a user or @group")
	}

	// Requires the user exists
	if !app.Passwd.Exists(args[0]) {
		return fmt.Errorf("user %q does not exist", args[0])
	}

	// Remove user from all groups
	for _, group := range app.Group.GroupsForUser(args[0]) {
		app.Group.RemoveUserFromGroup(args[0], group)
	}

	// Remove user
	app.Passwd.Delete(args[0])
	app.Dirty = true

	// List users and groups
	app.ListUsers(w)
	app.ListGroups(w)
	return nil
}

/////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func Password() (string, error) {
	password, err := terminal.ReadPassword(0)
	if err != nil {
		return "", err
	} else {
		return strings.TrimSpace(string(password)), nil
	}
}

/////////////////////////////////////////////////////////////////////
// MAIN

func main() {
	// New app
	app := NewApp()
	app.Commands = append(app.Commands, Command{
		Name:        "add",
		Description: "add a user, or a user to a @group",
		Command:     app.Add,
	})
	app.Commands = append(app.Commands, Command{
		Name:        "remove",
		Description: "remove a user, or a user from a @group",
		Command:     app.Remove,
	})
	app.Commands = append(app.Commands, Command{
		Name:        "verify",
		Description: "verify a user password",
		Command:     app.Verify,
	})
	app.Commands = append(app.Commands, Command{
		Name:        "update",
		Description: "update a user password",
		Command:     app.Update,
	})

	// Create flags
	flags := flag.NewFlagSet(filepath.Base(os.Args[0]), flag.ContinueOnError)
	flags.String(flagPasswd, "", "password file")
	flags.String(flagGroup, "", "group file")
	flags.Bool(flagCreate, false, "create password and/or group files if they don't exist")

	// Set usage function
	flags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "\t%s <flags> <command> <user|@group> (<members>...)\n", flags.Name())
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flags.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nCommands:\n")
		for _, command := range app.Commands {
			fmt.Fprintf(os.Stderr, "  %s\n\t%s\n", command.Name, command.Description)
		}
	}

	// Parse flags
	if err := flags.Parse(os.Args[1:]); errors.Is(err, flag.ErrHelp) {
		os.Exit(0)
	} else if err != nil {
		fmt.Fprintln(flags.Output(), err)
		os.Exit(-1)
	}

	// Read password file
	if passwd := flags.Lookup(flagPasswd).Value.String(); passwd != "" {
		r, err := os.Open(passwd)
		if err != nil {
			fmt.Fprintln(flags.Output(), err)
			os.Exit(-1)
		}
		defer r.Close()
		if passwd, err := htpasswd.Read(r); err != nil {
			fmt.Fprintln(flags.Output(), err)
			os.Exit(-1)
		} else {
			app.Passwd = passwd
		}
	}

	// Read groups file
	if groups := flags.Lookup(flagGroup).Value.String(); groups != "" {
		r, err := os.Open(groups)
		if err != nil {
			fmt.Fprintln(flags.Output(), err)
			os.Exit(-1)
		}
		defer r.Close()
		if groups, err := htpasswd.ReadGroups(r); err != nil {
			fmt.Fprintln(flags.Output(), err)
			os.Exit(-1)
		} else {
			app.Group = groups
		}
	}

	// By default, list the users and groups
	if len(flags.Args()) == 0 {
		app.ListUsers(os.Stdout)
		app.ListGroups(os.Stdout)
	} else if err := app.Run(os.Stdout, flags.Args()); err != nil {
		fmt.Fprintln(flags.Output(), err)
		os.Exit(-1)
	}

	// If dirty, then write password and group files
	if app.Dirty {
		if err := app.Passwd.Write(os.Stdout); err != nil {
			fmt.Fprintln(flags.Output(), err)
			os.Exit(-1)
		}
	}
}
