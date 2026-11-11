package sys

import (
	"context"
	"testing"
)

func TestDecodeUTF16IfNeeded(t *testing.T) {
	// "Ubuntu" in UTF-16LE
	raw := []byte{'U', 0, 'b', 0, 'u', 0, 'n', 0, 't', 0, 'u', 0}
	got := DecodeUTF16IfNeeded(raw)
	if got != "Ubuntu" {
		t.Fatalf("DecodeUTF16IfNeeded() = %q, want %q", got, "Ubuntu")
	}

	// Plain UTF-8
	utf8Raw := []byte("Docker-Desktop")
	if got := DecodeUTF16IfNeeded(utf8Raw); got != "Docker-Desktop" {
		t.Fatalf("DecodeUTF16IfNeeded(utf8) = %q, want Docker-Desktop", got)
	}
}

func TestDryRunRunner(t *testing.T) {
	dr := NewDryRunRunner()
	_, _ = dr.RunWSL(context.Background(), "Ubuntu", "root", "fstrim -v /")
	_, _ = dr.RunWSLHost(context.Background(), "--shutdown")

	if len(dr.Recorded) != 2 {
		t.Fatalf("expected 2 recorded commands, got %d", len(dr.Recorded))
	}
}
