package model

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/QuantumNous/new-api/common"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	agentShanghaiOffset = 8 * 3600
	agentSearchLimit    = 64
	agentMaxPageSize    = 100
)

var agentShanghaiLocation = time.FixedZone("Asia/Shanghai", agentShanghaiOffset)

// AgentStatsPoint is one Beijing-time day of invitee top-ups.
type AgentStatsPoint struct {
	Date            string  `json:"date"`
	Label           string  `json:"label"`
	Orders          int64   `json:"orders"`
	Success         int64   `json:"success"`
	Pending         int64   `json:"pending"`
	Failed          int64   `json:"failed"`
	RechargeQuota   int     `json:"rechargeQuota"`
	CommissionQuota int     `json:"commissionQuota"`
	Rate            float64 `json:"rate"`
}

// AgentPaymentSlice is successful recharge quota grouped by the stored method.
type AgentPaymentSlice struct {
	Method        string `json:"method"`
	Count         int64  `json:"count"`
	RechargeQuota int    `json:"rechargeQuota"`
}

// AgentStatsReport covers the selected number of Beijing-time days, including today.
type AgentStatsReport struct {
	Days            int                 `json:"days"`
	OrderCount      int64               `json:"orderCount"`
	SuccessCount    int64               `json:"successCount"`
	PendingCount    int64               `json:"pendingCount"`
	FailedCount     int64               `json:"failedCount"`
	SuccessRate     float64             `json:"successRate"`
	RechargeQuota   int                 `json:"rechargeQuota"`
	CommissionQuota int                 `json:"commissionQuota"`
	Points          []AgentStatsPoint   `json:"points"`
	Payments        []AgentPaymentSlice `json:"payments"`
}

// AgentOrderRecord is one invitee top-up. Commission is zero unless it succeeded.
type AgentOrderRecord struct {
	Id              int    `json:"id"`
	TradeNo         string `json:"tradeNo"`
	UserId          int    `json:"userId"`
	Username        string `json:"username"`
	DisplayName     string `json:"displayName"`
	Email           string `json:"email"`
	Level           int    `json:"level"`
	Status          string `json:"status"`
	PaymentMethod   string `json:"paymentMethod"`
	Amount          int64  `json:"amount"`
	RechargeQuota   int    `json:"rechargeQuota"`
	CommissionQuota int    `json:"commissionQuota"`
	CreatedAt       int64  `json:"createdAt"`
	CompletedAt     int64  `json:"completedAt"`
}

// AgentRecordPage is a page of recharge or commission rows plus unfiltered period totals.
// Period totals sum successful amounts per stored percent, then floor once per percent.
// Each row floors its own commission, so the rows can add up to slightly less.
type AgentRecordPage struct {
	Kind                     string             `json:"kind"`
	Page                     int                `json:"page"`
	PageSize                 int                `json:"pageSize"`
	Total                    int64              `json:"total"`
	TodayCommissionQuota     int                `json:"todayCommissionQuota"`
	YesterdayCommissionQuota int                `json:"yesterdayCommissionQuota"`
	WeekCommissionQuota      int                `json:"weekCommissionQuota"`
	Records                  []AgentOrderRecord `json:"records"`
}

// AgentInviteePage is one page of level-1 and level-2 invitees.
type AgentInviteePage struct {
	Page     int            `json:"page"`
	PageSize int            `json:"pageSize"`
	Total    int64          `json:"total"`
	Users    []AgentInvitee `json:"users"`
}

type agentDayStat struct {
	DayKey  int64
	Status  string
	Level   int
	Percent int
	Amount  int64
	Count   int64
}

type agentPayStat struct {
	Method   string
	Provider string
	Amount   int64
	Count    int64
}

type agentOrderRow struct {
	Id              int
	TradeNo         string
	UserId          int
	Username        string
	DisplayName     string
	Email           string
	InviterId       int
	Status          string
	PaymentMethod   string
	PaymentProvider string
	Amount          int64
	CreateTime      int64
	CompleteTime    int64
	Level1Percent   sql.NullInt64
	Level2Percent   sql.NullInt64
}

func GetAgentSystemStats(userID int, days int) (AgentStatsReport, error) {
	return getAgentSystemStatsAt(userID, days, time.Now())
}

