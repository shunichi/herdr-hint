package main

import "testing"

func TestHerdrBinDefault(t *testing.T) {
	t.Setenv("HERDR_BIN_PATH", "") // unset/empty falls back to PATH lookup
	if got := herdrBin(); got != "herdr" {
		t.Fatalf("herdrBin() = %q, want %q", got, "herdr")
	}
}

func TestHerdrBinFromEnv(t *testing.T) {
	t.Setenv("HERDR_BIN_PATH", "/custom/herdr")
	if got := herdrBin(); got != "/custom/herdr" {
		t.Fatalf("herdrBin() = %q, want %q", got, "/custom/herdr")
	}
}
