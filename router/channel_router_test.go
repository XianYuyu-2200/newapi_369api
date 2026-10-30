package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/controller"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/service/authz"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestChannelDefaultBaseURLsRequireReadPermission(t *testing.T) {
	assertChannelRoutePermission(t, http.MethodGet, "/default_base_urls", authz.ChannelRead, controller.GetChannelDefaultBaseURLs)

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	registerChannelRoutes(engine.Group("/api"))
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/channel/default_base_urls", nil))
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestChannelStatusRoutesUseExpectedPermissions(t *testing.T) {
	assertChannelRoutePermission(t, http.MethodGet, "/:id/vllm/status", authz.ChannelRead, controller.GetVLLMChannelStatus)
	assertChannelRoutePermission(t, http.MethodGet, "/:id/sglang/status", authz.ChannelRead, controller.GetSGLangChannelStatus)
	assertChannelRoutePermission(t, http.MethodPost, "/:id/status", authz.ChannelOperate, controller.UpdateChannelStatus)
	assertChannelRoutePermission(t, http.MethodPost, "/status/batch", authz.ChannelOperate, controller.BatchUpdateChannelStatus)
	assertChannelRoutePermission(t, http.MethodPut, "/", authz.ChannelWrite, controller.UpdateChannel)
}

func TestChannelDeleteRoutesUseSensitiveWritePermission(t *testing.T) {
	assertChannelRoutePermission(t, http.MethodDelete, "/:id", authz.ChannelSensitiveWrite, controller.DeleteChannel)
	assertChannelRoutePermission(t, http.MethodPost, "/batch", authz.ChannelSensitiveWrite, controller.DeleteChannelBatch)
	assertChannelRoutePermission(t, http.MethodDelete, "/disabled", authz.ChannelSensitiveWrite, controller.DeleteDisabledChannel)
	assertChannelRoutePermission(t, http.MethodPut, "/", authz.ChannelWrite, controller.UpdateChannel)
	assertChannelRoutePermission(t, http.MethodPut, "/tag", authz.ChannelWrite, controller.EditTagChannels)
	assertChannelRoutePermission(t, http.MethodPost, "/batch/tag", authz.ChannelWrite, controller.BatchSetChannelTag)
}

func TestChannelStatusRoutesRegisterWithoutConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	api := engine.Group("/api")

	require.NotPanics(t, func() {
		registerChannelRoutes(api)
	})
}

func assertChannelRoutePermission(t *testing.T, method string, path string, permission authz.Permission, handler any) {
	t.Helper()
	for _, route := range channelPermissionRoutes {
		if route.method == method && route.path == path {
			assert.Equal(t, permission, route.permission)
			assert.Equal(t, reflect.ValueOf(handler).Pointer(), reflect.ValueOf(route.handler).Pointer())
			return
		}
	}
	t.Fatalf("route %s %s not found", method, path)
}

