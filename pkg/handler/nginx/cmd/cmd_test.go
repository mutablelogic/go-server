package nginx_test

import (
	"bytes"
	"testing"

	// Packages
	cmd "github.com/mutablelogic/go-server/pkg/handler/nginx/cmd"
)

func Test_Cmd_000(t *testing.T) {
	if cmd, err := cmd.New("ls"); err != nil {
		t.Error(err)
	} else {
		t.Log(cmd)
	}
}

func Test_Cmd_001(t *testing.T) {
	if cmd, err := cmd.New("ls"); err != nil {
		t.Error(err)
	} else if err := cmd.Run(); err != nil {
		t.Error(err)
	} else {
		t.Log(cmd)
	}
}

func Test_Cmd_002(t *testing.T) {
	cmd, err := cmd.New("ls")
	if err != nil {
		t.Error(err)
	}
	cmd.Out = func(data []byte) {
		t.Log(string(bytes.TrimSpace(data)))
	}
	if err := cmd.Run(); err != nil {
		t.Error(err)
	}
}

func Test_Cmd_003(t *testing.T) {
	cmd, err := cmd.New("ls")
	if err != nil {
		t.Error(err)
	}
	cmd.Out = func(data []byte) {
		t.Log(string(bytes.TrimSpace(data)))
	}
	if err := cmd.Run(); err != nil {
		t.Error(err)
	}
	if err := cmd.Run(); err != nil {
		t.Error(err)
	}
}
