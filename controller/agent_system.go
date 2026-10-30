package controller

import (
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
)

func GetAgentSystem(c *gin.Context) {
	userID := c.GetInt("id")
	report, err := model.GetAgentSystem(userID)
	if err != nil {
		logger.LogError(c, common.LogText("failed to load agent system: %s", err.Error()))
		common.ApiErrorT(c, "Failed to load agent system")
		return
	}
	common.ApiSuccess(c, report)
}

func TransferAgentCommission(c *gin.Context) {
	var request struct {
		Quota int `json:"quota"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		common.ApiErrorT(c, "Invalid parameters")
		return
	}
	result, err := model.TransferAgentCommission(c.GetInt("id"), request.Quota)
	if err != nil {
		logger.LogError(c, common.LogText("failed to transfer agent commission: %s", err.Error()))
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, result)
}

func AdminSearchAgentCommission(c *gin.Context) {
	accounts, err := model.SearchAgentCommissionAccounts(c.Query("keyword"), c.GetInt("role"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, accounts)
}

func AdminListAgentCommissionDeductions(c *gin.Context) {
	rows, err := model.ListAgentCommissionDeductions(20)
	if err != nil {
		logger.LogError(c, common.LogText("failed to list agent commission deductions: %s", err.Error()))
		common.ApiErrorT(c, "Failed to load commission deductions")
		return
	}
	common.ApiSuccess(c, rows)
}

func AdminDeductAgentCommission(c *gin.Context) {
	var request struct {
		UserId int    `json:"user_id"`
		Quota  int    `json:"quota"`
		Note   string `json:"note"`
	}
	if err := c.ShouldBindJSON(&request); err != nil || request.UserId <= 0 {
		common.ApiErrorT(c, "Invalid parameters")
		return
	}
	target, err := model.GetUserById(request.UserId, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if !canManageTargetRole(c.GetInt("role"), target.Role) {
		common.ApiErrorT(c, "No permission to access users of same or higher level")
		return
	}
	result, err := model.DeductAgentCommission(c.GetInt("id"), request.UserId, request.Quota, request.Note)
	if err != nil {
		logger.LogError(c, common.LogText("failed to deduct agent commission: %s", err.Error()))
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, result)
}

func AdminListApprovedAgents(c *gin.Context) {
	page, ok := agentQueryInt(c.Query("page"), 1)
	pageSize, sizeOK := agentQueryInt(c.Query("pageSize"), 20)
	if !ok || !sizeOK {
		common.ApiErrorT(c, "Invalid parameters")
		return
	}
	pageResult, err := model.ListApprovedAgents(c.GetInt("role"), page, pageSize)
	if err != nil {
		logger.LogError(c, common.LogText("failed to list approved agents: %s", err.Error()))
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, pageResult)
}

func AdminSetAgentCommissionRates(c *gin.Context) {
	var request struct {
		UserId     int  `json:"user_id"`
		Level1Rate *int `json:"level1_rate"`
		Level2Rate *int `json:"level2_rate"`
	}
	if err := c.ShouldBindJSON(&request); err != nil || request.UserId <= 0 || request.Level1Rate == nil || request.Level2Rate == nil {
		common.ApiErrorT(c, "Invalid parameters")
		return
	}
	target, err := model.GetUserById(request.UserId, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if !canManageTargetRole(c.GetInt("role"), target.Role) {
		common.ApiErrorT(c, "No permission to access users of same or higher level")
		return
	}
	result, err := model.SetAgentCommissionRates(request.UserId, *request.Level1Rate, *request.Level2Rate)
	if err != nil {
		logger.LogError(c, common.LogText("failed to set agent commission rates: %s", err.Error()))
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, result)
}

func GetAgentSystemStats(c *gin.Context) {
	days, ok := agentQueryInt(c.Query("days"), 7)
	if !ok {
		common.ApiErrorT(c, "Invalid parameters")
		return
	}
	report, err := model.GetAgentSystemStats(c.GetInt("id"), days)
	if err != nil {
		logger.LogError(c, common.LogText("failed to load agent statistics: %s", err.Error()))
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, report)
}

func GetAgentSystemRecords(c *gin.Context) {
	page, ok := agentQueryInt(c.Query("page"), 1)
	pageSize, sizeOK := agentQueryInt(c.Query("pageSize"), 10)
	if !ok || !sizeOK || (pageSize != 10 && pageSize != 20 && pageSize != 50) {
		common.ApiErrorT(c, "Invalid parameters")
		return
	}
	pageResult, err := model.ListAgentSystemRecords(c.GetInt("id"), c.Query("kind"), c.Query("keyword"), page, pageSize)
	if err != nil {
		logger.LogError(c, common.LogText("failed to load agent records: %s", err.Error()))
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, pageResult)
}

func GetAgentSystemUsers(c *gin.Context) {
	page, ok := agentQueryInt(c.Query("page"), 1)
	pageSize, sizeOK := agentQueryInt(c.Query("pageSize"), 10)
	if !ok || !sizeOK || (pageSize != 10 && pageSize != 20 && pageSize != 50) {
		common.ApiErrorT(c, "Invalid parameters")
		return
	}
	pageResult, err := model.ListAgentInviteesPage(c.GetInt("id"), c.Query("keyword"), page, pageSize)
	if err != nil {
		logger.LogError(c, common.LogText("failed to load invited users: %s", err.Error()))
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, pageResult)
}

func agentQueryInt(raw string, fallback int) (int, bool) {
	if raw == "" {
		return fallback, true
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed <= 0 {
		return 0, false
	}
	return parsed, true
}
