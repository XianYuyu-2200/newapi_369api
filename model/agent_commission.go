package model

import (
	"errors"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/QuantumNous/new-api/common"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	agentLevel1CommissionPercent = 15
	agentLevel2CommissionPercent = 5
)

// AgentCommissionTransfer is an audit row for commission moved into quota.
// The available balance is still derived from successful top-ups minus
// users.agent_commission_transferred, so this table is not the source of truth.
type AgentCommissionTransfer struct {
	Id        int   `json:"id"`
	UserId    int   `json:"userId" gorm:"index;column:user_id"`
	Quota     int   `json:"quota"`
	CreatedAt int64 `json:"createdAt" gorm:"autoCreateTime;column:created_at"`
}

type AgentCommissionAccount struct {
	Enabled               bool  `json:"enabled"`
	TotalQuota            int   `json:"totalQuota"`
	TransferredQuota      int   `json:"transferredQuota"`
	DeductedQuota         int   `json:"deductedQuota"`
	AvailableQuota        int   `json:"availableQuota"`
	MinimumTransferQuota  int   `json:"minimumTransferQuota"`
	Level1RechargeQuota   int   `json:"level1RechargeQuota"`
	Level2RechargeQuota   int   `json:"level2RechargeQuota"`
	Level1CommissionQuota int   `json:"level1CommissionQuota"`
	Level2CommissionQuota int   `json:"level2CommissionQuota"`
	Level1OrderCount      int64 `json:"level1OrderCount"`
	Level2OrderCount      int64 `json:"level2OrderCount"`
}

type AgentCommissionTransferResult struct {
	Quota            int `json:"quota"`
	AvailableQuota   int `json:"availableQuota"`
	TransferredQuota int `json:"transferredQuota"`
	BalanceQuota     int `json:"balanceQuota"`
}

type inviteeRechargeStat struct {
	Count  int64
	Amount int64
}

func agentMinimumTransferQuota() int {
	quota, err := common.WalletQuotaFromDecimalStrict(decimal.NewFromFloat(common.QuotaPerUnit).Floor())
	if err != nil || quota <= 0 {
		return 1
	}
	return quota
}

func agentRechargeQuota(amount int64) (int, error) {
	if amount <= 0 {
		return 0, nil
	}
	value := decimal.NewFromInt(amount).Mul(decimal.NewFromFloat(common.QuotaPerUnit)).Floor()
	if value.Sign() <= 0 {
		return 0, nil
	}
	return common.WalletQuotaFromDecimalStrict(value)
}

func agentCommissionQuota(amount int64, percent int) (int, error) {
	if amount <= 0 || percent <= 0 {
		return 0, nil
	}
	value := decimal.NewFromInt(amount).
		Mul(decimal.NewFromFloat(common.QuotaPerUnit)).
		Mul(decimal.NewFromInt(int64(percent))).
		Div(decimal.NewFromInt(100)).
		Floor()
	if value.Sign() <= 0 {
		return 0, nil
	}
	return common.WalletQuotaFromDecimalStrict(value)
}

func addAgentQuota(current int, extra int) (int, error) {
	if extra == 0 {
		return current, nil
	}
	if extra < 0 || current > common.MaxWalletQuota-extra {
		return 0, common.NewMessage("Commission exceeds the amount that can be transferred")
	}
	return current + extra, nil
}

func commissionRemaining(total int, transferred int, deducted int) int {
	if transferred < 0 {
		transferred = 0
	}
	if deducted < 0 {
		deducted = 0
	}
	settled, err := addAgentQuota(transferred, deducted)
	if err != nil || settled >= total {
		return 0
	}
	return total - settled
}