func getAgentSystemStatsAt(userID int, days int, now time.Time) (AgentStatsReport, error) {
	report := AgentStatsReport{
		Points:   []AgentStatsPoint{},
		Payments: []AgentPaymentSlice{},
	}
	days, err := normalizeAgentDays(days)
	if err != nil {
		return report, err
	}
	report.Days = days
	if userID <= 0 {
		return report, common.NewMessage("Invalid parameters")
	}
	start, end := agentRange(now, days)
	rows, err := queryAgentDayStats(userID, start, end)
	if err != nil {
		return report, err
	}
	points := make(map[int64]*AgentStatsPoint, days)
	startKey := agentDayKeyFromUnix(start)
	for i := 0; i < days; i++ {
		key := startKey + int64(i)
		date, label := agentDayLabel(key)
		points[key] = &AgentStatsPoint{Date: date, Label: label}
	}
	var level1Amount int64
	var level2Amount int64
	for _, row := range rows {
		point := points[row.DayKey]
		if point == nil {
			continue
		}
		point.Orders += row.Count
		switch row.Status {
		case common.TopUpStatusSuccess:
			point.Success += row.Count
			if row.Level == 1 {
				level1Amount += row.Amount
			} else {
				level2Amount += row.Amount
			}
		case common.TopUpStatusPending:
			point.Pending += row.Count
		default:
			point.Failed += row.Count
		}
	}
	// Daily commission is floored once per day, level, and stored percent.
	// Recharge for that day is the summed amount, floored once.
	type dayLevelPercent struct {
		Day     int64
		Level   int
		Percent int
	}
	grouped := map[dayLevelPercent]int64{}
	dayAmounts := map[int64]int64{}
	for _, row := range rows {
		if row.Status != common.TopUpStatusSuccess || points[row.DayKey] == nil {
			continue
		}
		dayAmounts[row.DayKey] += row.Amount
		key := dayLevelPercent{Day: row.DayKey, Level: row.Level, Percent: row.Percent}
		grouped[key] += row.Amount
	}
	dayCommission := map[int64]int{}
	for key, amount := range grouped {
		part, partErr := agentCommissionQuota(amount, key.Percent)
		if partErr != nil {
			return report, partErr
		}
		next, addErr := addAgentQuota(dayCommission[key.Day], part)
		if addErr != nil {
			return report, addErr
		}
		dayCommission[key.Day] = next
	}
	for key, amount := range dayAmounts {
		recharge, rechargeErr := agentRechargeQuota(amount)
		if rechargeErr != nil {
			return report, rechargeErr
		}
		points[key].RechargeQuota = recharge
		points[key].CommissionQuota = dayCommission[key]
	}
	report.Points = make([]AgentStatsPoint, 0, days)
	for i := 0; i < days; i++ {
		point := points[startKey+int64(i)]
		point.Rate = agentRate(point.Success, point.Orders)
		report.Points = append(report.Points, *point)
		report.OrderCount += point.Orders
		report.SuccessCount += point.Success
		report.PendingCount += point.Pending
		report.FailedCount += point.Failed
	}
	report.SuccessRate = agentRate(report.SuccessCount, report.OrderCount)
	report.RechargeQuota, err = agentRechargeQuota(level1Amount + level2Amount)
	if err != nil {
		return report, err
	}
	level1Commission, _, _, err := sumAgentLevelCommission(DB, userID, 1, &start, &end)
	if err != nil {
		return report, err
	}
	level2Commission, _, _, err := sumAgentLevelCommission(DB, userID, 2, &start, &end)
	if err != nil {
		return report, err
	}
	report.CommissionQuota, err = addAgentQuota(level1Commission, level2Commission)
	if err != nil {
		return report, err
	}
	payments, err := queryAgentPaymentStats(userID, start, end)
	if err != nil {
		return report, err
	}
	merged := map[string]*AgentPaymentSlice{}
	for _, row := range payments {
		label := agentPaymentLabel(row.Method, row.Provider)
		slice := merged[label]
		if slice == nil {
			slice = &AgentPaymentSlice{Method: label}
			merged[label] = slice
		}
		slice.Count += row.Count
		quota, quotaErr := agentRechargeQuota(row.Amount)
		if quotaErr != nil {
			return report, quotaErr
		}
		slice.RechargeQuota, quotaErr = addAgentQuota(slice.RechargeQuota, quota)
		if quotaErr != nil {
			return report, quotaErr
		}
	}
	report.Payments = make([]AgentPaymentSlice, 0, len(merged))
	for _, slice := range merged {
		report.Payments = append(report.Payments, *slice)
	}
	sort.Slice(report.Payments, func(i, j int) bool {
		if report.Payments[i].RechargeQuota != report.Payments[j].RechargeQuota {
			return report.Payments[i].RechargeQuota > report.Payments[j].RechargeQuota
		}
		if report.Payments[i].Count != report.Payments[j].Count {
			return report.Payments[i].Count > report.Payments[j].Count
		}
		return report.Payments[i].Method < report.Payments[j].Method
	})
	return report, nil
}

