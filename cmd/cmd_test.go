package cmd

import (
	"strings"
	"testing"
)

func TestReadInputValueFromPipePreservesSingleLineWithoutTrailingNewline(t *testing.T) {
	expected := strings.Repeat("A", 1700)

	got, err := readInputValue(strings.NewReader(expected), 0, false)
	if err != nil {
		t.Fatalf("readInputValue returned error: %v", err)
	}

	if got != expected {
		t.Fatalf("value mismatch: got length %d, want length %d", len(got), len(expected))
	}
}

func TestEnvfileValueRoundTripsSpecialCharacters(t *testing.T) {
	values := []string{
		"contains spaces",
		"contains 'single quotes'",
		`contains "double quotes"`,
		"contains `backticks`",
		"contains $dollars",
		"contains # hash",
		`contains \ backslashes`,
		"contains literal #personal marker",
		"single: ' backtick marker: ` #personal slash: \\",
		"contains 'single', \"double\", and `backticks` with \\ slash #personal",
	}

	for _, want := range values {
		_, got, personal, ok := parseEnvfileLine("KEY=" + formatEnvfileValue(want))
		if !ok {
			t.Fatalf("parseEnvfileLine did not parse formatted value %q", want)
		}
		if personal {
			t.Fatalf("formatted shared value %q was marked personal", want)
		}
		if got != want {
			t.Fatalf("value mismatch: got %q, want %q", got, want)
		}
	}
}

func TestFormatEnvfileValueAvoidsBacktickDelimiterWhenValueContainsBacktick(t *testing.T) {
	value := "single: ' backtick marker: ` #personal slash: \\"
	got := formatEnvfileValue(value)
	want := `"` + value + `"`

	if got != want {
		t.Fatalf("formatEnvfileValue() = %q, want %q", got, want)
	}
}

func TestParseEnvfileLineDistinguishesPersonalAnnotation(t *testing.T) {
	_, got, personal, ok := parseEnvfileLine("KEY='foo #personal'")
	if !ok {
		t.Fatal("parseEnvfileLine did not parse quoted shared value")
	}
	if personal {
		t.Fatal("literal #personal inside quotes was marked personal")
	}
	if got != "foo #personal" {
		t.Fatalf("value mismatch: got %q", got)
	}

	_, got, personal, ok = parseEnvfileLine("KEY='foo #personal' #personal")
	if !ok {
		t.Fatal("parseEnvfileLine did not parse personal value")
	}
	if !personal {
		t.Fatal("trailing #personal annotation was not detected")
	}
	if got != "foo #personal" {
		t.Fatalf("value mismatch: got %q", got)
	}
}
