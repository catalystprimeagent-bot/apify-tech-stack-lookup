package main

import (
	"context"
	"crypto/tls"
	"net"
	"time"
)

// inspectTLS opens a TLS connection to host:443 and reports the leaf certificate's issuer and
// expiry. InsecureSkipVerify is intentional: we're reporting what the server presents, not
// making a trust decision — a self-signed or mismatched cert is data, not a reason to fail.
func inspectTLS(ctx context.Context, host string, timeout time.Duration) *TLSInfo {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	dialer := &net.Dialer{}
	rawConn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(host, "443"))
	if err != nil {
		return &TLSInfo{Error: err.Error()}
	}
	defer rawConn.Close()

	tlsConn := tls.Client(rawConn, &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: true, // reporting only; see doc comment above
		MinVersion:         tls.VersionTLS10,
	})
	if deadline, ok := ctx.Deadline(); ok {
		_ = tlsConn.SetDeadline(deadline)
	}
	if err := tlsConn.Handshake(); err != nil {
		return &TLSInfo{Error: err.Error()}
	}
	defer tlsConn.Close()

	state := tlsConn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return &TLSInfo{Error: "no certificate presented"}
	}
	cert := state.PeerCertificates[0]

	issuer := cert.Issuer.CommonName
	if issuer == "" && len(cert.Issuer.Organization) > 0 {
		issuer = cert.Issuer.Organization[0]
	}
	subject := cert.Subject.CommonName

	daysLeft := int(time.Until(cert.NotAfter).Hours() / 24)

	return &TLSInfo{
		Issuer:          issuer,
		Subject:         subject,
		NotAfter:        cert.NotAfter.UTC().Format(time.RFC3339),
		DaysUntilExpiry: daysLeft,
	}
}