func ListAgentSystemRecords(userID int, kind string, keyword string, page int, pageSize int) (AgentRecordPage, error) {
	return listAgentSystemRecordsAt(userID, kind, keyword, page, pageSize, time.Now())
}

func listAgentSystemRecordsAt(userID int, kind string, keyword string, page int, pageSize int, now time.Time) (AgentRecordPage, error) {
	pageResult := AgentRecordPage{Records: []AgentOrderRecord{}}
	if userID <= 0 || (kind != "recharge" && kind != "commission") {
		return pageResult, common.NewMessage("Invalid parameters")
	}
	page, pageSize, err := normalizeAgentPage(page, pageSize)
	if err != nil {
		return pageResult, err
	}
	keyword, err = normalizeAgentSearch(keyword)
	if err != nil {
		return pageResult, err
	}
	filtered := agentOrderQuery(userID, kind, keyword)
	var total int64
	if err := filtered.Count(&total).Error; err != nil {
		return pageResult, err
	}
	rows := make([]agentOrderRow, 0)
	if total > 0 {
		err = agentOrderQuery(userID, kind, keyword).
			Select(
				"top_ups.id",
				"top_ups.trade_no",
				"top_ups.user_id",
				"users.username",
				"users.display_name",
				"users.email",
				"users.inviter_id",
				"top_ups.status",
				"top_ups.payment_method",
				"top_ups.payment_provider",
				"top_ups.amount",
				"top_ups.create_time",
				"top_ups.complete_time",
				"top_ups.level1_commission_percent AS level1_percent",
				"top_ups.level2_commission_percent AS level2_percent",
			).
			Order(clause.OrderByColumn{Column: clause.Column{Name: agentEventUnixSQL() + " DESC, top_ups.id DESC", Raw: true}}).
			Offset((page - 1) * pageSize).
			Limit(pageSize).
			Scan(&rows).Error
		if err != nil {
			return pageResult, err
		}
	}
	records := make([]AgentOrderRecord, 0, len(rows))
	for _, row := range rows {
		level := 2
		if row.InviterId == userID {
			level = 1
		}
		recharge, rechargeErr := agentRechargeQuota(row.Amount)
		if rechargeErr != nil {
			return pageResult, rechargeErr
		}
		storedPercent := row.Level2Percent
		if level == 1 {
			storedPercent = row.Level1Percent
		}
		commission, commissionErr := orderCommission(level, row.Status, row.Amount, storedPercent)
		if commissionErr != nil {
			return pageResult, commissionErr
		}
		records = append(records, AgentOrderRecord{
			Id:              row.Id,
			TradeNo:         row.TradeNo,
			UserId:          row.UserId,
			Username:        row.Username,
			DisplayName:     row.DisplayName,
			Email:           row.Email,
			Level:           level,
			Status:          row.Status,
			PaymentMethod:   agentPaymentLabel(row.PaymentMethod, row.PaymentProvider),
			Amount:          row.Amount,
			RechargeQuota:   recharge,
			CommissionQuota: commission,
			CreatedAt:       row.CreateTime,
			CompletedAt:     row.CompleteTime,
		})
	}
	todayStart, tomorrow := agentRange(now, 1)
	yesterdayStart := todayStart - 86400
	weekStart, _ := agentRange(now, 7)
	todayCommission, err := agentCommissionInRange(DB, userID, todayStart, tomorrow)
	if err != nil {
		return pageResult, err
	}
	yesterdayCommission, err := agentCommissionInRange(DB, userID, yesterdayStart, todayStart)
	if err != nil {
		return pageResult, err
	}
	weekCommission, err := agentCommissionInRange(DB, userID, weekStart, tomorrow)
	if err != nil {
		return pageResult, err
	}
	return AgentRecordPage{
		Kind:                     kind,
		Page:                     page,
		PageSize:                 pageSize,
		Total:                    total,
		TodayCommissionQuota:     todayCommission,
		YesterdayCommissionQuota: yesterdayCommission,
		WeekCommissionQuota:      weekCommission,
		Records:                  records,
	}, nil
}

