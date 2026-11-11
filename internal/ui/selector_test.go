package ui

import (
	"testing"

	"spacereclaim/internal/model"
)

func TestToggleNumberSpec(t *testing.T) {
	items := []IndexedItem{
		{Index: 1, Item: &model.TargetItem{ID: "1", Selected: false}},
		{Index: 2, Item: &model.TargetItem{ID: "2", Selected: false}},
		{Index: 3, Item: &model.TargetItem{ID: "3", Selected: false}},
		{Index: 4, Item: &model.TargetItem{ID: "4", Selected: false}},
	}

	if !toggleNumberSpec("1,3-4", items) {
		t.Fatalf("expected toggleNumberSpec(1,3-4) to succeed")
	}
	if !items[0].Item.Selected || items[1].Item.Selected || !items[2].Item.Selected || !items[3].Item.Selected {
		t.Fatalf("unexpected selection state after toggle: %+v", items)
	}
}
