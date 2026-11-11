package stages

import (
	"fmt"
	"strings"

	"spacereclaim/internal/model"
	"spacereclaim/internal/sys"
)

// AllStages returns the ordered catalog of all 6 disk reclamation stages.
func AllStages(runner sys.Runner) []model.Stage {
	return []model.Stage{
		NewWSLStage(runner),
		NewDockerStage(runner),
		NewDevCacheStage(runner),
		NewInstallersStage(runner),
		NewToolchainsStage(runner),
		NewSystemStage(runner),
	}
}

// SelectStages filters stages by a comma-separated list of IDs (or "all").
func SelectStages(runner sys.Runner, filter string) ([]model.Stage, error) {
	all := AllStages(runner)
	f := strings.ToLower(strings.TrimSpace(filter))
	if f == "" || f == "all" {
		return all, nil
	}

	byID := make(map[string]model.Stage, len(all))
	for _, st := range all {
		byID[st.ID()] = st
	}

	var selected []model.Stage
	seen := make(map[string]bool)
	for _, part := range strings.Split(f, ",") {
		id := strings.TrimSpace(part)
		if id == "" {
			continue
		}
		if id == "all" {
			return all, nil
		}
		st, ok := byID[id]
		if !ok {
			return nil, fmt.Errorf("unknown stage %q (valid stages: wsl, docker, devcache, installers, toolchains, system, all)", id)
		}
		if !seen[id] {
			seen[id] = true
			selected = append(selected, st)
		}
	}
	return selected, nil
}
