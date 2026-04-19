package main

import "testing"

func TestDefaultConfigPath(t *testing.T) {
	if defaultConfigPath != "./configs/plugin-market.yaml" {
		t.Fatalf("defaultConfigPath = %q, want %q", defaultConfigPath, "./configs/plugin-market.yaml")
	}
}
