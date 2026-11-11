package stages

import (
	"testing"

	"spacereclaim/internal/sys"
)

func TestSelectStages(t *testing.T) {
	r := sys.NewDryRunRunner()
	all, err := SelectStages(r, "all")
	if err != nil {
		t.Fatalf("SelectStages(all) failed: %v", err)
	}
	if len(all) != 6 {
		t.Fatalf("expected 6 stages, got %d", len(all))
	}

	sub, err := SelectStages(r, "wsl,docker,devcache")
	if err != nil {
		t.Fatalf("SelectStages(wsl,docker,devcache) failed: %v", err)
	}
	if len(sub) != 3 {
		t.Fatalf("expected 3 stages, got %d", len(sub))
	}

	if _, err := SelectStages(r, "nonexistent"); err == nil {
		t.Fatalf("expected error for unknown stage name")
	}
}
