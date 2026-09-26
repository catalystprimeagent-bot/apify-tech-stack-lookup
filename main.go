package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	wappalyzer "github.com/projectdiscovery/wappalyzergo"
)

const (
	defaultTimeoutSeconds = 15
	minTimeoutSeconds     = 3
	maxTimeoutSeconds     = 60
	dnsLookupTimeout      = 8 * time.Second
	tlsLookupTimeout      = 8 * time.Second
	concurrency           = 10
	pushBatchSize         = 10
)

func main() {
	if err := run(); err != nil {
		log.Printf("FATAL: %v", err)
		os.Exit(1)
	}
}

func run() error {
	env := loadEnv()
	httpClient := &http.Client{}

	input, err := getInput(env, httpClient)
	if err != nil {
		return fmt.Errorf("reading input: %w", err)
	}

	entries, rejected := domainsFromInput(input)
	for _, r := range rejected {
		log.Printf("skipping unparseable entry: %q", r)
	}
	if len(entries) == 0 {
		return fmt.Errorf("no valid domains in input — provide at least one URL/domain via " +
			"\"urls\" (or its aliases startUrls, domains, websites, targetUrls, url)")
	}
	log.Printf("processing %d domain(s)", len(entries))

	timeout := time.Duration(clampInt(intOrDefault(input.TimeoutSeconds, defaultTimeoutSeconds), minTimeoutSeconds, maxTimeoutSeconds)) * time.Second
	includeVersions := boolOrDefault(input.IncludeVersions, true)
	includeConfidence := boolOrDefault(input.IncludeConfidence, true)
	categoryFilter := toLowerSet(input.CategoriesFilter)

	wClient, err := wappalyzer.New()
	if err != nil {
		return fmt.Errorf("loading tech-detection fingerprints: %w", err)
	}

	pricing := newPricingManager(env, httpClient, primaryChargeEvent)
	writer := newResultWriter(env, httpClient)

	results := processAll(entries, concurrency, func(e normalizedEntry) Result {
		return processDomain(httpClient, wClient, e, timeout, includeVersions, includeConfidence, categoryFilter, pricing)
	})

	charged := 0
	for i := 0; i < len(results); i += pushBatchSize {
		end := i + pushBatchSize
		if end > len(results) {
			end = len(results)
		}
		batch := results[i:end]
		if err := writer.Push(batch); err != nil {
			return fmt.Errorf("pushing results to dataset: %w", err)
		}
		for _, r := range batch {
			if r.Charged {
				charged++
			}
		}
	}

	log.Printf("done: %d domain(s) processed, %d dataset item(s) written, %d charged", len(results), len(results), charged)
	return nil
}

// processAll runs fn over items with bounded concurrency and returns results in input order.
func processAll(entries []normalizedEntry, workers int, fn func(normalizedEntry) Result) []Result {
	results := make([]Result, len(entries))
	if workers > len(entries) {
		workers = len(entries)
	}
	if workers < 1 {
		workers = 1
	}

	jobs := make(chan int)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				results[i] = fn(entries[i])
			}
		}()
	}
	for i := range entries {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	return results
}

func processDomain(httpClient *http.Client, wClient *wappalyzer.Wappalyze, e normalizedEntry, timeout time.Duration, includeVersions, includeConfidence bool, categoryFilter map[string]bool, pricing *pricingManager) Result {
	result := Result{URL: e.url, Domain: e.host}
	ctx := context.Background()

	hadExplicitScheme := strings.Contains(e.original, "://")

	var wg sync.WaitGroup
	wg.Add(3)
	go func() { defer wg.Done(); result.DNS = lookupDNS(ctx, e.host, dnsLookupTimeout) }()
	go func() { defer wg.Done(); result.Mail = lookupMail(ctx, e.host, dnsLookupTimeout) }()
	go func() { defer wg.Done(); result.TLS = inspectTLS(ctx, e.host, tlsLookupTimeout) }()

	fetch, fetchErr := fetchPage(httpClient, e.url, hadExplicitScheme, timeout)

	wg.Wait()

	result.CDN = detectCDN(headerMap(fetchHeaders(fetch)), result.DNS.CNAME)

	hasAnySignal := len(result.DNS.A) > 0 || len(result.DNS.AAAA) > 0 || len(result.Mail.Records) > 0 || (result.TLS != nil && result.TLS.Error == "")

	if fetchErr != nil {
		result.Error = fetchErr.Error()
	} else {
		result.HTTPStatus = fetch.status
		techs, names := detectTechnologies(wClient, headerMap(fetch.headers), fetch.body, includeVersions, includeConfidence, categoryFilter)
		result.Technologies = techs
		result.TechNames = names
		result.TechCount = len(techs)
		hasAnySignal = true
	}

	if hasAnySignal {
		result.Charged = pricing.ChargeDomain(e.host)
	}

	return result
}

func fetchHeaders(f *fetchResult) map[string][]string {
	if f == nil {
		return nil
	}
	return f.headers
}

func headerMap(h map[string][]string) http.Header {
	return http.Header(h)
}

func intOrDefault(v *int, def int) int {
	if v == nil {
		return def
	}
	return *v
}

func boolOrDefault(v *bool, def bool) bool {
	if v == nil {
		return def
	}
	return *v
}

func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func toLowerSet(items []string) map[string]bool {
	if len(items) == 0 {
		return nil
	}
	set := make(map[string]bool, len(items))
	for _, i := range items {
		set[strings.ToLower(i)] = true
	}
	return set
}