func ListAgentInviteesPage(userID int, keyword string, page int, pageSize int) (AgentInviteePage, error) {
	result := AgentInviteePage{Users: []AgentInvitee{}}
	if userID <= 0 {
		return result, common.NewMessage("Invalid parameters")
	}
	page, pageSize, err := normalizeAgentPage(page, pageSize)
	if err != nil {
		return result, err
	}
	keyword, err = normalizeAgentSearch(keyword)
	if err != nil {
		return result, err
	}
	filtered := agentInviteeSearch(userID, keyword)
	var total int64
	if err := filtered.Count(&total).Error; err != nil {
		return result, err
	}
	rows := make([]agentInviteeRow, 0)
	if total > 0 {
		err = agentInviteeSearch(userID, keyword).
			Select("id", "username", "display_name", "email", "status", "created_at", "inviter_id").
			Order("created_at DESC, id DESC").
			Offset((page - 1) * pageSize).
			Limit(pageSize).
			Find(&rows).Error
		if err != nil {
			return result, err
		}
	}
	users := make([]AgentInvitee, 0, len(rows))
	for _, row := range rows {
		level := 2
		if row.InviterId == userID {
			level = 1
		}
		users = append(users, AgentInvitee{
			Id:          row.Id,
			Username:    row.Username,
			DisplayName: row.DisplayName,
			Email:       row.Email,
			Status:      row.Status,
			Level:       level,
			CreatedAt:   row.CreatedAt,
		})
	}
	if err := attachInviteeRecharge(users); err != nil {
		return result, err
	}
	return AgentInviteePage{
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		Users:    users,
	}, nil
}

func agentInviteeTopUps(tx *gorm.DB, userID int) *gorm.DB {
	return tx.Model(&TopUp{}).
		Joins("JOIN users ON users.id = top_ups.user_id AND users.deleted_at IS NULL").
		Where(
			"users.id <> ? AND (users.inviter_id = ? OR (users.second_inviter_id = ? AND users.inviter_id <> ?))",
			userID,
			userID,
			userID,
			userID,
		)
}

func agentOrderQuery(userID int, kind string, keyword string) *gorm.DB {
	query := agentInviteeTopUps(DB, userID)
	if kind == "commission" {
		query = query.Where("top_ups.status = ?", common.TopUpStatusSuccess)
	}
	if keyword != "" {
		like := agentLike(keyword)
		query = query.Where(
			"(users.username LIKE ? ESCAPE '!' OR users.display_name LIKE ? ESCAPE '!' OR users.email LIKE ? ESCAPE '!' OR top_ups.trade_no LIKE ? ESCAPE '!')",
			like,
			like,
			like,
			like,
		)
	}
	return query
}

func agentInviteeSearch(userID int, keyword string) *gorm.DB {
	query := agentInviteeBase(userID)
	if keyword == "" {
		return query
	}
	like := agentLike(keyword)
	return query.Where(
		"(username LIKE ? ESCAPE '!' OR display_name LIKE ? ESCAPE '!' OR email LIKE ? ESCAPE '!')",
		like,
		like,
		like,
	)
}

func queryAgentDayStats(userID int, start int64, end int64) ([]agentDayStat, error) {
	event := agentEventUnixSQL()
	day := agentDayKeySQL()
	level := fmt.Sprintf("CASE WHEN users.inviter_id = %d THEN 1 ELSE 2 END", userID)
	percent := fmt.Sprintf(
		"CASE WHEN users.inviter_id = %d THEN (%s) ELSE (%s) END",
		userID,
		agentStoredPercentSQL("top_ups.level1_commission_percent", agentLevel1CommissionPercent),
		agentStoredPercentSQL("top_ups.level2_commission_percent", agentLevel2CommissionPercent),
	)
	sql := fmt.Sprintf(`
SELECT %s AS day_key,
       top_ups.status AS status,
       %s AS level,
       %s AS percent,
       COALESCE(SUM(top_ups.amount), 0) AS amount,
       COUNT(top_ups.id) AS count
FROM top_ups
JOIN users ON users.id = top_ups.user_id AND users.deleted_at IS NULL
WHERE users.id <> ?
  AND (users.inviter_id = ? OR (users.second_inviter_id = ? AND users.inviter_id <> ?))
  AND (%s) >= ? AND (%s) < ?
GROUP BY %s, top_ups.status, %s, %s`, day, level, percent, event, event, day, level, percent)
	rows := make([]agentDayStat, 0)
	err := DB.Raw(sql, userID, userID, userID, userID, start, end).Scan(&rows).Error
	return rows, err
}

