package main

import "testing"

func TestDomainsFromInput_AllAliases(t *testing.T) {
	in := Input{
		Urls:       []interface{}{"https://a.example"},
		StartUrls:  []interface{}{map[string]interface{}{"url": "b.example"}},
		Domains:    "c.example",
		Websites:   "d.example, e.example",
		TargetUrls: "f.example\ng.example",
		URL:        "h.example",
	}
	entries, rejected := domainsFromInput(in)

	if len(rejected) != 0 {
		t.Fatalf("unexpected rejects: %v", rejected)
	}
	want := []string{"a.example", "b.example", "c.example", "d.example", "e.example", "f.example", "g.example", "h.example"}
	if len(entries) != len(want) {
		t.Fatalf("got %d entries, want %d: %+v", len(entries), len(want), entries)
	}
	for i, w := range want {
		if entries[i].host != w {
			t.Errorf("entry %d: got host %q, want %q", i, entries[i].host, w)
		}
	}
}

func TestDomainsFromInput_Dedup(t *testing.T) {
	in := Input{Urls: []interface{}{"example.com", "https://example.com", "EXAMPLE.com/"}}
	entries, _ := domainsFromInput(in)
	if len(entries) != 1 {
		t.Fatalf("expected dedup to 1 entry, got %d: %+v", len(entries), entries)
	}
}

func TestDomainsFromInput_RejectsUnparseable(t *testing.T) {
	in := Input{Urls: []interface{}{"://not-a-url", "good.example"}}
	entries, rejected := domainsFromInput(in)
	if len(entries) != 1 || entries[0].host != "good.example" {
		t.Fatalf("expected only good.example to survive, got %+v", entries)
	}
	if len(rejected) != 1 {
		t.Fatalf("expected 1 rejected entry, got %v", rejected)
	}
}

func TestDomainsFromInput_Empty(t *testing.T) {
	entries, rejected := domainsFromInput(Input{})
	if len(entries) != 0 || len(rejected) != 0 {
		t.Fatalf("expected nothing from empty input, got entries=%+v rejected=%v", entries, rejected)
	}
}

func TestNormalizeURL_AddsScheme(t *testing.T) {
	u, host, err := normalizeURL("stripe.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u != "https://stripe.com" {
		t.Errorf("got url %q, want https://stripe.com", u)
	}
	if host != "stripe.com" {
		t.Errorf("got host %q, want stripe.com", host)
	}
}

func TestNormalizeURL_PreservesExplicitScheme(t *testing.T) {
	u, _, err := normalizeURL("http://stripe.com/pricing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u != "http://stripe.com/pricing" {
		t.Errorf("got %q, want scheme preserved", u)
	}
}
