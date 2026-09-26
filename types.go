package main

// Input mirrors the incumbent's (nexgendata/wappalyzer-replacement) input schema so a buyer
// can swap tools without editing their saved input. See README.md "Compatibility" section.
type Input struct {
	Urls              interface{} `json:"urls"`
	StartUrls         interface{} `json:"startUrls"`
	Domains           interface{} `json:"domains"`
	Websites          interface{} `json:"websites"`
	TargetUrls        interface{} `json:"targetUrls"`
	URL               interface{} `json:"url"`
	CategoriesFilter  []string    `json:"categories_filter"`
	IncludeConfidence *bool       `json:"include_confidence"`
	IncludeVersions   *bool       `json:"include_versions"`
	TimeoutSeconds    *int        `json:"timeout_seconds"`
}

// Technology is one detected tech entry in the output.
type Technology struct {
	Name       string   `json:"name"`
	Version    string   `json:"version,omitempty"`
	Categories []string `json:"categories,omitempty"`
	Confidence int      `json:"confidence,omitempty"`
}

// DNSInfo holds the resolved DNS records for a domain.
type DNSInfo struct {
	A     []string `json:"a,omitempty"`
	AAAA  []string `json:"aaaa,omitempty"`
	CNAME string   `json:"cname,omitempty"`
	NS    []string `json:"ns,omitempty"`
	Error string   `json:"error,omitempty"`
}

// TLSInfo holds what the TLS handshake on port 443 revealed.
type TLSInfo struct {
	Issuer          string `json:"issuer,omitempty"`
	Subject         string `json:"subject,omitempty"`
	NotAfter        string `json:"notAfter,omitempty"`
	DaysUntilExpiry int    `json:"daysUntilExpiry,omitempty"`
	Error           string `json:"error,omitempty"`
}

// MailInfo holds the MX records and the guessed mail provider.
type MailInfo struct {
	Records  []string `json:"records,omitempty"`
	Provider string   `json:"provider,omitempty"`
	Error    string   `json:"error,omitempty"`
}

// Result is one dataset item: one row per input domain.
type Result struct {
	URL          string       `json:"url"`
	Domain       string       `json:"domain"`
	Technologies []Technology `json:"technologies"`
	TechNames    []string     `json:"techNames"`
	TechCount    int          `json:"techCount"`
	DNS          DNSInfo      `json:"dns"`
	CDN          string       `json:"cdn,omitempty"`
	TLS          *TLSInfo     `json:"tls,omitempty"`
	Mail         MailInfo     `json:"mail"`
	HTTPStatus   int          `json:"httpStatus,omitempty"`
	Charged      bool         `json:"charged"`
	Error        string       `json:"error,omitempty"`
}
