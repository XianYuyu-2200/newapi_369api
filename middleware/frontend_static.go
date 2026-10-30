package middleware

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"mime"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"

	ginGzip "github.com/gin-contrib/gzip"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

// precompressedFile is the gzip encoding of one frontend build file.
type precompressedFile struct {
	once        sync.Once
	body        []byte
	contentType string
	err         error
}

// ServeFrontendFiles serves files from the embedded frontend build. The build
// never changes while the process runs, so each file is gzip-compressed once
// and the bytes are reused, instead of recompressing multi-megabyte chunks on
// every request. The cache holds at most one entry per build file. Clients
// without gzip, the image types the gzip middleware skips, and the
// index.html redirect keep using http.FileServer. Directories are not.
func ServeFrontendFiles(frontendFS static.ServeFileSystem) gin.HandlerFunc {
	fileServer := http.FileServer(frontendFS)
	var cache sync.Map // URL path -> *precompressedFile
	return func(c *gin.Context) {
		urlPath := c.Request.URL.Path
		if !frontendFS.Exists("/", urlPath) {
			return
		}
		// A directory must not be passed to FileServer. FileServer adds a
		// trailing slash, and the /:mode/mj route makes Gin remove that slash
		// again, so /downloads, /support, /agent and /canvas-app loop.
		// Nested apps keep their own index.html. Asset directories without one
		// are SPA routes and fall through to the dashboard index.
		if path.Ext(urlPath) == "" {
			if indexPath, isDir := frontendDirectoryIndex(frontendFS, urlPath); isDir {
				if indexPath == "" {
					return
				}
				c.Abort()
				serveEmbeddedFile(c, frontendFS, indexPath)
				return
			}
		}
		c.Abort()
		// FileServer redirects every index.html to its directory. Keep that
		// only for the dashboard entry, which the SPA fallback replaces.
		// Nested apps such as /canvas-app/index.html must be returned directly.
		if strings.HasSuffix(urlPath, "/index.html") && urlPath != "/index.html" {
			serveEmbeddedFile(c, frontendFS, urlPath)
			return
		}
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") ||
			ginGzip.DefaultExcludedExtentions.Contains(path.Ext(urlPath)) ||
			strings.HasSuffix(urlPath, "/index.html") {
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		value, ok := cache.Load(urlPath)
		if !ok {
			value, _ = cache.LoadOrStore(urlPath, &precompressedFile{})
		}
		file := value.(*precompressedFile)
		file.once.Do(func() {
			f, err := frontendFS.Open(urlPath)
			if err != nil {
				file.err = err
				return
			}
			defer f.Close()
			info, err := f.Stat()
			if err != nil {
				file.err = err
				return
			}
			if info.IsDir() {
				file.err = errors.New("directory")
				return
			}
			raw, err := io.ReadAll(f)
			if err != nil {
				file.err = err
				return
			}
			var compressed bytes.Buffer
			writer := gzip.NewWriter(&compressed)
			if _, err := writer.Write(raw); err != nil {
				file.err = err
				return
			}
			if err := writer.Close(); err != nil {
				file.err = err
				return
			}
			// Match http.FileServer: extension first, then content sniffing.
			file.contentType = mime.TypeByExtension(path.Ext(urlPath))
			if file.contentType == "" {
				file.contentType = http.DetectContentType(raw)
			}
			file.body = bytes.Clone(compressed.Bytes())
		})
		if file.err != nil {
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		// Range requests get the whole body; a server may ignore Range.
		header := c.Writer.Header()
		header.Set("Content-Encoding", "gzip")
		header.Set("Vary", "Accept-Encoding")
		header.Set("Content-Type", file.contentType)
		header.Set("Content-Length", strconv.Itoa(len(file.body)))
		c.Writer.WriteHeader(http.StatusOK)
		if c.Request.Method != http.MethodHead {
			_, _ = c.Writer.Write(file.body)
		}
	}
}

func frontendDirectoryIndex(frontendFS static.ServeFileSystem, urlPath string) (string, bool) {
	file, err := frontendFS.Open(urlPath)
	if err != nil {
		return "", false
	}
	info, statErr := file.Stat()
	file.Close()
	if statErr != nil || !info.IsDir() {
		return "", false
	}
	indexPath := strings.TrimRight(urlPath, "/") + "/index.html"
	index, err := frontendFS.Open(indexPath)
	if err != nil {
		return "", true
	}
	index.Close()
	return indexPath, true
}

func serveEmbeddedFile(c *gin.Context, frontendFS static.ServeFileSystem, urlPath string) {
	file, err := frontendFS.Open(urlPath)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || info.IsDir() {
		c.Status(http.StatusNotFound)
		return
	}
	body, err := io.ReadAll(file)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	contentType := mime.TypeByExtension(path.Ext(urlPath))
	if contentType == "" {
		contentType = http.DetectContentType(body)
	}
	header := c.Writer.Header()
	header.Set("Content-Type", contentType)
	header.Set("Content-Length", strconv.Itoa(len(body)))
	c.Writer.WriteHeader(http.StatusOK)
	if c.Request.Method != http.MethodHead {
		_, _ = c.Writer.Write(body)
	}
}
