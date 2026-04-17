package main

import "testing"

func TestDefaultConfigPath(t *testing.T) {
	if defaultConfigPath != "./configs/config.yaml" {
		t.Fatalf("defaultConfigPath = %q, want %q", defaultConfigPath, "./configs/config.yaml")
	}
}
