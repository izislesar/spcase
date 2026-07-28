package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestNormalizeSolutionURLAcceptsHTTPAndHTTPS(t *testing.T) {
	t.Parallel()

	for _, rawURL := range []string{
		"http://example.com/result",
		" https://example.com/result ",
	} {
		rawURL := rawURL
		t.Run(rawURL, func(t *testing.T) {
			t.Parallel()

			normalized, err := NormalizeSolutionURL(rawURL)
			if err != nil {
				t.Fatalf("NormalizeSolutionURL(%q): %v", rawURL, err)
			}
			if normalized != strings.TrimSpace(rawURL) {
				t.Fatalf("normalized URL = %q, want %q", normalized, strings.TrimSpace(rawURL))
			}
		})
	}
}

func TestNormalizeSolutionURLRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	overlongURL := "https://example.com/" +
		strings.Repeat("a", MaximumSolutionURLLength-len("https://example.com/")+1)
	testCases := []struct {
		name   string
		rawURL string
	}{
		{name: "unsupported scheme", rawURL: "ftp://example.com/result"},
		{name: "malformed URL", rawURL: "https://exa mple.com/result"},
		{name: "missing host", rawURL: "https:///result"},
		{name: "too long", rawURL: overlongURL},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if _, err := NormalizeSolutionURL(testCase.rawURL); !errors.Is(err, ErrInvalidURLFormat) {
				t.Fatalf("NormalizeSolutionURL() error = %v, want %v", err, ErrInvalidURLFormat)
			}
		})
	}
}

func TestNormalizeSolutionURLAcceptsMaximumLength(t *testing.T) {
	t.Parallel()

	const prefix = "https://example.com/"
	rawURL := prefix + strings.Repeat("a", MaximumSolutionURLLength-len(prefix))
	normalized, err := NormalizeSolutionURL(rawURL)
	if err != nil {
		t.Fatalf("NormalizeSolutionURL() at maximum length: %v", err)
	}
	if normalized != rawURL {
		t.Fatalf("normalized URL length = %d, want unchanged length %d", len(normalized), len(rawURL))
	}
}
