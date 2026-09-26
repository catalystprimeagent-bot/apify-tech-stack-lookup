package main

import (
	"net/http"
	"sort"
	"strings"

	wappalyzer "github.com/projectdiscovery/wappalyzergo"
)

// detectTechnologies runs wappalyzergo against the fetched page and shapes the result into our
// output format. wappalyzergo does not expose a numeric confidence score through its public API
// (it uses one internally to decide whether a match counts at all, then discards it) — when the
// buyer asks for include_confidence we report a coarse two-tier estimate instead of pretending to
// replicate the incumbent's internal scoring: 100 when a version number was extracted (implies a
// specific, high-precision pattern match), 85 otherwise. This is documented in the README.
func detectTechnologies(client *wappalyzer.Wappalyze, headers http.Header, body []byte, includeVersions, includeConfidence bool, categoryFilter map[string]bool) ([]Technology, []string) {
	info := client.FingerprintWithInfo(headers, body)

	techs := make([]Technology, 0, len(info))
	for key, appInfo := range info {
		name := key
		version := ""
		if idx := strings.LastIndex(key, ":"); idx > 0 {
			name = key[:idx]
			version = key[idx+1:]
		}

		if len(categoryFilter) > 0 {
			matched := false
			for _, cat := range appInfo.Categories {
				if categoryFilter[strings.ToLower(cat)] {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		t := Technology{
			Name:       name,
			Categories: appInfo.Categories,
		}
		if includeVersions {
			t.Version = version
		}
		if includeConfidence {
			if version != "" {
				t.Confidence = 100
			} else {
				t.Confidence = 85
			}
		}
		techs = append(techs, t)
	}

	sort.Slice(techs, func(i, j int) bool { return techs[i].Name < techs[j].Name })

	names := make([]string, len(techs))
	for i, t := range techs {
		names[i] = t.Name
	}
	return techs, names
}
