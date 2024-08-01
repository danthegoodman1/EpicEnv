package cmd

import "testing"

func TestWrapQuotesIfNeeded(t *testing.T) {
	tests := []struct {
		Val          string
		ShouldChange bool
	}{
		{Val: "thing", ShouldChange: false},
		{Val: "hey who", ShouldChange: true},
		{Val: "\"hey who\"", ShouldChange: false},
	}

	for _, item := range tests {
		newVal := wrapQuotesIfNeeded(item.Val)
		if (newVal == item.Val) == item.ShouldChange {
			t.Fatalf("Failed on %s got %s", item.Val, newVal)
		}
	}
}
