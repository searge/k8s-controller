package cmd

import (
	"testing"
)

func TestServeCommandDefined(t *testing.T) {
	if serveCmd == nil {
		t.Fatal("serveCmd should be defined")
	}

	if serveCmd.Use != "serve" {
		t.Errorf("expected command use 'serve', got %s", serveCmd.Use)
	}

	portFlag := serveCmd.Flags().Lookup("port")
	if portFlag == nil {
		t.Error("expected 'port' flag to be defined")
	}
}
