package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// maxBodyBytes caps how much of a response body we read per domain. Wappalyzer signatures live
// in the head and the first part of the body; capping keeps memory flat regardless of page size.
const maxBodyBytes = 2 << 20 // 2 MiB

type fetchResult struct {
	finalURL   string
	status     int
	headers    map[string][]string
	body       []byte
	usedScheme string
}

// fetchPage does one GET against the domain's homepage. If the caller didn't specify a scheme
// explicitly and HTTPS fails outright (dial/TLS error, not a non-2xx status), it retries once
// over plain HTTP — some marketing sites still don't terminate TLS.
func fetchPage(client *http.Client, rawURL string, hadExplicitScheme bool, timeout time.Duration) (*fetchResult, error) {
	res, err := doFetch(client, rawURL, timeout)
	if err == nil {
		return res, nil
	}
	if hadExplicitScheme || !strings.HasPrefix(rawURL, "https://") {
		return nil, err
	}
	httpURL := "http://" + strings.TrimPrefix(rawURL, "https://")
	res, err2 := doFetch(client, httpURL, timeout)
	if err2 != nil {
		return nil, fmt.Errorf("https failed (%v), http fallback failed (%v)", err, err2)
	}
	return res, nil
}

func doFetch(client *http.Client, rawURL string, timeout time.Duration) (*fetchResult, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; CatalystTechStackLookup/1.0; +https://apify.com)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	c := *client
	c.Timeout = timeout
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil && len(body) == 0 {
		return nil, err
	}

	finalURL := rawURL
	if resp.Request != nil && resp.Request.URL != nil {
		finalURL = resp.Request.URL.String()
	}

	return &fetchResult{
		finalURL: finalURL,
		status:   resp.StatusCode,
		headers:  map[string][]string(resp.Header),
		body:     body,
	}, nil
}