// Channel management allows administrators; model management remains root-only.
func TestChannelAccessAndRootOnlyModelManagement(t *testing.T) {
	previousDB, previousLogDB := model.DB, model.LOG_DB
	previousType, previousRedis := common.MainDatabaseType(), common.RedisEnabled
	previousSecret, previousMaster := common.SessionSecret, common.IsMasterNode
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.UserSession{}, &model.AuditLog{}, &model.UserAccessToken{}, &model.Option{}, &model.CasbinRule{}, &model.AuthzRole{}, &model.PrefillGroup{}, &model.Vendor{}, &model.Model{}, &model.Ability{}, &model.Channel{}))
	model.DB, model.LOG_DB = db, db
	common.SetMainDatabaseType(common.DatabaseTypeSQLite)
	common.RedisEnabled, common.IsMasterNode = false, true
	common.SessionSecret = "root-management-regression-secret"
	require.NoError(t, authz.Init(db))
	t.Cleanup(func() {
		model.DB, model.LOG_DB = previousDB, previousLogDB
		common.SetMainDatabaseType(previousType)
		common.RedisEnabled, common.IsMasterNode = previousRedis, previousMaster
		common.SessionSecret = previousSecret
		require.NoError(t, sqlDB.Close())
	})
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	SetApiRouter(engine)
	baseURL := "https://upstream.example.com"
	channel := model.Channel{Type: 1, Name: "editable-channel", Key: "original-channel-key", BaseURL: &baseURL, Models: "test-model", Group: "default", Status: common.ChannelStatusEnabled}
	require.NoError(t, channel.Insert())
	for _, role := range []int{common.RoleCommonUser, common.RoleAdminUser, common.RoleRootUser} {
		name := fmt.Sprintf("management-%d", role)
		user := model.User{Username: name, Password: "unused-hash", Role: role, Status: common.UserStatusEnabled, Group: "default", AuthVersion: 1, AffCode: name}
		require.NoError(t, db.Create(&user).Error)
		bundle, err := service.CreateLoginSession(user.Id, "password", "127.0.0.1", "regression")
		require.NoError(t, err)
		// Public model discovery must remain usable by all authenticated roles.
		discovery := httptest.NewRequest(http.MethodGet, "/api/models", nil)
		discovery.Header.Set("Authorization", "Bearer "+bundle.AccessToken)
		discovery.Header.Set("New-Api-User", strconv.Itoa(user.Id))
		response := httptest.NewRecorder()
		engine.ServeHTTP(response, discovery)
		require.Equal(t, http.StatusOK, response.Code)
		assert.Contains(t, response.Body.String(), `"success":true`)

		for _, path := range []string{"/api/channel/default_base_urls", "/api/models/missing", "/api/vendors/", "/api/deployments/settings", "/api/prefill_group/"} {
			t.Run(name+path, func(t *testing.T) {
				request := httptest.NewRequest(http.MethodGet, path, nil)
				request.Header.Set("Authorization", "Bearer "+bundle.AccessToken)
				request.Header.Set("New-Api-User", strconv.Itoa(user.Id))
				recorder := httptest.NewRecorder()
				engine.ServeHTTP(recorder, request)
				expected := http.StatusForbidden
				if role == common.RoleRootUser || (role == common.RoleAdminUser && strings.HasPrefix(path, "/api/channel/")) {
					expected = http.StatusOK
				}
				require.Equal(t, expected, recorder.Code, recorder.Body.String())
				if expected == http.StatusOK {
					assert.Contains(t, recorder.Body.String(), `"success":true`)
				}
			})
		}
		t.Run(name+"/edit-channel", func(t *testing.T) {
			newURL := fmt.Sprintf("https://upstream-%d.example.com", role)
			newKey := fmt.Sprintf("replacement-key-%d", role)
			payload := fmt.Sprintf(`{"id":%d,"type":1,"name":"updated-channel","models":"test-model","group":"default","base_url":%q,"key":%q}`, channel.Id, newURL, newKey)
			request := httptest.NewRequest(http.MethodPut, "/api/channel/", strings.NewReader(payload))
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Authorization", "Bearer "+bundle.AccessToken)
			request.Header.Set("New-Api-User", strconv.Itoa(user.Id))
			recorder := httptest.NewRecorder()
			engine.ServeHTTP(recorder, request)
			stored, err := model.GetChannelById(channel.Id, true)
			require.NoError(t, err)
			if role == common.RoleCommonUser {
				require.Equal(t, http.StatusForbidden, recorder.Code)
				assert.Equal(t, baseURL, *stored.BaseURL)
				assert.Equal(t, "original-channel-key", stored.Key)
				return
			}
			require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
			assert.Contains(t, recorder.Body.String(), `"success":true`)
			assert.NotContains(t, recorder.Body.String(), newKey)
			assert.Equal(t, newURL, *stored.BaseURL)
			assert.Equal(t, newKey, stored.Key)
			assert.Equal(t, "updated-channel", stored.Name)
			if role != common.RoleAdminUser {
				return
			}
			keyRequest := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/channel/%d/key", channel.Id), nil)
			keyRequest.Header.Set("Authorization", "Bearer "+bundle.AccessToken)
			keyRequest.Header.Set("New-Api-User", strconv.Itoa(user.Id))
			keyResponse := httptest.NewRecorder()
			engine.ServeHTTP(keyResponse, keyRequest)
			require.Equal(t, http.StatusForbidden, keyResponse.Code)
			assert.NotContains(t, keyResponse.Body.String(), newKey)

			require.NoError(t, authz.SetUserPermissions(user.Id, authz.PermissionsMap{authz.ResourceChannel: {authz.ActionSensitiveWrite: false}}))
			request = httptest.NewRequest(http.MethodPut, "/api/channel/", strings.NewReader(strings.ReplaceAll(payload, newKey, "denied-key")))
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Authorization", "Bearer "+bundle.AccessToken)
			request.Header.Set("New-Api-User", strconv.Itoa(user.Id))
			recorder = httptest.NewRecorder()
			engine.ServeHTTP(recorder, request)
			assert.Contains(t, recorder.Body.String(), `"success":false`)
			stored, err = model.GetChannelById(channel.Id, true)
			require.NoError(t, err)
			assert.Equal(t, newKey, stored.Key)
			require.NoError(t, authz.ClearUserPermissions(user.Id))
		})
		if role == common.RoleAdminUser {
			// Check every registered method, not only the navigation's GETs.
			for _, route := range engine.Routes() {
				restricted := false
				for _, prefix := range []string{"/api/models/", "/api/vendors/", "/api/deployments/", "/api/prefill_group/"} {
					if strings.HasPrefix(route.Path, prefix) {
						restricted = true
						break
					}
				}
				if !restricted {
					continue
				}
				path := route.Path
				for _, parameter := range []string{":id", ":container_id"} {
					path = strings.ReplaceAll(path, parameter, "1")
				}
				request := httptest.NewRequest(route.Method, path, strings.NewReader(`{}`))
				request.Header.Set("Content-Type", "application/json")
				request.Header.Set("Authorization", "Bearer "+bundle.AccessToken)
				request.Header.Set("New-Api-User", strconv.Itoa(user.Id))
				recorder := httptest.NewRecorder()
				engine.ServeHTTP(recorder, request)
				require.Equal(t, http.StatusForbidden, recorder.Code, route.Method+" "+route.Path)
			}
		}
	}
}
