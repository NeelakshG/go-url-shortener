package shortcode

import (
	"strings"
	"testing"
)

func TestGenerate_Length(t *testing.T) {
	code, err := Generate()
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	if len(code) != 6 {
		t.Fatalf("expected length 6, got %d", len(code))
	}
}

func TestGenerate_Base62Characters(t *testing.T) {
	for i := 0; i < 1000; i++ {
		code, err := Generate()
		if err != nil {
			t.Fatalf("Generate() returned error: %v", err)
		}

		for _, r := range code {
			if !strings.ContainsRune(alphabet, r) {
				t.Fatalf("invalid character %q in code %q", r, code)
			}
		}
	}
}

func TestGenerate_Unique(t *testing.T) {
	seen := make(map[string]bool)

	for i := 0; i < 1000; i++ {
		code, err := Generate()
		if err != nil {
			t.Fatalf("Generate() returned error: %v", err)
		}

		if seen[code] {
			t.Fatalf("duplicate code generated: %s", code)
		}

		seen[code] = true
	}
}