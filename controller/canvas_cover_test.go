package controller

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestNormalizeCanvasCoverURL(t *testing.T) {
	t.Parallel()
	got, err := normalizeCanvasCoverURL("https://cdn.jsdelivr.net/gh/example/repo/a.png")
	require.NoError(t, err)
	require.Equal(t, "https://cdn.jsdelivr.net/gh/example/repo/a.png", got)

	got, err = normalizeCanvasCoverURL("https://github.com/foo/bar/blob/main/images/a.png")
	require.NoError(t, err)
	require.Equal(t, "https://raw.githubusercontent.com/foo/bar/main/images/a.png", got)

	for _, raw := range []string{
		"",
		"http://cdn.jsdelivr.net/a.png",
		"https://127.0.0.1/a.png",
		"https://169.254.169.254/latest/meta-data",
		"https://user:pass@cdn.jsdelivr.net/a.png",
		"file:///etc/passwd",
		"https://cdn.jsdelivr.net:8443/a.png",
	} {
		_, err := normalizeCanvasCoverURL(raw)
		require.Error(t, err, raw)
	}
}

func TestCanvasImageContentType(t *testing.T) {
	t.Parallel()
	require.Equal(t, "image/jpeg", canvasImageContentType([]byte{0xFF, 0xD8, 0xFF, 0x00}))
	require.Equal(t, "image/png", canvasImageContentType([]byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0x00}))
	require.Equal(t, "image/gif", canvasImageContentType([]byte("GIF89aXXXX")))
	webp := append([]byte("RIFF"), []byte{0, 0, 0, 0}...)
	webp = append(webp, []byte("WEBP")...)
	require.Equal(t, "image/webp", canvasImageContentType(webp))
	require.Equal(t, "", canvasImageContentType([]byte("<html></html>")))
	require.Equal(t, "", canvasImageContentType([]byte("<svg xmlns='http://www.w3.org/2000/svg'></svg>")))
}

func TestCanvasCoverProxyRejectsPrivateURLWithoutFetch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	called := false
	previous := canvasCoverDo
	canvasCoverDo = func(req *http.Request) (*http.Response, error) {
		called = true
		return nil, nil
	}
	t.Cleanup(func() { canvasCoverDo = previous })

	router := gin.New()
	router.GET("/api/canvas/cover", CanvasCoverProxy)
	req := httptest.NewRequest(http.MethodGet, "/api/canvas/cover?url=https://127.0.0.1/secret.png", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.False(t, called)
}

func TestCanvasCoverProxyReturnsImage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previous := canvasCoverDo
	canvasCoverDo = func(req *http.Request) (*http.Response, error) {
		require.Equal(t, "https://raw.githubusercontent.com/foo/bar/main/a.png", req.URL.String())
		body := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0x00}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(string(body))),
			Header:     make(http.Header),
		}, nil
	}
	t.Cleanup(func() { canvasCoverDo = previous })

	router := gin.New()
	router.GET("/api/canvas/cover", CanvasCoverProxy)
	req := httptest.NewRequest(http.MethodGet, "/api/canvas/cover?url="+urlQuery("https://github.com/foo/bar/blob/main/a.png"), nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "image/png", rec.Header().Get("Content-Type"))
	require.Contains(t, rec.Header().Get("Cache-Control"), "max-age=86400")
}

func urlQuery(raw string) string {
	return strings.ReplaceAll(strings.ReplaceAll(raw, ":", "%3A"), "/", "%2F")
}
