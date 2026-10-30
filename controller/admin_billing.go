package controller

import (
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

type adminBillingRecord struct {
	model.TopUp
	Username  string `json:"username"`
	AgentName string `json:"agent_name"`
}

type billingPeriodSummary struct {
	Amount float64 `json:"amount"`
	Count  int64   `json:"count"`
}

// AdminGetBillingRecords returns filtered, paginated top-up orders for the admin ledger.
func AdminGetBillingRecords(c *gin.Context) {
	page := common.GetPageQuery(c)
	query := model.DB.Model(&model.TopUp{}).
		Joins("LEFT JOIN users AS billing_users ON billing_users.id = top_ups.user_id").
		Joins("LEFT JOIN users AS billing_agents ON billing_agents.id = billing_users.inviter_id")

	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		pattern := "%" + keyword + "%"
		if userID, err := strconv.Atoi(keyword); err == nil {
			query = query.Where("(top_ups.trade_no LIKE ? OR billing_users.username LIKE ? OR top_ups.user_id = ?)", pattern, pattern, userID)
		} else {
			query = query.Where("(top_ups.trade_no LIKE ? OR billing_users.username LIKE ?)", pattern, pattern)
		}
	}
	if method := strings.TrimSpace(c.Query("payment_method")); method != "" {
		query = query.Where("top_ups.payment_method = ?", method)
	}
	if status := strings.TrimSpace(c.Query("status")); status != "" {
		query = query.Where("top_ups.status = ?", status)
	}
	if start, ok := parseBillingTimestamp(c.Query("start_time")); ok {
		query = query.Where("top_ups.create_time >= ?", start)
	}
	if end, ok := parseBillingTimestamp(c.Query("end_time")); ok {
		query = query.Where("top_ups.create_time < ?", end)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		common.ApiError(c, err)
		return
	}
	var records []adminBillingRecord
	if err := query.Select("top_ups.*, billing_users.username AS username, billing_agents.username AS agent_name").
		Order("top_ups.id DESC").Limit(page.GetPageSize()).Offset(page.GetStartIdx()).Find(&records).Error; err != nil {
		common.ApiError(c, err)
		return
	}
	page.SetTotal(int(total))
	page.SetItems(records)
	common.ApiSuccess(c, page)
}

func parseBillingTimestamp(value string) (int64, bool) {
	if value == "" {
		return 0, false
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return 0, false
	}
	return parsed.Unix(), true
}

// AdminGetBillingSummary returns successful top-up totals for common calendar periods.
func AdminGetBillingSummary(c *gin.Context) {
	now := time.Now()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := dayStart.AddDate(0, 0, -((int(dayStart.Weekday()) + 6) % 7))
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	periods := map[string]time.Time{
		"all": {}, "month": monthStart, "week": weekStart,
		"last_week": weekStart.AddDate(0, 0, -7), "today": dayStart,
		"yesterday": dayStart.AddDate(0, 0, -1),
	}
	results := make(map[string]billingPeriodSummary, len(periods))
	for name, start := range periods {
		var result billingPeriodSummary
		query := model.DB.Model(&model.TopUp{}).Where("status = ?", common.TopUpStatusSuccess)
		if !start.IsZero() {
			query = query.Where("create_time >= ?", start.Unix())
			switch name {
			case "yesterday":
				query = query.Where("create_time < ?", dayStart.Unix())
			case "last_week":
				query = query.Where("create_time < ?", weekStart.Unix())
			}
		}
		if err := query.Select("COALESCE(SUM(money), 0) AS amount, COUNT(*) AS count").Scan(&result).Error; err != nil {
			common.ApiError(c, err)
			return
		}
		results[name] = result
	}
	common.ApiSuccess(c, results)
}
