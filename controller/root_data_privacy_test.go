package controller

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The controller receives an already authenticated identity. Request parameters
// must never override it; main and log storage deliberately use separate DBs.
func privacyRequest(t *testing.T, role int, route, path string, handler gin.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	router := gin.New()
	router.GET(route, func(c *gin.Context) {
		c.Set("role", role)
		id, username := 2, "privacy-admin"
		if role == common.RoleRootUser {
			id, username = 3, "privacy-root"
		}
		c.Set("id", id)
		c.Set("username", username)
		handler(c)
	})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
	return response
}

func privacyData(t *testing.T, response *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	require.NoError(t, common.Unmarshal(response.Body.Bytes(), &body))
	require.Equal(t, true, body["success"], response.Body.String())
	data, ok := body["data"].(map[string]any)
	require.True(t, ok, response.Body.String())
	return data
}

func TestRootDataPrivacyDatabaseMatrix(t *testing.T) {
	for _, dialect := range []string{"sqlite", "mysql", "postgres"} {
		t.Run(dialect, func(t *testing.T) {
			env := map[string]string{"mysql": "PRIVACY_MYSQL_DSN", "postgres": "PRIVACY_POSTGRES_DSN"}[dialect]
			dsn := os.Getenv(env)
			if env != "" && dsn == "" {
				t.Skip(env + " not configured")
			}
			db, _ := newAuditTestDatabase(t, dialect, dsn)
			logDB, _ := newAuditTestDatabase(t, dialect, dsn)
			oldDB, oldLogDB := model.DB, model.LOG_DB
			oldMain, oldLog := common.MainDatabaseType(), common.LogDatabaseType()
			oldRedis, oldMemory := common.RedisEnabled, common.MemoryCacheEnabled
			model.DB, model.LOG_DB = db, logDB
			common.RedisEnabled, common.MemoryCacheEnabled = false, false
			typ := map[string]common.DatabaseType{"sqlite": common.DatabaseTypeSQLite, "mysql": common.DatabaseTypeMySQL, "postgres": common.DatabaseTypePostgreSQL}[dialect]
			common.SetDatabaseTypes(typ, typ)
			t.Cleanup(func() {
				model.DB, model.LOG_DB = oldDB, oldLogDB
				common.SetDatabaseTypes(oldMain, oldLog)
				common.RedisEnabled, common.MemoryCacheEnabled = oldRedis, oldMemory
			})
			require.NoError(t, db.AutoMigrate(&model.User{}, &model.Task{}, &model.Midjourney{}, &model.QuotaData{}))
			require.NoError(t, logDB.AutoMigrate(&model.Log{}, &model.AuditLog{}))
			var version string
			versionQuery := "SELECT version()"
			if dialect == "sqlite" {
				versionQuery = "SELECT sqlite_version()"
			}
			require.NoError(t, db.Raw(versionQuery).Scan(&version).Error)
			t.Logf("engine=%s version=%s; separate primary and log databases", dialect, version)
			now := time.Now().Unix()
			for i, role := range []int{common.RoleCommonUser, common.RoleAdminUser, common.RoleRootUser, common.RoleRootUser} {
				id := i + 1
				username := []string{"privacy-member", "privacy-admin", "privacy-root", "privacy-deleted-root"}[i]
				u := model.User{Id: id, Username: username, AffCode: fmt.Sprintf("privacy-%d", id), Role: role, Status: common.UserStatusEnabled, Group: "default"}
				require.NoError(t, db.Create(&u).Error)
				if id == 4 {
					require.NoError(t, db.Delete(&u).Error)
				}
				// Renamed root rows intentionally carry a different historical username.
				logUsername := username
				if id == 3 {
					logUsername = "root-before-rename"
				}
				require.NoError(t, logDB.Create(&model.Log{UserId: id, Username: logUsername, CreatedAt: now - 5 + int64(id), Type: model.LogTypeConsume, Quota: id * 10, PromptTokens: id, CompletionTokens: id, Content: fmt.Sprintf("private-message-user-%d", id), RequestId: fmt.Sprintf("privacy-request-%d", id), UpstreamRequestId: fmt.Sprintf("privacy-upstream-%d", id)}).Error)
				require.NoError(t, db.Create(&model.Task{TaskID: fmt.Sprintf("privacy-task-%d", id), UserId: id, Platform: "privacy-test", Status: model.TaskStatusSuccess, Progress: "100%"}).Error)
				require.NoError(t, db.Create(&model.Midjourney{MjId: fmt.Sprintf("privacy-mj-%d", id), UserId: id, Prompt: fmt.Sprintf("private-message-user-%d", id)}).Error)
				require.NoError(t, db.Create(&model.QuotaData{UserID: id, Username: logUsername, ModelName: "privacy-model", UseGroup: "default", CreatedAt: now, Count: 1, Quota: id * 10, TokenUsed: id * 2}).Error)
				// Historical events can predate promotion, so actor role alone is insufficient.
				require.NoError(t, logDB.Create(&model.AuditLog{EventId: fmt.Sprintf("privacy-event-%d", id), UserId: id, Username: logUsername, ActorRole: common.RoleCommonUser, CreatedAt: now, Content: fmt.Sprintf("private-message-user-%d", id)}).Error)
			}
			// Root actors' historical events remain hidden even after demotion.
			require.NoError(t, logDB.Create(&model.AuditLog{EventId: "root-actor-on-member", UserId: 1, Username: "privacy-member", ActorRole: common.RoleRootUser, CreatedAt: now, Content: "private-root-actor-message"}).Error)
			listCases := []struct {
				name, route, path string
				handler           gin.HandlerFunc
			}{
				{"usage logs", "/api/log/", "/api/log/?page_size=1", GetAllLogs},
				{"accounts", "/api/user/", "/api/user/?page_size=1", GetAllUsers},
				{"account search", "/api/user/search", "/api/user/search?keyword=privacy&page_size=1", SearchUsers},
				{"tasks", "/api/task", "/api/task?page_size=1", GetAllTask},
				{"midjourney", "/api/mj/", "/api/mj/?page_size=1", GetAllMidjourney},
				{"audit", "/api/audit", "/api/audit?page_size=1", GetAuditLogs},
			}
			for _, tc := range listCases {
				t.Run(tc.name, func(t *testing.T) {
					for _, role := range []int{common.RoleAdminUser, common.RoleRootUser} {
						wantTotal := 2
						if role == common.RoleRootUser {
							wantTotal = 4
							if tc.name == "audit" {
								wantTotal = 5
							}
						}
						for _, page := range []int{1, 2, 3} {
							response := privacyRequest(t, role, tc.route, tc.path+fmt.Sprintf("&p=%d", page), tc.handler)
							data := privacyData(t, response)
							assert.EqualValues(t, wantTotal, data["total"])
							if role == common.RoleAdminUser {
								for _, marker := range []string{"private-message-user-3", "private-message-user-4", "private-root-actor-message", "privacy-task-3", "privacy-task-4", "privacy-mj-3", "privacy-mj-4", "privacy-root", "privacy-deleted-root", "root-before-rename"} {
									assert.NotContains(t, response.Body.String(), marker)
								}
								if page <= 2 {
									assert.Len(t, data["items"], 1)
								} else {
									assert.Empty(t, data["items"])
								}
							}
						}
					}
				})
			}
			t.Run("targeted bypasses", func(t *testing.T) {
				for _, query := range []string{"username=root-before-rename", "request_id=privacy-request-3", "upstream_request_id=privacy-upstream-3", "username=privacy-deleted-root", "username=%25root%25&role=100"} {
					response := privacyRequest(t, common.RoleAdminUser, "/api/log/", "/api/log/?"+query, GetAllLogs)
					data := privacyData(t, response)
					assert.EqualValues(t, 0, data["total"])
					assert.Empty(t, data["items"])
				}
				response := privacyRequest(t, common.RoleAdminUser, "/api/user/search", "/api/user/search?keyword=3&role=100", SearchUsers)
				assert.EqualValues(t, 0, privacyData(t, response)["total"])
				response = privacyRequest(t, common.RoleAdminUser, "/api/task", "/api/task?task_id=privacy-task-3", GetAllTask)
				assert.EqualValues(t, 0, privacyData(t, response)["total"])
				response = privacyRequest(t, common.RoleAdminUser, "/api/mj/", "/api/mj/?mj_id=privacy-mj-4", GetAllMidjourney)
				assert.EqualValues(t, 0, privacyData(t, response)["total"])
			})
			t.Run("statistics and self scope", func(t *testing.T) {
				for _, tc := range []struct {
					role            int
					route, path     string
					handler         gin.HandlerFunc
					quota, rpm, tpm int
				}{
					{10, "/api/log/stat", "/api/log/stat", GetLogsStat, 30, 2, 6},
					{100, "/api/log/stat", "/api/log/stat", GetLogsStat, 100, 4, 20},
					{10, "/api/log/stat", "/api/log/stat?username=root-before-rename", GetLogsStat, 0, 0, 0},
					{100, "/api/log/self/stat", "/api/log/self/stat?username=privacy-member", GetLogsSelfStat, 30, 1, 6},
					{10, "/api/log/self/stat", "/api/log/self/stat?username=privacy-root", GetLogsSelfStat, 20, 1, 4},
				} {
					data := privacyData(t, privacyRequest(t, tc.role, tc.route, tc.path, tc.handler))
					assert.EqualValues(t, tc.quota, data["quota"])
					assert.EqualValues(t, tc.rpm, data["rpm"])
					assert.EqualValues(t, tc.tpm, data["tpm"])
				}
			})
			t.Run("dashboard summaries", func(t *testing.T) {
				query := fmt.Sprintf("?start_timestamp=%d&end_timestamp=%d", now-10, now+10)
				for _, tc := range []struct {
					route   string
					handler gin.HandlerFunc
				}{
					{"/api/data/", GetAllQuotaDates}, {"/api/data/users", GetQuotaDatesByUser}, {"/api/data/flow", GetAllFlowQuotaDates},
				} {
					for _, role := range []int{10, 100} {
						response := privacyRequest(t, role, tc.route, tc.route+query, tc.handler)
						var body struct {
							Success bool
							Data    []model.QuotaData
						}
						require.NoError(t, common.Unmarshal(response.Body.Bytes(), &body))
						require.True(t, body.Success, response.Body.String())
						sum := 0
						for _, row := range body.Data {
							sum += row.Quota
						}
						want := 30
						if role == 100 {
							want = 100
						}
						assert.Equal(t, want, sum)
						if role == 10 {
							assert.NotContains(t, response.Body.String(), "root-before-rename")
							assert.NotContains(t, response.Body.String(), "privacy-deleted-root")
						}
					}
				}
			})
			t.Run("artifact access", func(t *testing.T) {
				for _, id := range []int{1, 3, 4} {
					response := privacyRequest(t, common.RoleAdminUser, "/api/task/:task_id/artifacts", fmt.Sprintf("/api/task/privacy-task-%d/artifacts", id), GetDashboardTaskArtifacts)
					if id == 1 {
						assert.Equal(t, http.StatusOK, response.Code)
					} else {
						assert.Equal(t, http.StatusNotFound, response.Code)
						assert.NotContains(t, response.Body.String(), fmt.Sprintf("privacy-task-%d", id))
					}
				}
				response := privacyRequest(t, common.RoleRootUser, "/api/task/:task_id/artifacts", "/api/task/privacy-task-3/artifacts", GetDashboardTaskArtifacts)
				assert.Equal(t, http.StatusOK, response.Code)
			})
			t.Run("failed protection lookup must fail closed", func(t *testing.T) {
				old := model.DB
				// A log database can remain healthy while the primary user lookup fails.
				model.DB = logDB
				t.Cleanup(func() { model.DB = old })
				response := privacyRequest(t, common.RoleAdminUser, "/api/log/", "/api/log/", GetAllLogs)
				var body map[string]any
				require.NoError(t, common.Unmarshal(response.Body.Bytes(), &body))
				assert.Equal(t, false, body["success"])
				assert.NotContains(t, response.Body.String(), "private-message")
			})
		})
	}
}
