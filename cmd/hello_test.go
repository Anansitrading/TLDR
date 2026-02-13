package cmd

import (
	"bytes"
	"testing"
)

func TestHelloCommand(t *testing.T) {
	// Capture output
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	// Run the hello command
	rootCmd.SetArgs([]string{"hello"})
	err := rootCmd.Execute()

	if err != nil {
		t.Fatalf("hello command failed: %v", err)
	}

	output := buf.String()
	expected := "Hello World\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}