func loadAgentCommissionAccount(tx *gorm.DB, userID int, transferred int, deducted int) (AgentCommissionAccount, error) {
	if transferred < 0 {
		transferred = 0
	}
	if deducted < 0 {
		deducted = 0
	}
	account := AgentCommissionAccount{
		Enabled:              true,
		TransferredQuota:     transferred,
		DeductedQuota:        deducted,
		MinimumTransferQuota: agentMinimumTransferQuota(),
	}
	level1Commission, level1Amount, level1Count, err := sumAgentLevelCommission(tx, userID, 1, nil, nil)
	if err != nil {
		return account, err
	}
	level2Commission, level2Amount, level2Count, err := sumAgentLevelCommission(tx, userID, 2, nil, nil)
	if err != nil {
		return account, err
	}
	account.Level1OrderCount = level1Count
	account.Level2OrderCount = level2Count
	account.Level1RechargeQuota, err = agentRechargeQuota(level1Amount)
	if err != nil {
		return account, err
	}
	account.Level2RechargeQuota, err = agentRechargeQuota(level2Amount)
	if err != nil {
		return account, err
	}
	account.Level1CommissionQuota = level1Commission
	account.Level2CommissionQuota = level2Commission
	total, err := addAgentQuota(account.Level1CommissionQuota, account.Level2CommissionQuota)
	if err != nil {
		return account, err
	}
	account.TotalQuota = total
	account.AvailableQuota = commissionRemaining(total, transferred, deducted)
	return account, nil
}

func sumAgentLevelCommission(tx *gorm.DB, userID int, level int, start *int64, end *int64) (int, int64, int64, error) {
	if tx == nil {
		tx = DB
	}
	column := "top_ups.level2_commission_percent"
	fallback := agentLevel2CommissionPercent
	relation := "users.second_inviter_id = ? AND users.inviter_id <> ?"
	args := []any{common.TopUpStatusSuccess, userID, userID, userID}
	if level == 1 {
		column = "top_ups.level1_commission_percent"
		fallback = agentLevel1CommissionPercent
		relation = "users.inviter_id = ?"
		args = []any{common.TopUpStatusSuccess, userID, userID}
	}
	timeClause := ""
	if start != nil && end != nil {
		event := agentEventUnixSQL()
		timeClause = " AND (" + event + ") >= ? AND (" + event + ") < ?"
		args = append(args, *start, *end)
	}
	percentSQL := agentStoredPercentSQL(column, fallback)
	sql := `
SELECT ` + percentSQL + ` AS percent,
       COALESCE(SUM(top_ups.amount), 0) AS amount,
       COUNT(top_ups.id) AS count
FROM top_ups
JOIN users ON users.id = top_ups.user_id AND users.deleted_at IS NULL
WHERE top_ups.status = ? AND users.id <> ? AND ` + relation + timeClause + `
GROUP BY ` + percentSQL
	var rows []struct {
		Percent int
		Amount  int64
		Count   int64
	}
	if err := tx.Raw(sql, args...).Scan(&rows).Error; err != nil {
		return 0, 0, 0, err
	}
	var amount int64
	var count int64
	commission := 0
	for _, row := range rows {
		amount += row.Amount
		count += row.Count
		part, err := agentCommissionQuota(row.Amount, row.Percent)
		if err != nil {
			return 0, 0, 0, err
		}
		commission, err = addAgentQuota(commission, part)
		if err != nil {
			return 0, 0, 0, err
		}
	}
	return commission, amount, count, nil
}

func agentStoredPercentSQL(column string, fallback int) string {
	return "CASE WHEN " + column + " IS NULL OR " + column + " < 0 THEN " + strconv.Itoa(fallback) + " ELSE " + column + " END"
}

func commissionQuotaByUser(userIDs []int, column string, fallback int) (map[int]int, error) {
	result := make(map[int]int, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}
	percentSQL := agentStoredPercentSQL(column, fallback)
	sql := `
SELECT top_ups.user_id AS user_id,
       ` + percentSQL + ` AS percent,
       COALESCE(SUM(top_ups.amount), 0) AS amount
FROM top_ups
WHERE top_ups.status = ? AND top_ups.user_id IN ?
GROUP BY top_ups.user_id, ` + percentSQL
	var rows []struct {
		UserId  int
		Percent int
		Amount  int64
	}
	if err := DB.Raw(sql, common.TopUpStatusSuccess, userIDs).Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		part, err := agentCommissionQuota(row.Amount, row.Percent)
		if err != nil {
			return nil, err
		}
		total, err := addAgentQuota(result[row.UserId], part)
		if err != nil {
			return nil, err
		}
		result[row.UserId] = total
	}
	return result, nil
}

