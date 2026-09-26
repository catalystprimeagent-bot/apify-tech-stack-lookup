package main

import (
	"net/http"
	"strings"
)

// cdnSignature describes one way to recognize a CDN/hosting-edge provider: a set of response
// header substrings (any one hit is enough) and/or CNAME target suffixes.
type cdnSignature struct {
	name          string
	headerHints   map[string][]string // header name (lowercase) -> substrings, any match wins
	headerPresent []string            // header names whose mere presence is a signal
	cnameSuffixes []string
	serverSubstr  []string
}

var cdnSignatures = []cdnSignature{
	{
		name:          "Cloudflare",
		headerPresent: []string{"cf-ray", "cf-cache-status"},
		serverSubstr:  []string{"cloudflare"},
		cnameSuffixes: []string{".cloudflare.net"},
	},
	{
		name:          "Amazon CloudFront",
		headerPresent: []string{"x-amz-cf-id", "x-amz-cf-pop"},
		cnameSuffixes: []string{".cloudfront.net"},
	},
	{
		name:          "Fastly",
		headerPresent: []string{"x-fastly-request-id"},
		headerHints:   map[string][]string{"via": {"fastly"}, "x-served-by": {"cache-"}},
		cnameSuffixes: []string{".fastly.net", ".fastlylb.net"},
	},
	{
		name:          "Akamai",
		headerPresent: []string{"x-akamai-transformed"},
		headerHints:   map[string][]string{"server": {"akamaighost"}},
		cnameSuffixes: []string{".akamaiedge.net", ".akamai.net", ".edgekey.net", ".edgesuite.net"},
	},
	{
		name:          "Microsoft Azure Front Door / CDN",
		headerPresent: []string{"x-azure-ref"},
		cnameSuffixes: []string{".azureedge.net", ".azurefd.net"},
	},
	{
		name:          "Vercel",
		headerPresent: []string{"x-vercel-id", "x-vercel-cache"},
		cnameSuffixes: []string{".vercel-dns.com", ".vercel.app"},
	},
	{
		name:          "Netlify",
		headerPresent: []string{"x-nf-request-id"},
		serverSubstr:  []string{"netlify"},
		cnameSuffixes: []string{".netlify.app", ".netlifyglobalcdn.com"},
	},
	{
		name:          "Google Cloud CDN / Hosting",
		cnameSuffixes: []string{"ghs.google.com", ".googlehosted.com"},
		headerHints:   map[string][]string{"server": {"gws"}},
	},
	{
		name:          "Sucuri",
		headerPresent: []string{"x-sucuri-id", "x-sucuri-cache"},
	},
	{
		name:          "Imperva Incapsula",
		headerPresent: []string{"x-iinfo"},
		cnameSuffixes: []string{".incapdns.net"},
	},
	{
		name:          "StackPath",
		cnameSuffixes: []string{".stackpathdns.com", ".stackpathcdn.com"},
	},
	{
		name:          "KeyCDN",
		cnameSuffixes: []string{".kxcdn.com"},
	},
	{
		name:          "Bunny CDN",
		cnameSuffixes: []string{".b-cdn.net"},
	},
	{
		name:          "GitHub Pages",
		cnameSuffixes: []string{".github.io", "githubusercontent.com"},
	},
	{
		name:          "WP Engine",
		cnameSuffixes: []string{".wpengine.com"},
	},
	{
		name:          "Shopify",
		cnameSuffixes: []string{".myshopify.com", "shops.myshopify.com"},
	},
}

// detectCDN looks at the response headers first (cheapest, most specific) and falls back to
// the CNAME chain. Returns "" when nothing matches — most origin-hosted sites have no CDN.
func detectCDN(headers http.Header, cname string) string {
	server := strings.ToLower(headers.Get("Server"))
	lowerCNAME := strings.ToLower(cname)

	for _, sig := range cdnSignatures {
		for _, h := range sig.headerPresent {
			if headers.Get(h) != "" {
				return sig.name
			}
		}
		for headerName, substrs := range sig.headerHints {
			val := strings.ToLower(headers.Get(headerName))
			if val == "" {
				continue
			}
			for _, s := range substrs {
				if strings.Contains(val, s) {
					return sig.name
				}
			}
		}
		for _, s := range sig.serverSubstr {
			if strings.Contains(server, s) {
				return sig.name
			}
		}
		if lowerCNAME != "" {
			for _, suffix := range sig.cnameSuffixes {
				if strings.HasSuffix(lowerCNAME, suffix) {
					return sig.name
				}
			}
		}
	}
	return ""
}
