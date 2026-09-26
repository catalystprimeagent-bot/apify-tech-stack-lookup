package main

import (
	"net/http"
	"testing"

	wappalyzer "github.com/projectdiscovery/wappalyzergo"
)

const wordpressHTML = `<html><head><meta name="generator" content="WordPress 6.4"></head><body></body></html>`

func mustClient(t *testing.T) *wappalyzer.Wappalyze {
	t.Helper()
	c, err := wappalyzer.New()
	if err != nil {
		t.Fatalf("wappalyzer.New: %v", err)
	}
	return c
}

func TestDetectTechnologies_IncludeVersions(t *testing.T) {
	client := mustClient(t)
	headers := http.Header{}
	techs, names := detectTechnologies(client, headers, []byte(wordpressHTML), true, false, nil)
	if len(techs) == 0 {
		t.Fatal("expected at least one detected technology (WordPress)")
	}
	found := false
	for _, n := range names {
		if n == "WordPress" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected WordPress in detected names, got %v", names)
	}
}

func TestDetectTechnologies_IncludeConfidence(t *testing.T) {
	client := mustClient(t)
	headers := http.Header{}
	techs, _ := detectTechnologies(client, headers, []byte(wordpressHTML), true, true, nil)
	for _, tech := range techs {
		if tech.Confidence == 0 {
			t.Errorf("tech %q: expected non-zero confidence when include_confidence is set", tech.Name)
		}
	}
}

func TestDetectTechnologies_ExcludeVersions(t *testing.T) {
	client := mustClient(t)
	headers := http.Header{}
	techs, _ := detectTechnologies(client, headers, []byte(wordpressHTML), false, false, nil)
	for _, tech := range techs {
		if tech.Version != "" {
			t.Errorf("tech %q: expected empty version when include_versions is false, got %q", tech.Name, tech.Version)
		}
		if tech.Confidence != 0 {
			t.Errorf("tech %q: expected zero confidence when include_confidence is false, got %d", tech.Name, tech.Confidence)
		}
	}
}

func TestDetectTechnologies_CategoryFilter(t *testing.T) {
	client := mustClient(t)
	headers := http.Header{}
	filter := map[string]bool{"cms": true}
	techs, _ := detectTechnologies(client, headers, []byte(wordpressHTML), true, false, filter)
	if len(techs) == 0 {
		t.Fatal("expected WordPress to match the 'cms' category filter")
	}
	for _, tech := range techs {
		matched := false
		for _, cat := range tech.Categories {
			if filter[lower(cat)] {
				matched = true
			}
		}
		if !matched {
			t.Errorf("tech %q leaked through category filter with categories %v", tech.Name, tech.Categories)
		}
	}
}

func TestDetectTechnologies_CategoryFilterExcludesEverything(t *testing.T) {
	client := mustClient(t)
	headers := http.Header{}
	filter := map[string]bool{"no-such-category-xyz": true}
	techs, names := detectTechnologies(client, headers, []byte(wordpressHTML), true, true, filter)
	if len(techs) != 0 || len(names) != 0 {
		t.Errorf("expected no matches for a nonexistent category, got %v", names)
	}
}

func lower(s string) string {
	out := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		out[i] = c
	}
	return string(out)
}