func inviteeRechargeStats(userIDs []int) (map[int]inviteeRechargeStat, error) {
	stats := make(map[int]inviteeRechargeStat, len(userIDs))
	if len(userIDs) == 0 {
		return stats, nil
	}
	var rows []struct {
		UserId int
		Amount int64
		Count  int64
	}
	err := DB.Model(&TopUp{}).
		Select("user_id, COALESCE(SUM(amount), 0) AS amount, COUNT(id) AS count").
		Where("status = ? AND user_id IN ?", common.TopUpStatusSuccess, userIDs).
		Group("user_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		stats[row.UserId] = inviteeRechargeStat{Count: row.Count, Amount: row.Amount}
	}
	return stats, nil
}

// TransferAgentCommission moves quota from calculated commission into the
// user's wallet balance. The conditional update is the lock: a concurrent
// transfer cannot spend the same available commission.
func TransferAgentCommission(userID int, quota int) (AgentCommissionTransferResult, error) {
	if userID <= 0 {
		return AgentCommissionTransferResult{}, errors.New("user id is empty")
	}
	if quota <= 0 {
		return AgentCommissionTransferResult{}, common.NewMessage("Transfer amount must be greater than 0")
	}
	if err := common.ValidateWalletQuota(quota); err != nil {
		return AgentCommissionTransferResult{}, common.NewMessage("Transfer amount is too large")
	}
	minimum := agentMinimumTransferQuota()
	if quota < minimum {
		return AgentCommissionTransferResult{}, common.NewMessage("Transfer amount is below the minimum")
	}

	var result AgentCommissionTransferResult
	err := DB.Transaction(func(tx *gorm.DB) error {
		var user User
		if err := lockForUpdate(tx).Select("id", "quota", "agent_commission_transferred", "agent_commission_deducted").First(&user, userID).Error; err != nil {
			return err
		}
		account, err := loadAgentCommissionAccount(tx, userID, user.AgentCommissionTransferred, user.AgentCommissionDeducted)
		if err != nil {
			return err
		}
		if quota > account.AvailableQuota {
			return common.NewMessage("Available commission is insufficient")
		}
		maxCurrent, err := topUpQuotaMaxCurrent(quota)
		if err != nil {
			return common.NewMessage("Transfer amount is too large")
		}
		updated := tx.Model(&User{}).
			Where(
				"id = ? AND COALESCE(agent_commission_transferred, 0) + COALESCE(agent_commission_deducted, 0) + ? <= ? AND quota <= ?",
				userID,
				quota,
				account.TotalQuota,
				maxCurrent,
			).
			Updates(map[string]any{
				"quota":                        gorm.Expr("quota + ?", quota),
				"agent_commission_transferred": gorm.Expr("COALESCE(agent_commission_transferred, 0) + ?", quota),
			})
		if updated.Error != nil {
			return updated.Error
		}
		if updated.RowsAffected != 1 {
			var current User
			if err := tx.Select("id", "quota").First(&current, userID).Error; err != nil {
				return err
			}
			if current.Quota > maxCurrent {
				return ErrWalletQuotaLimitExceeded
			}
			return common.NewMessage("Available commission is insufficient")
		}
		if err := tx.Create(&AgentCommissionTransfer{UserId: userID, Quota: quota}).Error; err != nil {
			return err
		}
		var current User
		if err := tx.Select("quota", "agent_commission_transferred", "agent_commission_deducted").First(&current, userID).Error; err != nil {
			return err
		}
		result.Quota = quota
		result.BalanceQuota = current.Quota
		result.TransferredQuota = current.AgentCommissionTransferred
		result.AvailableQuota = commissionRemaining(account.TotalQuota, current.AgentCommissionTransferred, current.AgentCommissionDeducted)
		return nil
	})
	if err != nil {
		return AgentCommissionTransferResult{}, err
	}
	syncCreditUserQuotaCache(userID, quota, "agent commission transfer")
	return result, nil
}

const agentCommissionNoteLimit = 200

// AgentCommissionDeduction records commission an admin settled outside the
// wallet, such as a WeChat or Alipay payout. It reduces available commission
// and does not increase users.quota.
type AgentCommissionDeduction struct {
	Id        int    `json:"id"`
	UserId    int    `json:"userId" gorm:"index;column:user_id"`
	AdminId   int    `json:"adminId" gorm:"index;column:admin_id"`
	Quota     int    `json:"quota"`
	Note      string `json:"note" gorm:"type:varchar(200)"`
	CreatedAt int64  `json:"createdAt" gorm:"autoCreateTime;column:created_at"`
}

