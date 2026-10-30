package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestRewriteCanvasImageRequestMarksFlatPrice(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.POST("/v1/canvas/images/generations", rewriteCanvasImageRequest("/v1/images/generations"), func(c *gin.Context) {
		require.True(t, common.GetContextKeyBool(c, constant.ContextKeyCanvasFlatImagePrice))
		require.Equal(t, "/v1/images/generations", c.Request.URL.Path)
		require.Equal(t, "/v1/images/generations?x=1", c.Request.RequestURI)
		c.Status(http.StatusNoContent)
	})
	engine.POST("/v1/canvas/images/edits", rewriteCanvasImageRequest("/v1/images/edits"), func(c *gin.Context) {
		require.True(t, common.GetContextKeyBool(c, constant.ContextKeyCanvasFlatImagePrice))
		require.Equal(t, "/v1/images/edits", c.Request.URL.Path)
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/canvas/images/generations?x=1", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code)

	req = httptest.NewRequest(http.MethodPost, "/v1/canvas/images/edits", nil)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code)
}
