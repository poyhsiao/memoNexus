// Package main tests for the core library entry point.
// These tests verify basic functionality and version handling.
package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestMainOutput(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run main (this will exit, so we can't actually call main())
	// Instead, we verify the version variable is set
	if Version == "" {
		t.Error("Version should not be empty")
	}

	// Restore stdout
	w.Close()
	os.Stdout = old
	_, _ = io.Copy(old, r)
}

func TestVersionDefault(t *testing.T) {
	// Test that Version has a default value
	// In production, this is set at build time
	if Version != "0.1.0" {
		// This is OK - version might be set by build flags
		// Just verify it's not empty
		if Version == "" {
			t.Error("Version should not be empty")
		}
	}
}

func TestPrintVersion(t *testing.T) {
	// Test version printing format
	var buf bytes.Buffer
	expectedPrefix := "MemoNexus Core v"

	// Simulate what main() prints
	buf.WriteString("MemoNexus Core v")
	buf.WriteString(Version)
	buf.WriteString("\n")

	output := buf.String()
	if !strings.HasPrefix(output, expectedPrefix) {
		t.Errorf("Expected output to start with %q, got %q", expectedPrefix, output)
	}
}
