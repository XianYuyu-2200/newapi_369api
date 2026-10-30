package service

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const canvasCoverTimeout = 20 * time.Second

// NormalizeCanvasCoverURL accepts a public https image URL.
// 198.18.0.0/15 is allowed because local proxy fake-ip uses that range;
// loopback, link-local, and private ranges stay blocked.
func NormalizeCanvasCoverURL(raw string) (string, error) {
	parsed, err := url.Parse(rewriteGitHubContentURL(strings.TrimSpace(raw)))
	if err != nil || parsed.Hostname() == "" || parsed.User != nil || !strings.EqualFold(parsed.Scheme, "https") {
		return "", errInvalidCanvasCover
	}
	parsed.Scheme = "https"
	parsed.Fragment = ""
	host := parsed.Hostname()
	port := parsed.Port()
	if port == "" {
		port = "443"
	}
	if port != "443" {
		return "", errInvalidCanvasCover
	}
	if ip := net.ParseIP(host); ip != nil {
		if !canvasCoverIPAllowed(ip) {
			return "", errInvalidCanvasCover
		}
		return parsed.String(), nil
	}
	ips, err := net.DefaultResolver.LookupIPAddr(context.Background(), host)
	if err != nil || len(ips) == 0 {
		return "", fmt.Errorf("cover host lookup failed")
	}
	for _, ip := range ips {
		if !canvasCoverIPAllowed(ip.IP) {
			return "", errInvalidCanvasCover
		}
	}
	return parsed.String(), nil
}

func rewriteGitHubContentURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil || !strings.EqualFold(parsed.Hostname(), "github.com") {
		return raw
	}
	parts := strings.Split(strings.Trim(parsed.EscapedPath(), "/"), "/")
	if len(parts) < 5 || (parts[2] != "blob" && parts[2] != "raw") {
		return raw
	}
	return "https://raw.githubusercontent.com/" + parts[0] + "/" + parts[1] + "/" + parts[3] + "/" + strings.Join(parts[4:], "/")
}

var errInvalidCanvasCover = invalidCanvasCoverError{}

type invalidCanvasCoverError struct{}

func (invalidCanvasCoverError) Error() string { return "invalid cover url" }

func canvasCoverIPAllowed(ip net.IP) bool {
	if ip == nil {
		return false
	}
	if ip.IsLoopback() || ip.IsUnspecified() || ip.IsMulticast() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsPrivate() {
		return false
	}
	v4 := ip.To4()
	if v4 == nil {
		return true
	}
	if v4[0] == 0 || v4[0] >= 224 {
		return false
	}
	if v4[0] == 100 && v4[1] >= 64 && v4[1] <= 127 {
		return false
	}
	// 198.18.0.0/15 is IANA benchmarking space and also the fake-ip range used
	// by the local proxy. Dialing it reaches the proxy, not a private network.
	if v4[0] == 198 && (v4[1] == 18 || v4[1] == 19) {
		return true
	}
	if v4[0] == 192 && v4[1] == 0 && v4[2] == 0 {
		return false
	}
	if v4[0] == 192 && v4[1] == 0 && v4[2] == 2 {
		return false
	}
	if v4[0] == 198 && v4[1] == 51 && v4[2] == 100 {
		return false
	}
	if v4[0] == 203 && v4[1] == 0 && v4[2] == 113 {
		return false
	}
	return true
}

func canvasCoverDialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil || port != "443" {
		return nil, errInvalidCanvasCover
	}
	dial := (&net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}).DialContext
	if ip := net.ParseIP(host); ip != nil {
		if !canvasCoverIPAllowed(ip) {
			return nil, errInvalidCanvasCover
		}
		return dial(ctx, network, net.JoinHostPort(ip.String(), port))
	}
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	var last error
	for _, ip := range ips {
		if !canvasCoverIPAllowed(ip.IP) {
			return nil, errInvalidCanvasCover
		}
		conn, dialErr := dial(ctx, network, net.JoinHostPort(ip.IP.String(), port))
		if dialErr == nil {
			return conn, nil
		}
		last = dialErr
	}
	if last != nil {
		return nil, last
	}
	return nil, errInvalidCanvasCover
}

func NewCanvasCoverHTTPClient() *http.Client {
	transport := &http.Transport{
		Proxy:               nil,
		DialContext:         canvasCoverDialContext,
		ForceAttemptHTTP2:   true,
		TLSHandshakeTimeout: 10 * time.Second,
		IdleConnTimeout:     30 * time.Second,
		MaxIdleConns:        32,
	}
	return &http.Client{
		Transport: transport,
		Timeout:   canvasCoverTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 || req == nil || req.URL == nil {
				return errInvalidCanvasCover
			}
			if _, err := NormalizeCanvasCoverURL(req.URL.String()); err != nil {
				return err
			}
			return nil
		},
	}
}

var (
	canvasCoverClientOnce sync.Once
	canvasCoverClient     *http.Client
)

func CanvasCoverHTTPClient() *http.Client {
	canvasCoverClientOnce.Do(func() {
		canvasCoverClient = NewCanvasCoverHTTPClient()
	})
	return canvasCoverClient
}
