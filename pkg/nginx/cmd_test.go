package nginx_test

import (
	"bytes"
	"testing"

	"github.com/mutablelogic/go-server/pkg/nginx"
)

func Test_Cmd_000(t *testing.T) {
	if cmd, err := nginx.NewWithCommand("ls"); err != nil {
		t.Error(err)
	} else {
		t.Log(cmd)
	}
}

func Test_Cmd_001(t *testing.T) {
	if cmd, err := nginx.NewWithCommand("ls"); err != nil {
		t.Error(err)
	} else if err := cmd.Run(); err != nil {
		t.Error(err)
	} else {
		t.Log(cmd)
	}
}

func Test_Cmd_002(t *testing.T) {
	cmd, err := nginx.NewWithCommand("ls")
	if err != nil {
		t.Error(err)
	}
	cmd.Out = func(cmd *nginx.Cmd, data []byte) {
		t.Log(string(bytes.TrimSpace(data)))
	}
	if err := cmd.Run(); err != nil {
		t.Error(err)
	}
}
