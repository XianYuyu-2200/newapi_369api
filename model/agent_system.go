package model

import (
	"errors"

	"github.com/QuantumNous/new-api/common"

	"gorm.io/gorm"
)

const AgentInviteeListLimit = 200

type AgentInvitee struct {
	Id              int    `json:"id"`
	Username        string `json:"username"`
	DisplayName     string `json:"displayName"`
	Email           string `json:"email"`
	Status          int    `json:"status"`
	Level           int    `json:"level"`
	CreatedAt       int64  `json:"createdAt"`
	RechargeCount   int64  `json:"rechargeCount"`
	RechargeQuota   int    `json:"rechargeQuota"`
	CommissionQuota int    `json:"commissionQuota"`
}

type AgentSystemReport struct {
	CommissionEnabled      bool           `json:"commissionEnabled"`
	AgentStatus            string         `json:"agentStatus"`
	InviteCode             string         `json:"inviteCode"`
	ApplyNote              string         `json:"applyNote"`
	ReviewNote             string         `json:"reviewNote"`
	Level1Rate             int            `json:"level1Rate"`
	Level2Rate             int            `json:"level2Rate"`
	Level1RateCustom       bool           `json:"level1RateCustom"`
	Level2RateCustom       bool           `json:"level2RateCustom"`
	Level1Count            int64          `json:"level1Count"`
	Level2Count            int64          `json:"level2Count"`
	Users                  []AgentInvitee `json:"users"`
	UserLimit              int            `json:"userLimit"`
	UsersTruncated         bool           `json:"usersTruncated"`
	TotalQuota             int            `json:"totalQuota"`
	TransferredQuota       int            `json:"transferredQuota"`
	DeductedQuota          int            `json:"deductedQuota"`
	AvailableQuota         int            `json:"availableQuota"`
	MinimumTransferQuota   int            `json:"minimumTransferQuota"`
	PendingWithdrawalQuota int            `json:"pendingWithdrawalQuota"`
	AlipayAccount          string         `json:"alipayAccount"`
	Level1RechargeQuota    int            `json:"level1RechargeQuota"`
	Level2RechargeQuota    int            `json:"level2RechargeQuota"`
	Level1CommissionQuota  int            `json:"level1CommissionQuota"`
	Level2CommissionQuota  int            `json:"level2CommissionQuota"`
	Level1OrderCount       int64          `json:"level1OrderCount"`
	Level2OrderCount       int64          `json:"level2OrderCount"`
}

type agentInviteeRow struct {
	Id          int
	Username    string
	DisplayName string
	Email       string
	Status      int
	CreatedAt   int64
	InviterId   int
}

// EnsureUserAffCode returns the user's invite code, creating one when missing.
// It updates only aff_code and does not rewrite the rest of the user row.
func EnsureUserAffCode(userID int) (string, error) {
	if userID <= 0 {
		return "", errors.New("user id is empty")
	}
	var current User
	if err := DB.Select("id", "aff_code").First(&current, userID).Error; err != nil {
		return "", err
	}
	if current.AffCode != "" {
		if err := DB.Model(&User{}).Where("id = ? AND agent_status <> ?", userID, AgentStatusApproved).Update("agent_status", AgentStatusApproved).Error; err != nil {
			return "", err
		}
		return current.AffCode, nil
	}
	code := common.GetRandomString(4)
	result := DB.Model(&User{}).
		Where("id = ? AND (aff_code = ? OR aff_code IS NULL)", userID, "").
		Updates(map[string]any{"aff_code": code, "agent_status": AgentStatusApproved})
	if result.Error != nil {
		return "", result.Error
	}
	if err := DB.Select("aff_code").First(&current, userID).Error; err != nil {
		return "", err
	}
	if current.AffCode == "" {
		return "", errors.New("failed to create invite code")
	}
	return current.AffCode, nil
}

