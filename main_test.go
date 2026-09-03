package main

import "testing"

func TestGetVersion(t *testing.T) {
	orig := version
	defer func() { version = orig }()

	version = "1.2.3"
	if got := getVersion(); got != "1.2.3" {
		t.Fatalf("expected 1.2.3, got %s", got)
	}

	version = "dev"
	got := getVersion()
	if got == "" {
		t.Fatalf("expected non-empty version, got empty string")
	}
}
