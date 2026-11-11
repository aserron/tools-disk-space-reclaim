package ui

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"spacereclaim/internal/model"
)

// SelectorOutcome represents the user's decision in the interactive TUI selector.
type SelectorOutcome struct {
	Proceed bool
	DryRun  bool
}

// RunInteractiveSelector lets the user toggle individual items, stages, or risk presets before execution.
func RunInteractiveSelector(in io.Reader, results []*model.StageScanResult, indexed []IndexedItem, isAdmin bool, initialDryRun bool) SelectorOutcome {
	reader := bufio.NewReader(in)

	for {
		fmt.Printf("\n%sInteractive Selection Commands:%s\n", model.ANSIBold, model.ANSIReset)
		fmt.Printf("  %s[1..%d]%s      Toggle item(s) by number or range (e.g. %s1,3,5-8%s)\n",
			model.ANSICyan, len(indexed), model.ANSIReset, model.ANSIBold, model.ANSIReset)
		fmt.Printf("  %s[stage:<id>]%s Toggle all items in a stage (e.g. %sstage:wsl%s, %sstage:docker%s, %sstage:devcache%s)\n",
			model.ANSICyan, model.ANSIReset, model.ANSIBold, model.ANSIReset, model.ANSIBold, model.ANSIReset, model.ANSIBold, model.ANSIReset)
		fmt.Printf("  %s[s]%s Safe preset only   %s[m]%s Safe + Moderate   %s[a]%s Select All   %s[n]%s Select None\n",
			model.ANSIGreen, model.ANSIReset,
			model.ANSIYellow, model.ANSIReset,
			model.ANSIRed, model.ANSIReset,
			model.ANSICyan, model.ANSIReset)
		fmt.Printf("  %s[l]%s Re-list table      %s[d]%s Run in Dry-Run    %s[r]%s Run Selected %s[q]%s Quit without changes\n",
			model.ANSICyan, model.ANSIReset,
			model.ANSICyan, model.ANSIReset,
			model.ANSIGreen+model.ANSIBold, model.ANSIReset,
			model.ANSIRed, model.ANSIReset)
		fmt.Printf("\n%sSelect action > %s", model.ANSIBold, model.ANSIReset)

		line, err := reader.ReadString('\n')
		if err != nil && len(line) == 0 {
			return SelectorOutcome{Proceed: false}
		}
		cmd := strings.TrimSpace(line)
		lower := strings.ToLower(cmd)

		switch lower {
		case "q", "quit", "exit":
			fmt.Println("Aborted by user. No changes were made.")
			return SelectorOutcome{Proceed: false}

		case "r", "run", "y", "yes":
			count := countSelected(indexed)
			if count == 0 {
				fmt.Printf("%sNo items are currently selected. Toggle items first or press 'q' to quit.%s\n", model.ANSIYellow, model.ANSIReset)
				continue
			}
			return SelectorOutcome{Proceed: true, DryRun: initialDryRun}

		case "d", "dry", "dry-run":
			count := countSelected(indexed)
			if count == 0 {
				fmt.Printf("%sNo items are currently selected.%s\n", model.ANSIYellow, model.ANSIReset)
				continue
			}
			return SelectorOutcome{Proceed: true, DryRun: true}

		case "l", "list":
			indexed = PrintScanReport(results, isAdmin)

		case "s", "safe":
			for _, entry := range indexed {
				entry.Item.Selected = (entry.Item.Risk == model.RiskSafe && entry.Item.SizeBytes > 0) || entry.Item.ID == "temp-recycle-bin" || strings.HasPrefix(entry.Item.ID, "wsl-clean-")
			}
			indexed = PrintScanReport(results, isAdmin)

		case "m", "moderate":
			for _, entry := range indexed {
				entry.Item.Selected = entry.Item.Risk <= model.RiskModerate
			}
			indexed = PrintScanReport(results, isAdmin)

		case "a", "all":
			for _, entry := range indexed {
				entry.Item.Selected = true
			}
			indexed = PrintScanReport(results, isAdmin)

		case "n", "none", "clear":
			for _, entry := range indexed {
				entry.Item.Selected = false
			}
			indexed = PrintScanReport(results, isAdmin)

		default:
			if strings.HasPrefix(lower, "stage:") {
				stID := strings.TrimPrefix(lower, "stage:")
				toggled := 0
				// If any item in stage is unselected, select all; otherwise unselect all
				anyUnselected := false
				for _, entry := range indexed {
					if entry.Item.StageID == stID && !entry.Item.Selected {
						anyUnselected = true
						break
					}
				}
				for _, entry := range indexed {
					if entry.Item.StageID == stID {
						entry.Item.Selected = anyUnselected
						toggled++
					}
				}
				if toggled == 0 {
					fmt.Printf("%sUnknown or empty stage %q.%s\n", model.ANSIYellow, stID, model.ANSIReset)
				} else {
					indexed = PrintScanReport(results, isAdmin)
				}
				continue
			}

			if toggleNumberSpec(cmd, indexed) {
				indexed = PrintScanReport(results, isAdmin)
			} else {
				fmt.Printf("%sUnrecognized command %q. Enter item numbers (e.g. 1,3-5), 's', 'm', 'a', 'n', 'd', 'r', or 'q'.%s\n",
					model.ANSIYellow, cmd, model.ANSIReset)
			}
		}
	}
}

func countSelected(indexed []IndexedItem) int {
	n := 0
	for _, e := range indexed {
		if e.Item.Selected {
			n++
		}
	}
	return n
}

// toggleNumberSpec parses comma-separated indices or ranges like "1,3,5-8" and flips their Selected state.
func toggleNumberSpec(spec string, indexed []IndexedItem) bool {
	parts := strings.Split(spec, ",")
	toggledAny := false

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.Contains(p, "-") {
			bounds := strings.SplitN(p, "-", 2)
			start, err1 := strconv.Atoi(strings.TrimSpace(bounds[0]))
			end, err2 := strconv.Atoi(strings.TrimSpace(bounds[1]))
			if err1 != nil || err2 != nil || start > end {
				return false
			}
			for i := start; i <= end; i++ {
				if i >= 1 && i <= len(indexed) {
					indexed[i-1].Item.Selected = !indexed[i-1].Item.Selected
					toggledAny = true
				}
			}
		} else {
			idx, err := strconv.Atoi(p)
			if err != nil || idx < 1 || idx > len(indexed) {
				return false
			}
			indexed[idx-1].Item.Selected = !indexed[idx-1].Item.Selected
			toggledAny = true
		}
	}
	return toggledAny
}