func GetAgentSystem(userID int) (AgentSystemReport, error) {
	report := AgentSystemReport{
		CommissionEnabled: true,
		Users:             make([]AgentInvitee, 0),
		UserLimit:         AgentInviteeListLimit,
	}
	if userID <= 0 {
		return report, errors.New("user id is empty")
	}
	var current User
	if err := DB.Select(
		"aff_code",
		"agent_status",
		"agent_level1_rate",
		"agent_level2_rate",
		"agent_apply_note",
		"agent_review_note",
		"agent_commission_transferred",
		"agent_commission_deducted",
		"agent_alipay_account",
	).First(&current, userID).Error; err != nil {
		return report, err
	}
	report.AlipayAccount = current.AgentAlipayAccount
	report.AgentStatus = agentStatusName(current.AgentStatus)
	report.ApplyNote = current.AgentApplyNote
	report.ReviewNote = current.AgentReviewNote
	report.Level1Rate, report.Level1RateCustom = describeAgentRate(current.AgentLevel1Rate, agentLevel1CommissionPercent)
	report.Level2Rate, report.Level2RateCustom = describeAgentRate(current.AgentLevel2Rate, agentLevel2CommissionPercent)
	code, err := EnsureUserAffCode(userID)
	if err != nil {
		return report, err
	}
	report.InviteCode = code
	level1, level2, err := countAgentInvitees(userID)
	if err != nil {
		return report, err
	}
	users, err := listAgentInvitees(userID, AgentInviteeListLimit)
	if err != nil {
		return report, err
	}
	report.Level1Count = level1
	report.Level2Count = level2
	if err := attachInviteeRecharge(users); err != nil {
		return report, err
	}
	report.Users = users
	report.UsersTruncated = level1+level2 > int64(len(users))
	account, err := loadAgentCommissionAccount(DB, userID, current.AgentCommissionTransferred, current.AgentCommissionDeducted)
	if err != nil {
		return report, err
	}
	report.TotalQuota = account.TotalQuota
	report.TransferredQuota = account.TransferredQuota
	report.DeductedQuota = account.DeductedQuota
	report.PendingWithdrawalQuota = account.PendingWithdrawalQuota
	report.AvailableQuota = account.AvailableQuota
	report.MinimumTransferQuota = account.MinimumTransferQuota
	report.Level1RechargeQuota = account.Level1RechargeQuota
	report.Level2RechargeQuota = account.Level2RechargeQuota
	report.Level1CommissionQuota = account.Level1CommissionQuota
	report.Level2CommissionQuota = account.Level2CommissionQuota
	report.Level1OrderCount = account.Level1OrderCount
	report.Level2OrderCount = account.Level2OrderCount
	return report, nil
}

func attachInviteeRecharge(users []AgentInvitee) error {
	if len(users) == 0 {
		return nil
	}
	ids := make([]int, 0, len(users))
	for _, user := range users {
		ids = append(ids, user.Id)
	}
	stats, err := inviteeRechargeStats(ids)
	if err != nil {
		return err
	}
	for i := range users {
		stat := stats[users[i].Id]
		users[i].RechargeCount = stat.Count
		quota, err := agentRechargeQuota(stat.Amount)
		if err != nil {
			return err
		}
		users[i].RechargeQuota = quota
	}
	return attachInviteeCommission(users)
}

func attachInviteeCommission(users []AgentInvitee) error {
	level1IDs := make([]int, 0)
	level2IDs := make([]int, 0)
	for _, user := range users {
		if user.Level == 1 {
			level1IDs = append(level1IDs, user.Id)
			continue
		}
		level2IDs = append(level2IDs, user.Id)
	}
	level1, err := commissionQuotaByUser(level1IDs, "top_ups.level1_commission_percent", agentLevel1CommissionPercent)
	if err != nil {
		return err
	}
	level2, err := commissionQuotaByUser(level2IDs, "top_ups.level2_commission_percent", agentLevel2CommissionPercent)
	if err != nil {
		return err
	}
	for i := range users {
		commission := level2[users[i].Id]
		if users[i].Level == 1 {
			commission = level1[users[i].Id]
		}
		users[i].CommissionQuota = commission
	}
	return nil
}

func countAgentInvitees(userID int) (int64, int64, error) {
	var level1 int64
	var level2 int64
	if err := level1Invitees(userID).Count(&level1).Error; err != nil {
		return 0, 0, err
	}
	if err := level2Invitees(userID).Count(&level2).Error; err != nil {
		return 0, 0, err
	}
	return level1, level2, nil
}

func listAgentInvitees(userID int, limit int) ([]AgentInvitee, error) {
	if limit <= 0 || limit > AgentInviteeListLimit {
		limit = AgentInviteeListLimit
	}
	rows := make([]agentInviteeRow, 0)
	err := agentInviteeBase(userID).
		Select("id", "username", "display_name", "email", "status", "created_at", "inviter_id").
		Order("created_at DESC, id DESC").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, err
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
	return users, nil
}

func agentInviteeBase(userID int) *gorm.DB {
	return DB.Model(&User{}).Where(
		"id <> ? AND (inviter_id = ? OR (second_inviter_id = ? AND inviter_id <> ?))",
		userID,
		userID,
		userID,
		userID,
	)
}

func level1Invitees(userID int) *gorm.DB {
	return DB.Model(&User{}).Where("id <> ? AND inviter_id = ?", userID, userID)
}

func level2Invitees(userID int) *gorm.DB {
	return DB.Model(&User{}).Where(
		"id <> ? AND second_inviter_id = ? AND inviter_id <> ?",
		userID,
		userID,
		userID,
	)
}
