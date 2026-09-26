package main

import (
	"encoding/json"
	"net/url"
	"strings"
)

// domainsFromInput gathers every domain/URL the buyer supplied across the incumbent's five
// aliases (urls, startUrls, domains, websites, targetUrls) plus the single-value "url" field,
// normalizes each into a fetchable URL + bare host, and dedupes while preserving order.
func domainsFromInput(in Input) (entries []normalizedEntry, rejected []string) {
	seen := make(map[string]bool)

	add := func(raw string) {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return
		}
		norm, host, err := normalizeURL(raw)
		if err != nil {
			rejected = append(rejected, raw)
			return
		}
		key := strings.ToLower(host)
		if seen[key] {
			return
		}
		seen[key] = true
		entries = append(entries, normalizedEntry{url: norm, host: host, original: raw})
	}

	for _, raw := range coerceToList(in.Urls) {
		add(raw)
	}
	for _, raw := range coerceToList(in.StartUrls) {
		add(raw)
	}
	for _, raw := range coerceToList(in.Domains) {
		add(raw)
	}
	for _, raw := range coerceToList(in.Websites) {
		add(raw)
	}
	for _, raw := range coerceToList(in.TargetUrls) {
		add(raw)
	}
	for _, raw := range coerceToList(in.URL) {
		add(raw)
	}

	return entries, rejected
}

type normalizedEntry struct {
	url      string
	host     string
	original string
}

// coerceToList mirrors the incumbent's documented input flexibility: a JSON array of strings,
// a JSON array of {url: "..."} objects, a single string, or a comma/newline-separated string
// are all accepted for every URL-bearing field.
func coerceToList(v interface{}) []string {
	switch val := v.(type) {
	case nil:
		return nil
	case string:
		return splitStringList(val)
	case []interface{}:
		var out []string
		for _, item := range val {
			switch it := item.(type) {
			case string:
				out = append(out, it)
			case map[string]interface{}:
				if u, ok := it["url"].(string); ok {
					out = append(out, u)
				}
			}
		}
		return out
	case []string:
		return val
	default:
		// Might be a raw json.RawMessage-shaped value coming through re-marshal; try once more.
		b, err := json.Marshal(val)
		if err != nil {
			return nil
		}
		var arr []interface{}
		if err := json.Unmarshal(b, &arr); err == nil {
			return coerceToList(arr)
		}
		var s string
		if err := json.Unmarshal(b, &s); err == nil {
			return splitStringList(s)
		}
		return nil
	}
}

func splitStringList(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	fields := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r'
	})
	var out []string
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f != "" {
			out = append(out, f)
		}
	}
	return out
}

// normalizeURL turns a bare host or full URL into a fetchable absolute URL plus its hostname.
// A scheme is added (https) when missing, matching the incumbent's documented behavior.
func normalizeURL(raw string) (absoluteURL string, host string, err error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", errEmptyInput
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, parseErr := url.Parse(raw)
	if parseErr != nil {
		return "", "", parseErr
	}
	h := u.Hostname()
	if h == "" {
		return "", "", errNoHost
	}
	return u.String(), h, nil
}