type AgentCommissionDeductResult struct {
	Quota            int `json:"quota"`
	AvailableQuota   int `json:"availableQuota"`
	TransferredQuota int `json:"transferredQuota"`
	DeductedQuota    int `json:"deductedQuota"`
	TotalQuota       int `json:"totalQuota"`
	BalanceQuota     int `json:"balanceQuota"`
}

type AgentCommissionAdminAccount struct {
	Id               int    `json:"id"`
	Username         string `json:"username"`
	DisplayName      string `json:"displayName"`
	AgentStatus      string `json:"agentStatus"`
	TotalQuota       int    `json:"totalQuota"`
	TransferredQuota int    `json:"transferredQuota"`
	DeductedQuota    int    `json:"deductedQuota"`
	AvailableQuota   int    `json:"availableQuota"`
	Level1Rate       int    `json:"level1Rate"`
	Level2Rate       int    `json:"level2Rate"`
	Level1Custom     bool   `json:"level1Custom"`
	Level2Custom     bool   `json:"level2Custom"`
}

type AgentCommissionDeductionView struct {
	Id            int    `json:"id"`
	UserId        int    `json:"userId"`
	Username      string `json:"username"`
	AdminId       int    `json:"adminId"`
	AdminUsername string `json:"adminUsername"`
	Quota         int    `json:"quota"`
	Note          string `json:"note"`
	CreatedAt     int64  `json:"createdAt"`
}

func normalizeDeductionNote(note string) (string, error) {
	note = strings.TrimSpace(note)
	if note == "" {
		return "", common.NewMessage("Enter a deduction note")
	}
	if utf8.RuneCountInString(note) > agentCommissionNoteLimit {
		return "", common.NewMessage("Commission deduction note is too long")
	}
	return note, nil
}

// DeductAgentCommission settles commission without adding it to the user's
// balance. Available commission is total commission minus transfers minus
// deductions, so the same amount cannot be transferred afterwards.
func DeductAgentCommission(adminID int, userID int, quota int, note string) (AgentCommissionDeductResult, error) {
	if adminID <= 0 || userID <= 0 {
		return AgentCommissionDeductResult{}, errors.New("user id is empty")
	}
	if quota <= 0 {
		return AgentCommissionDeductResult{}, common.NewMessage("Deduct amount must be greater than 0")
	}
	if err := common.ValidateWalletQuota(quota); err != nil {
		return AgentCommissionDeductResult{}, common.NewMessage("Deduct amount is too large")
	}
	note, err := normalizeDeductionNote(note)
	if err != nil {
		return AgentCommissionDeductResult{}, err
	}

	var result AgentCommissionDeductResult
	err = DB.Transaction(func(tx *gorm.DB) error {
		var user User
		if err := lockForUpdate(tx).Select("id", "quota", "agent_commission_transferred", "agent_commission_deducted").First(&user, userID).Error; err != nil {
			return err
		}
		account, err := loadAgentCommissionAccount(tx, userID, user.AgentCommissionTransferred, user.AgentCommissionDeducted)
		if err != nil {
			return err
		}
		if quota > account.AvailableQuota {
			return common.NewMessage("Available commission is insufficient")
		}
		updated := tx.Model(&User{}).
			Where(
				"id = ? AND COALESCE(agent_commission_transferred, 0) + COALESCE(agent_commission_deducted, 0) + ? <= ?",
				userID,
				quota,
				account.TotalQuota,
			).
			Updates(map[string]any{
				"agent_commission_deducted": gorm.Expr("COALESCE(agent_commission_deducted, 0) + ?", quota),
			})
		if updated.Error != nil {
			return updated.Error
		}
		if updated.RowsAffected != 1 {
			return common.NewMessage("Available commission is insufficient")
		}
		if err := tx.Create(&AgentCommissionDeduction{
			UserId:  userID,
			AdminId: adminID,
			Quota:   quota,
			Note:    note,
		}).Error; err != nil {
			return err
		}
		var current User
		if err := tx.Select("quota", "agent_commission_transferred", "agent_commission_deducted").First(&current, userID).Error; err != nil {
			return err
		}
		result.Quota = quota
		result.BalanceQuota = current.Quota
		result.TransferredQuota = current.AgentCommissionTransferred
		result.DeductedQuota = current.AgentCommissionDeducted
		result.TotalQuota = account.TotalQuota
		result.AvailableQuota = commissionRemaining(account.TotalQuota, current.AgentCommissionTransferred, current.AgentCommissionDeducted)
		return nil
	})
	if err != nil {
		return AgentCommissionDeductResult{}, err
	}
	return result, nil
}

