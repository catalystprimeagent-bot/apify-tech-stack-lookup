package main

import (
	"context"
	"net"
	"sort"
	"strings"
	"time"
)

// lookupDNS resolves A, AAAA, CNAME and NS records for host. Each sub-lookup fails
// independently — a domain with no AAAA record, for instance, still reports its A/CNAME/NS.
func lookupDNS(ctx context.Context, host string, timeout time.Duration) DNSInfo {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var info DNSInfo
	resolver := net.DefaultResolver

	ips, err := resolver.LookupIPAddr(ctx, host)
	if err != nil {
		info.Error = err.Error()
	} else {
		for _, ip := range ips {
			if ip.IP.To4() != nil {
				info.A = append(info.A, ip.IP.String())
			} else {
				info.AAAA = append(info.AAAA, ip.IP.String())
			}
		}
		sort.Strings(info.A)
		sort.Strings(info.AAAA)
	}

	if cname, err := resolver.LookupCNAME(ctx, host); err == nil {
		cname = strings.TrimSuffix(cname, ".")
		if !strings.EqualFold(cname, host) {
			info.CNAME = cname
		}
	}

	if ns, err := resolver.LookupNS(ctx, host); err == nil {
		for _, n := range ns {
			info.NS = append(info.NS, strings.TrimSuffix(n.Host, "."))
		}
		sort.Strings(info.NS)
	}

	return info
}

// lookupMail resolves MX records and maps the lowest-preference target to a known mail
// provider. Unrecognized targets are still reported under "records" — only "provider" is
// left empty when we don't recognize the pattern.
func lookupMail(ctx context.Context, host string, timeout time.Duration) MailInfo {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var info MailInfo
	mxs, err := net.DefaultResolver.LookupMX(ctx, host)
	if err != nil {
		info.Error = err.Error()
		return info
	}
	sort.Slice(mxs, func(i, j int) bool { return mxs[i].Pref < mxs[j].Pref })
	for _, mx := range mxs {
		info.Records = append(info.Records, strings.TrimSuffix(mx.Host, "."))
	}
	if len(info.Records) > 0 {
		info.Provider = guessMailProvider(info.Records[0])
	}
	return info
}

// mailProviderSignatures maps a substring found in the primary MX hostname to a human-readable
// provider name. Ordered by specificity where it matters; the first match wins.
var mailProviderSignatures = []struct {
	substr   string
	provider string
}{
	{"google.com", "Google Workspace"},
	{"googlemail.com", "Google Workspace"},
	{"outlook.com", "Microsoft 365"},
	{"protection.outlook.com", "Microsoft 365"},
	{"pphosted.com", "Proofpoint"},
	{"mimecast.com", "Mimecast"},
	{"zoho.com", "Zoho Mail"},
	{"zohomail.com", "Zoho Mail"},
	{"yandex.net", "Yandex Mail"},
	{"qq.com", "Tencent QQ Mail"},
	{"mail.ru", "Mail.ru"},
	{"secureserver.net", "GoDaddy Email"},
	{"mailgun.org", "Mailgun"},
	{"sendgrid.net", "SendGrid"},
	{"amazonses.com", "Amazon SES"},
	{"icloud.com", "iCloud Mail"},
	{"fastmail.com", "Fastmail"},
	{"fastmail.fm", "Fastmail"},
	{"titan.email", "Titan Email"},
	{"improvmx.com", "ImprovMX"},
	{"messagingengine.com", "Fastmail"},
	{"emailsrvr.com", "Rackspace Email"},
	{"registrar-servers.com", "Namecheap Private Email"},
	{"privateemail.com", "Namecheap Private Email"},
	{"ovh.net", "OVH Mail"},
	{"hostgator.com", "HostGator Email"},
	{"bluehost.com", "Bluehost Email"},
	{"1and1.com", "IONOS Mail"},
	{"ionos.com", "IONOS Mail"},
}

func guessMailProvider(mxHost string) string {
	lower := strings.ToLower(mxHost)
	for _, sig := range mailProviderSignatures {
		if strings.Contains(lower, sig.substr) {
			return sig.provider
		}
	}
	return ""
}
