# Tech Stack & Infrastructure Lookup

Bulk domain lookup. For each domain you get the detected technology stack (CMS, analytics,
frameworks, etc.) plus four things most tech-stack tools don't report: **DNS records, CDN,
TLS certificate issuer, and mail provider.**

No headless browser, no residential proxy, one HTTP GET plus a handful of DNS/TLS lookups per
domain. Built to run in 256 MB and finish in seconds.

**Run it on Apify:** [https://apify.com/catalyst_prime/tech-stack-lookup](https://apify.com/catalyst_prime/tech-stack-lookup)

Free to run, no setup. You only use your own Apify account's included usage.

## Input

Give it one or more domains via `urls` (a JSON array, a single string, or a comma/newline
separated list of domains). If you're switching from another tech-detection tool, you can also
use whichever of these field names your existing input already uses. This Actor accepts all of
them as aliases for the same thing: `startUrls`, `domains`, `websites`, `targetUrls`, `url`.
Supply domains in any one of them, or mix several; they are combined and deduplicated.

| Field | Type | Default | Description |
|---|---|---|---|
| `urls` | array of strings | none | Domains or URLs to look up. Supply this or any one alias below. |
| `startUrls` | array | none | Alias of `urls`. Accepts plain strings or `{"url": "..."}` objects. |
| `domains` | array | none | Alias of `urls`. |
| `websites` | array | none | Alias of `urls`. |
| `targetUrls` | array | none | Alias of `urls`. |
| `url` | string | none | Alias of `urls` for a single domain. |
| `categories_filter` | array of strings | none (all categories) | Only report technologies in these categories, e.g. `["CMS", "Analytics"]`. |
| `include_versions` | boolean | `true` | Include the detected version number, when found. |
| `include_confidence` | boolean | `true` | Include a 0-100 confidence score per technology. |
| `timeout_seconds` | integer | `15` | Per-domain page-load timeout, clamped to 3-60. |

## Output

One dataset item per input domain:

```json
{
  "url": "https://stripe.com",
  "domain": "stripe.com",
  "technologies": [
    { "name": "Nginx", "categories": ["Web servers", "Reverse proxies"], "confidence": 85 }
  ],
  "techNames": ["Amazon S3", "Amazon Web Services", "HSTS", "Nginx"],
  "techCount": 4,
  "dns": {
    "a": ["198.137.150.231", "198.202.176.231"],
    "ns": ["ns-1087.awsdns-07.org", "ns-423.awsdns-52.com"]
  },
  "cdn": "Cloudflare",
  "tls": {
    "issuer": "DigiCert Global G3 TLS ECC SHA384 2020 CA1",
    "subject": "stripe.com",
    "notAfter": "2026-11-12T23:59:59Z",
    "daysUntilExpiry": 49
  },
  "mail": {
    "records": ["aspmx.l.google.com", "alt1.aspmx.l.google.com"],
    "provider": "Google Workspace"
  },
  "httpStatus": 200,
  "charged": false
}
```

`cdn` is omitted when no CDN was detected. `tls`/`dns`/`mail` always appear, each with its own
`error` field when that particular lookup failed. A domain that can't be reached still gets a
full row showing what *could* be resolved, so nothing is silently dropped. `charged` reports
whether this row's lookup was billed (see Pricing).

## Compatibility

This Actor's input schema mirrors the market's existing tech-detection tool so you can swap
tools without editing a saved input. What's new here: DNS records, CDN, TLS certificate issuer,
and mail provider, none of which the incumbent reports.

## Pricing

**Free.** This Actor has no price set, so every run costs you only your own Apify platform
usage (about $0.00002 of compute per domain, measured over a 10-domain run). The `charged`
field in each output row stays `false` while the Actor is free.

The code does support pay-per-event billing on a single `domain-lookup` event, priced per
successfully processed domain, if a price is ever set on the Apify Console. A domain that fails
entirely (unreachable, no DNS, no TLS, no mail records) would not be charged. Nothing is
hardcoded in the Actor: it reads its own current price from the platform at startup and runs
unmetered when there isn't one.

## Notes on the technology confidence score

The detection engine ([`wappalyzergo`](https://github.com/projectdiscovery/wappalyzergo)) does
not expose a numeric confidence score through its public API: it uses one internally to decide
whether a match counts at all, then discards it. When `include_confidence` is on, this Actor
reports a coarse two-tier estimate instead of pretending to replicate a score it doesn't have:
`100` when a version number was extracted (implies a specific, high-precision pattern match),
`85` otherwise.

## Example tasks

Saved, pre-configured runs you can try without writing any input:

- [Which sites use Cloudflare](https://apify.com/catalyst_prime/tech-stack-lookup/examples/which-sites-use-cloudflare)
- [Find expiring SSL certificates](https://apify.com/catalyst_prime/tech-stack-lookup/examples/find-expiring-ssl-certificates)
- [Compare competitor tech stacks](https://apify.com/catalyst_prime/tech-stack-lookup/examples/compare-competitor-tech-stacks)
- [Find a company's email provider](https://apify.com/catalyst_prime/tech-stack-lookup/examples/find-company-email-provider)