func SearchAgentCommissionAccounts(keyword string, actorRole int) ([]AgentCommissionAdminAccount, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" || len(keyword) > 64 || strings.ContainsAny(keyword, "%\\") {
		return nil, common.NewMessage("Enter a username or user ID")
	}
	like := agentLike(keyword)
	id := 0
	if parsed, err := strconv.Atoi(keyword); err == nil && parsed > 0 && strconv.Itoa(parsed) == keyword {
		id = parsed
	}
	query := DB.Model(&User{}).
		Select("id", "username", "display_name", "role", "agent_commission_transferred", "agent_commission_deducted", "agent_status", "agent_level1_rate", "agent_level2_rate").
		Where("(username LIKE ? ESCAPE '!' OR display_name LIKE ? ESCAPE '!' OR id = ?)", like, like, id)
	if actorRole != common.RoleRootUser {
		query = query.Where("role < ?", actorRole)
	}
	var users []User
	if err := query.Order("id DESC").Limit(20).Find(&users).Error; err != nil {
		return nil, err
	}
	accounts := make([]AgentCommissionAdminAccount, 0, len(users))
	for _, user := range users {
		account, err := loadAgentCommissionAccount(DB, user.Id, user.AgentCommissionTransferred, user.AgentCommissionDeducted)
		if err != nil {
			return nil, err
		}
		level1Rate, level1Custom := describeAgentRate(user.AgentLevel1Rate, agentLevel1CommissionPercent)
		level2Rate, level2Custom := describeAgentRate(user.AgentLevel2Rate, agentLevel2CommissionPercent)
		accounts = append(accounts, AgentCommissionAdminAccount{
			Id:               user.Id,
			Username:         user.Username,
			DisplayName:      user.DisplayName,
			AgentStatus:      agentStatusName(user.AgentStatus),
			TotalQuota:       account.TotalQuota,
			TransferredQuota: account.TransferredQuota,
			DeductedQuota:    account.DeductedQuota,
			AvailableQuota:   account.AvailableQuota,
			Level1Rate:       level1Rate,
			Level2Rate:       level2Rate,
			Level1Custom:     level1Custom,
			Level2Custom:     level2Custom,
		})
	}
	return accounts, nil
}

func ListAgentCommissionDeductions(limit int) ([]AgentCommissionDeductionView, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	var rows []AgentCommissionDeduction
	if err := DB.Order("id DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []AgentCommissionDeductionView{}, nil
	}
	ids := make([]int, 0, len(rows)*2)
	seen := map[int]struct{}{}
	for _, row := range rows {
		for _, id := range []int{row.UserId, row.AdminId} {
			if _, ok := seen[id]; ok || id <= 0 {
				continue
			}
			seen[id] = struct{}{}
			ids = append(ids, id)
		}
	}
	names := map[int]string{}
	if len(ids) > 0 {
		var users []User
		if err := DB.Select("id", "username").Where("id IN ?", ids).Find(&users).Error; err != nil {
			return nil, err
		}
		for _, user := range users {
			names[user.Id] = user.Username
		}
	}
	views := make([]AgentCommissionDeductionView, 0, len(rows))
	for _, row := range rows {
		views = append(views, AgentCommissionDeductionView{
			Id:            row.Id,
			UserId:        row.UserId,
			Username:      names[row.UserId],
			AdminId:       row.AdminId,
			AdminUsername: names[row.AdminId],
			Quota:         row.Quota,
			Note:          row.Note,
			CreatedAt:     row.CreatedAt,
		})
	}
	return views, nil
}
