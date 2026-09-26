package main

import "testing"

func TestClampInt_TimeoutSecondsBounds(t *testing.T) {
	cases := []struct{ in, want int }{
		{0, minTimeoutSeconds},
		{2, minTimeoutSeconds},
		{15, 15},
		{60, maxTimeoutSeconds},
		{999, maxTimeoutSeconds},
	}
	for _, c := range cases {
		if got := clampInt(c.in, minTimeoutSeconds, maxTimeoutSeconds); got != c.want {
			t.Errorf("clampInt(%d): got %d, want %d", c.in, got, c.want)
		}
	}
}

func TestIntOrDefault(t *testing.T) {
	if got := intOrDefault(nil, defaultTimeoutSeconds); got != defaultTimeoutSeconds {
		t.Errorf("nil timeout_seconds: got %d, want default %d", got, defaultTimeoutSeconds)
	}
	v := 42
	if got := intOrDefault(&v, defaultTimeoutSeconds); got != 42 {
		t.Errorf("explicit timeout_seconds: got %d, want 42", got)
	}
}

func TestBoolOrDefault(t *testing.T) {
	if !boolOrDefault(nil, true) {
		t.Error("nil include_versions/include_confidence should default to true")
	}
	f := false
	if boolOrDefault(&f, true) {
		t.Error("explicit false should override default true")
	}
}

func TestToLowerSet(t *testing.T) {
	set := toLowerSet([]string{"CMS", "Analytics"})
	if !set["cms"] || !set["analytics"] {
		t.Errorf("expected lowercased keys, got %v", set)
	}
	if toLowerSet(nil) != nil {
		t.Error("expected nil categories_filter to produce a nil set (no filtering)")
	}
}
