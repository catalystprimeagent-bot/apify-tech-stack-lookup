package main

import (
	"net/http"
	"testing"
)

func TestDetectCDN_ByHeader(t *testing.T) {
	h := http.Header{}
	h.Set("CF-Ray", "abc123")
	if got := detectCDN(h, ""); got != "Cloudflare" {
		t.Errorf("got %q, want Cloudflare", got)
	}
}

func TestDetectCDN_ByCNAME(t *testing.T) {
	h := http.Header{}
	if got := detectCDN(h, "d111111abcdef8.cloudfront.net"); got != "Amazon CloudFront" {
		t.Errorf("got %q, want Amazon CloudFront", got)
	}
}

func TestDetectCDN_NoneFound(t *testing.T) {
	h := http.Header{}
	h.Set("Server", "nginx")
	if got := detectCDN(h, ""); got != "" {
		t.Errorf("got %q, want empty string for no CDN detected", got)
	}
}
