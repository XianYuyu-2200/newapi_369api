package controller

import (
	"bytes"
	"io"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

const (
	canvasCoverMaxBytes = 8 << 20
	canvasCoverRate     = 12.0
	canvasCoverBurst    = 48.0
)

var canvasCoverDo = func(req *http.Request) (*http.Response, error) {
	return service.CanvasCoverHTTPClient().Do(req)
}

// CanvasCoverProxy serves prompt-library thumbnails from this origin.
// Browsers request GET /api/canvas/cover?url=https://... and this handler
// fetches a public https image. It is intentionally outside the global API
// rate limit and gzip middleware.
func CanvasCoverProxy(c *gin.Context) {
	target, err := normalizeCanvasCoverURL(c.Query("url"))
	if err != nil {
		c.String(http.StatusBadRequest, "invalid cover url")
		return
	}
	if !allowCanvasCover(c.ClientIP()) {
		c.Header("Retry-After", "1")
		c.String(http.StatusTooManyRequests, "too many cover requests")
		return
	}

	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, target, nil)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid cover url")
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; 369api-cover/1.0)")
	req.Header.Set("Accept", "image/avif,image/webp,image/png,image/jpeg,image/gif;q=0.9,*/*;q=0.1")

	resp, err := canvasCoverDo(req)
	if err != nil {
		c.String(http.StatusBadGateway, "cover fetch failed")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.String(http.StatusBadGateway, "cover fetch failed")
		return
	}
	if resp.ContentLength > canvasCoverMaxBytes {
		c.String(http.StatusRequestEntityTooLarge, "cover too large")
		return
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, canvasCoverMaxBytes+1))
	if err != nil {
		c.String(http.StatusBadGateway, "cover fetch failed")
		return
	}
	if len(body) == 0 || len(body) > canvasCoverMaxBytes {
		c.String(http.StatusRequestEntityTooLarge, "cover too large")
		return
	}
	contentType := canvasImageContentType(body)
	if contentType == "" {
		c.String(http.StatusUnsupportedMediaType, "cover is not an image")
		return
	}

	c.Header("Cache-Control", "public, max-age=86400")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Data(http.StatusOK, contentType, body)
}

func normalizeCanvasCoverURL(raw string) (string, error) {
	return service.NormalizeCanvasCoverURL(raw)
}

func canvasImageContentType(data []byte) string {
	switch {
	case len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF:
		return "image/jpeg"
	case len(data) >= 8 && bytes.Equal(data[:8], []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}):
		return "image/png"
	case len(data) >= 6 && (bytes.HasPrefix(data, []byte("GIF87a")) || bytes.HasPrefix(data, []byte("GIF89a"))):
		return "image/gif"
	case len(data) >= 12 && bytes.Equal(data[:4], []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP")):
		return "image/webp"
	case len(data) >= 12 && bytes.Equal(data[4:8], []byte("ftyp")) && isAVIFBrand(string(data[8:12])):
		return "image/avif"
	default:
		return ""
	}
}

func isAVIFBrand(brand string) bool {
	switch brand {
	case "avif", "avis", "mif1":
		return true
	default:
		return false
	}
}

type canvasCoverBucket struct {
	tokens  float64
	updated time.Time
}

var canvasCoverLimiter = struct {
	mu      sync.Mutex
	buckets map[string]*canvasCoverBucket
}{buckets: map[string]*canvasCoverBucket{}}

func allowCanvasCover(ip string) bool {
	if ip == "" {
		ip = "unknown"
	}
	now := time.Now()
	canvasCoverLimiter.mu.Lock()
	defer canvasCoverLimiter.mu.Unlock()
	if len(canvasCoverLimiter.buckets) > 1024 {
		canvasCoverLimiter.buckets = map[string]*canvasCoverBucket{}
	}
	bucket := canvasCoverLimiter.buckets[ip]
	if bucket == nil {
		bucket = &canvasCoverBucket{tokens: canvasCoverBurst, updated: now}
		canvasCoverLimiter.buckets[ip] = bucket
	}
	elapsed := now.Sub(bucket.updated).Seconds()
	if elapsed > 0 {
		bucket.tokens = math.Min(canvasCoverBurst, bucket.tokens+elapsed*canvasCoverRate)
		bucket.updated = now
	}
	if bucket.tokens < 1 {
		return false
	}
	bucket.tokens--
	return true
}
