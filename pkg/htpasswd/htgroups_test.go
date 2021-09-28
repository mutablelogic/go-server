package htpasswd_test

import (
	"bytes"
	"os"
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-server/pkg/htpasswd"
)

const (
	FILE_GROUPS = "../../etc/htgroups"
)

func Test_Groups_000(t *testing.T) {
	groups := NewGroups()
	if groups == nil {
		t.Errorf("Expected non-nil htgroups")
	}
}

func Test_Groups_001(t *testing.T) {
	r, err := os.Open(FILE_GROUPS)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	groups, err := ReadGroups(r)
	if err != nil {
		t.Fatal(err)
	}

	if err := groups.AddUserToGroup("test", "test"); err != nil {
		t.Error(err)
	} else if groups.UserInGroup("test", "test") != true {
		t.Error("Expected user to be in group")
	}
	var buf bytes.Buffer
	if err := groups.Write(&buf); err != nil {
		t.Error(err)
	} else {
		t.Log(buf.String())
	}
}

func Test_Groups_002(t *testing.T) {
	groups := NewGroups()
	if err := groups.AddUserToGroup("test", "test"); err != nil {
		t.Error(err)
	} else if groups.UserInGroup("test", "test") != true {
		t.Error("Expected user to be in group")
	}
	if err := groups.AddUserToGroup("test", "test2"); err != nil {
		t.Error(err)
	} else if groups.UserInGroup("test", "test2") != true {
		t.Error("Expected user to be in group")
	}
	if err := groups.RemoveUserFromGroup("test", "test"); err != nil {
		t.Error(err)
	} else if groups.UserInGroup("test", "test") != false {
		t.Error("Expected user to NOT be in group")
	}
	if err := groups.RemoveUserFromGroup("test", "test2"); err != nil {
		t.Error(err)
	} else if groups.UserInGroup("test", "test2") != false {
		t.Error("Expected user to NOT be in group")
	}

	var buf bytes.Buffer
	if err := groups.Write(&buf); err != nil {
		t.Error(err)
	} else {
		t.Log(buf.String())
	}
}

func Test_Groups_003(t *testing.T) {
	groups := NewGroups()
	if err := groups.AddUserToGroup("a", "test"); err != nil {
		t.Error(err)
	} else if groups.UserInGroup("a", "test") != true {
		t.Error("Expected user to be in group")
	}
	if err := groups.AddUserToGroup("b", "test"); err != nil {
		t.Error(err)
	} else if groups.UserInGroup("b", "test") != true {
		t.Error("Expected user to be in group")
	}
	if err := groups.RemoveUserFromGroup("a", "test"); err != nil {
		t.Error(err)
	} else if groups.UserInGroup("a", "test") != false {
		t.Error("Expected user to NOT be in group")
	}
	if err := groups.RemoveUserFromGroup("b", "test"); err != nil {
		t.Error(err)
	} else if groups.UserInGroup("b", "test") != false {
		t.Error("Expected user to NOT be in group")
	}

	var buf bytes.Buffer
	if err := groups.Write(&buf); err != nil {
		t.Error(err)
	} else {
		t.Log(buf.String())
	}
}