func queryAgentPaymentStats(userID int, start int64, end int64) ([]agentPayStat, error) {
	event := agentEventUnixSQL()
	sql := fmt.Sprintf(`
SELECT COALESCE(top_ups.payment_method, '') AS method,
       COALESCE(top_ups.payment_provider, '') AS provider,
       COALESCE(SUM(top_ups.amount), 0) AS amount,
       COUNT(top_ups.id) AS count
FROM top_ups
JOIN users ON users.id = top_ups.user_id AND users.deleted_at IS NULL
WHERE users.id <> ?
  AND (users.inviter_id = ? OR (users.second_inviter_id = ? AND users.inviter_id <> ?))
  AND top_ups.status = ?
  AND (%s) >= ? AND (%s) < ?
GROUP BY top_ups.payment_method, top_ups.payment_provider`, event, event)
	rows := make([]agentPayStat, 0)
	err := DB.Raw(sql, userID, userID, userID, userID, common.TopUpStatusSuccess, start, end).Scan(&rows).Error
	return rows, err
}

func agentCommissionInRange(tx *gorm.DB, userID int, start int64, end int64) (int, error) {
	level1, _, _, err := sumAgentLevelCommission(tx, userID, 1, &start, &end)
	if err != nil {
		return 0, err
	}
	level2, _, _, err := sumAgentLevelCommission(tx, userID, 2, &start, &end)
	if err != nil {
		return 0, err
	}
	return addAgentQuota(level1, level2)
}

func orderCommission(level int, status string, amount int64, stored sql.NullInt64) (int, error) {
	if status != common.TopUpStatusSuccess || amount <= 0 {
		return 0, nil
	}
	percent := agentLevel2CommissionPercent
	if level == 1 {
		percent = agentLevel1CommissionPercent
	}
	if stored.Valid && stored.Int64 >= 0 && stored.Int64 <= 100 {
		percent = int(stored.Int64)
	}
	return agentCommissionQuota(amount, percent)
}

func agentEventUnixSQL() string {
	return fmt.Sprintf(
		"CASE WHEN top_ups.status = '%s' AND COALESCE(top_ups.complete_time, 0) > 0 THEN top_ups.complete_time ELSE COALESCE(top_ups.create_time, 0) END",
		common.TopUpStatusSuccess,
	)
}

func agentDayKeySQL() string {
	shifted := fmt.Sprintf("(%s + %d)", agentEventUnixSQL(), agentShanghaiOffset)
	switch {
	case common.UsingMainDatabase(common.DatabaseTypePostgreSQL):
		return fmt.Sprintf("((%s) / 86400)", shifted)
	case common.UsingMainDatabase(common.DatabaseTypeMySQL):
		return fmt.Sprintf("((%s) DIV 86400)", shifted)
	default:
		return fmt.Sprintf("((%s) / 86400)", shifted)
	}
}

func agentRange(now time.Time, days int) (int64, int64) {
	local := now.In(agentShanghaiLocation)
	today := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, agentShanghaiLocation)
	start := today.AddDate(0, 0, -(days - 1)).Unix()
	end := today.AddDate(0, 0, 1).Unix()
	return start, end
}

func agentDayKeyFromUnix(ts int64) int64 {
	return (ts + agentShanghaiOffset) / 86400
}

func agentDayLabel(dayKey int64) (string, string) {
	wall := time.Unix(dayKey*86400-agentShanghaiOffset, 0).In(agentShanghaiLocation)
	return wall.Format("2006-01-02"), wall.Format("01-02")
}

func agentRate(success int64, total int64) float64 {
	if total <= 0 {
		return 0
	}
	return float64(success) / float64(total)
}

func agentPaymentLabel(method string, provider string) string {
	method = strings.TrimSpace(method)
	if method != "" {
		return method
	}
	provider = strings.TrimSpace(provider)
	if provider != "" {
		return provider
	}
	return "unknown"
}

func normalizeAgentDays(days int) (int, error) {
	switch days {
	case 7, 15, 30:
		return days, nil
	default:
		return 0, common.NewMessage("Invalid parameters")
	}
}

func normalizeAgentPage(page int, pageSize int) (int, int, error) {
	if page <= 0 || pageSize <= 0 || pageSize > agentMaxPageSize || page > 1000000 {
		return 0, 0, common.NewMessage("Invalid parameters")
	}
	return page, pageSize, nil
}

func normalizeAgentSearch(keyword string) (string, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return "", nil
	}
	if utf8.RuneCountInString(keyword) > agentSearchLimit {
		return "", common.NewMessage("Search keyword is too long")
	}
	return keyword, nil
}

func agentLike(keyword string) string {
	escaped := strings.NewReplacer(`!`, `!!`, `%`, `!%`, `_`, `!_`).Replace(keyword)
	return "%" + escaped + "%"
}
