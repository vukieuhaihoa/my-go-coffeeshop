package catalog

import (
	"slices"
	"testing"
)

func TestAll(t *testing.T) {
	catalog := New()

	itemTypes := catalog.All()
	if len(itemTypes) == 0 {
		t.Fatal("expected non-empty item types, got empty")
	}

	if itemTypes[0].Name != "CAPPUCCINO" {
		t.Errorf("expected item type name 'CAPPUCCINO', got '%s'", itemTypes[0].Name)
	}
}

func TestGetByNames(t *testing.T) {
	catalog := New()

	tests := []struct {
		name        string
		input       []string
		wantCount   int
		wantUnknown []string
	}{
		{
			name:        "valid types",
			input:       []string{"CAPPUCCINO", "ESPRESSO"},
			wantCount:   2,
			wantUnknown: nil,
		},
		{
			name:        "unknown type",
			input:       []string{"PHO"},
			wantCount:   0,
			wantUnknown: []string{"PHO"},
		},
		{
			name:        "mixed known and unknown types",
			input:       []string{"CAPPUCCINO", "PHO"},
			wantCount:   1,
			wantUnknown: []string{"PHO"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, unknown := catalog.GetByNames(tt.input)
			if len(items) != tt.wantCount {
				t.Errorf("expected %d items, got %d", tt.wantCount, len(items))
			}
			if !slices.Equal(unknown, tt.wantUnknown) {
				t.Errorf("expected unknown types %v, got %v", tt.wantUnknown, unknown)
			}
		})
	}
}

func TestAllReturnsCopy(t *testing.T) {
	catalog := New()

	itemTypes1 := catalog.All()
	itemTypes2 := catalog.All()

	if &itemTypes1[0] == &itemTypes2[0] {
		t.Error("expected All() to return a copy, but got the same reference")
	}

	itemTypes1[0].Name = "CHANGED"
	if catalog.All()[0].Name != "CAPPUCCINO" {
		t.Errorf("expected item type name 'CAPPUCCINO', got '%s'", catalog.All()[0].Name)
	}
}
