package cmd

import "testing"

func TestWrapQuotesIfNeeded(t *testing.T) {
	tests := []struct {
		Val  string
		Want string
	}{
		{Val: "thing", Want: "thing"},
		{Val: "hey who", Want: "'hey who'"},
		{Val: "\"hey who\"", Want: "'\"hey who\"'"},
		{Val: "can't", Want: "'can'\\''t'"},
		{Val: "$TOKEN", Want: "'$TOKEN'"},
	}

	for _, item := range tests {
		got := wrapQuotesIfNeeded(item.Val)
		if got != item.Want {
			t.Fatalf("wrapQuotesIfNeeded(%q) = %q, want %q", item.Val, got, item.Want)
		}
	}
}
