package tui

import (
	"bytes"
	"strings"
	"testing"
)

func TestProgressWrite(t *testing.T) {
	widget := Progress(SetWidth(50))
	var buffer bytes.Buffer
	if _, err := widget.Write(&buffer, "pulling", 50); err != nil {
		t.Fatal(err)
	}
	out := buffer.String()

	if !strings.Contains(out, "pulling") {
		t.Fatalf("expected status in output, got %q", out)
	}
	if !strings.Contains(out, "50.0%") {
		t.Fatalf("expected percent in output, got %q", out)
	}
	if !strings.Contains(out, "█████") {
		t.Fatalf("expected filled bar in output, got %q", out)
	}
	if !strings.Contains(out, "░░░░░") {
		t.Fatalf("expected empty bar in output, got %q", out)
	}
}

func TestProgressWriteStatusOnly(t *testing.T) {
	widget := Progress(SetWidth(50))
	var buffer bytes.Buffer

	if _, err := widget.Write(&buffer, "starting", 0); err != nil {
		t.Fatal(err)
	}
	if got := buffer.String(); !strings.Contains(got, "starting") {
		t.Fatalf("expected status in output, got %q", got)
	}
	if got := buffer.String(); !strings.Contains(got, "0.0%") {
		t.Fatalf("expected percent in output, got %q", got)
	}
}

func TestProgressWriteLongStatusSingleLine(t *testing.T) {
	widget := Progress(SetWidth(50))
	var buffer bytes.Buffer

	if _, err := widget.Write(&buffer, "As_Iran_protests_continue,_is_US_intervention_likely_Plus_Myanmar", 25); err != nil {
		t.Fatal(err)
	}

	if got := buffer.String(); strings.Contains(got, "\n") || strings.Contains(got, "\r") {
		t.Fatalf("expected single-line output, got %q", got)
	}
}

func TestProgressWritePreservesLeadingPadding(t *testing.T) {
	widget := Progress(SetWidth(50))
	var buffer bytes.Buffer

	if _, err := widget.Write(&buffer, "  [ 41/272] Claude.dmg", 1); err != nil {
		t.Fatal(err)
	}

	if got := buffer.String(); !strings.HasPrefix(got, "  [ 41/272] Claude.dmg") {
		t.Fatalf("expected leading padding to be preserved, got %q", got)
	}
}
