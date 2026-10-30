package model

import (
	"errors"

	"github.com/QuantumNous/new-api/common"

	"gorm.io/gorm"
)

const (
	AgentStatusNone     = 0
	AgentStatusPending  = 1 // Legacy status, normalized during startup.
	AgentStatusApproved = 2
	AgentStatusRejected = 3 // Legacy status, normalized during startup.
)

// AgentRateUpdate is the effective rate after an admin saves a per-agent rate.
type AgentRateUpdate struct {
	UserId       int  `json:"userId"`
	Level1Rate   int  `json:"level1Rate"`
	Level2Rate   int  `json:"level2Rate"`
	Level1Custom bool `json:"level1Custom"`
	Level2Custom bool `json:"level2Custom"`
}

func agentStatusName(status int) string {
	return "approved"
}

func describeAgentRate(custom *int, fallback int) (int, bool) {
	if custom == nil || *custom < 0 || *custom > 100 {
		return fallback, false
	}
	return *custom, true
}

// SetAgentCommissionRates stores a per-agent rate. -1 restores that level's default.
// Rates apply only to recharges completed after this change.
func SetAgentCommissionRates(userID int, level1Rate int, level2Rate int) (AgentRateUpdate, error) {
	if userID <= 0 {
		return AgentRateUpdate{}, errors.New("user id is empty")
	}
	level1, err := normalizeStoredAgentRate(level1Rate)
	if err != nil {
		return AgentRateUpdate{}, err
	}
	level2, err := normalizeStoredAgentRate(level2Rate)
	if err != nil {
		return AgentRateUpdate{}, err
	}
	result := DB.Exec(
		"UPDATE users SET agent_level1_rate = ?, agent_level2_rate = ? WHERE id = ? AND deleted_at IS NULL",
		level1,
		level2,
		userID,
	)
	if result.Error != nil {
		return AgentRateUpdate{}, result.Error
	}
	if result.RowsAffected != 1 {
		var count int64
		if err := DB.Model(&User{}).Where("id = ?", userID).Count(&count).Error; err != nil {
			return AgentRateUpdate{}, err
		}
		if count == 0 {
			return AgentRateUpdate{}, gorm.ErrRecordNotFound
		}
		return AgentRateUpdate{}, common.NewMessage("The user is unavailable for commission rate updates")
	}
	var user User
	if err := DB.Select("agent_level1_rate", "agent_level2_rate").First(&user, userID).Error; err != nil {
		return AgentRateUpdate{}, err
	}
	rate1, custom1 := describeAgentRate(user.AgentLevel1Rate, agentLevel1CommissionPercent)
	rate2, custom2 := describeAgentRate(user.AgentLevel2Rate, agentLevel2CommissionPercent)
	return AgentRateUpdate{
		UserId:       userID,
		Level1Rate:   rate1,
		Level2Rate:   rate2,
		Level1Custom: custom1,
		Level2Custom: custom2,
	}, nil
}

func normalizeStoredAgentRate(rate int) (any, error) {
	if rate == -1 {
		return nil, nil
	}
	if rate < 0 || rate > 100 {
		return nil, common.NewMessage("Commission rate must be between 0 and 100")
	}
	return rate, nil
}

// BackfillApprovedAgents enables existing accounts for universal invite-code
// access and ensures legacy accounts without a code receive one.
func BackfillApprovedAgents() error {
	if err := DB.Model(&User{}).
		Where("agent_status <> ? OR agent_apply_note <> ? OR agent_review_note <> ? OR agent_applied_at <> ? OR agent_reviewed_at <> ?", AgentStatusApproved, "", "", 0, 0).
		Updates(map[string]any{
			"agent_status":      AgentStatusApproved,
			"agent_apply_note":  "",
			"agent_review_note": "",
			"agent_applied_at":  0,
			"agent_reviewed_at": 0,
		}).Error; err != nil {
		return err
	}
	var userIDs []int
	if err := DB.Model(&User{}).
		Where("aff_code = ? OR aff_code IS NULL", "").
		Pluck("id", &userIDs).Error; err != nil {
		return err
	}
	for _, userID := range userIDs {
		if _, err := EnsureUserAffCode(userID); err != nil {
			return err
		}
	}
	return nil
}

func snapshotAgentCommissionRates(tx *gorm.DB, topUp *TopUp) error {
	if topUp == nil {
		return errors.New("topup is empty")
	}
	if topUp.Level1CommissionPercent != nil && topUp.Level2CommissionPercent != nil {
		return nil
	}
	level1 := 0
	level2 := 0
	var payer User
	err := tx.Select("id", "inviter_id", "second_inviter_id").Where("id = ?", topUp.UserId).First(&payer).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err == nil {
		if payer.InviterId > 0 && payer.InviterId != payer.Id {
			rate, rateErr := agentCommissionPercentFor(tx, payer.InviterId, 1)
			if rateErr != nil {
				return rateErr
			}
			level1 = rate
		}
		if payer.SecondInviterId > 0 && payer.SecondInviterId != payer.Id && payer.SecondInviterId != payer.InviterId {
			rate, rateErr := agentCommissionPercentFor(tx, payer.SecondInviterId, 2)
			if rateErr != nil {
				return rateErr
			}
			level2 = rate
		}
	}
	topUp.Level1CommissionPercent = &level1
	topUp.Level2CommissionPercent = &level2
	return nil
}

func agentCommissionPercentFor(tx *gorm.DB, userID int, level int) (int, error) {
	var user User
	err := tx.Select("id", "agent_status", "agent_level1_rate", "agent_level2_rate").First(&user, userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	custom := user.AgentLevel2Rate
	fallback := agentLevel2CommissionPercent
	if level == 1 {
		custom = user.AgentLevel1Rate
		fallback = agentLevel1CommissionPercent
	}
	rate, _ := describeAgentRate(custom, fallback)
	return rate, nil
}

// ApprovedAgent is one approved agent in the admin directory.
// Commission amounts come from successful recharge snapshots, not the current rate.
type ApprovedAgent struct {
	Id                    int    `json:"id"`
	Username              string `json:"username"`
	DisplayName           string `json:"displayName"`
	InviteCode            string `json:"inviteCode"`
	Quota                 int    `json:"quota"`
	UsedQuota             int    `json:"usedQuota"`
	Level1Count           int64  `json:"level1Count"`
	Level2Count           int64  `json:"level2Count"`
	TotalQuota            int    `json:"totalQuota"`
	TransferredQuota      int    `json:"transferredQuota"`
	DeductedQuota         int    `json:"deductedQuota"`
	AvailableQuota        int    `json:"availableQuota"`
	Level1CommissionQuota int    `json:"level1CommissionQuota"`
	Level2CommissionQuota int    `json:"level2CommissionQuota"`
	Level1Rate            int    `json:"level1Rate"`
	Level2Rate            int    `json:"level2Rate"`
	Level1Custom          bool   `json:"level1Custom"`
	Level2Custom          bool   `json:"level2Custom"`
	AppliedAt             int64  `json:"appliedAt"`
	ReviewedAt            int64  `json:"reviewedAt"`
}

// ApprovedAgentPage is one page of approved agents.
type ApprovedAgentPage struct {
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
	Total    int64           `json:"total"`
	Items    []ApprovedAgent `json:"items"`
}

// ListApprovedAgents returns every approved agent the actor is allowed to see.
// People approved by the earlier promoter backfill are included.
func ListApprovedAgents(actorRole int, page int, pageSize int) (ApprovedAgentPage, error) {
	result := ApprovedAgentPage{Items: []ApprovedAgent{}}
	page, pageSize, err := normalizeAgentPage(page, pageSize)
	if err != nil {
		return result, err
	}
	if pageSize != 20 {
		return result, common.NewMessage("Invalid parameters")
	}
	query := DB.Model(&User{}).Where("agent_status = ?", AgentStatusApproved)
	if actorRole != common.RoleRootUser {
		query = query.Where("role < ?", actorRole)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return result, err
	}
	rows := make([]User, 0)
	if total > 0 {
		err = query.
			Select(
				"id",
				"username",
				"display_name",
				"quota",
				"used_quota",
				"aff_code",
				"agent_level1_rate",
				"agent_level2_rate",
				"agent_commission_transferred",
				"agent_commission_deducted",
				"agent_applied_at",
				"agent_reviewed_at",
			).
			Order("agent_reviewed_at DESC, id DESC").
			Offset((page - 1) * pageSize).
			Limit(pageSize).
			Find(&rows).Error
		if err != nil {
			return result, err
		}
	}
	ids := make([]int, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.Id)
	}
	level1Count, err := countInviteesByAgent(ids, 1)
	if err != nil {
		return result, err
	}
	level2Count, err := countInviteesByAgent(ids, 2)
	if err != nil {
		return result, err
	}
	level1Commission, err := commissionByAgent(ids, 1)
	if err != nil {
		return result, err
	}
	level2Commission, err := commissionByAgent(ids, 2)
	if err != nil {
		return result, err
	}
	items := make([]ApprovedAgent, 0, len(rows))
	for _, row := range rows {
		level1Rate, level1Custom := describeAgentRate(row.AgentLevel1Rate, agentLevel1CommissionPercent)
		level2Rate, level2Custom := describeAgentRate(row.AgentLevel2Rate, agentLevel2CommissionPercent)
		level1 := level1Commission[row.Id]
		level2 := level2Commission[row.Id]
		totalQuota, err := addAgentQuota(level1, level2)
		if err != nil {
			return result, err
		}
		transferred := row.AgentCommissionTransferred
		deducted := row.AgentCommissionDeducted
		if transferred < 0 {
			transferred = 0
		}
		if deducted < 0 {
			deducted = 0
		}
		items = append(items, ApprovedAgent{
			Id:                    row.Id,
			Username:              row.Username,
			DisplayName:           row.DisplayName,
			InviteCode:            row.AffCode,
			Quota:                 row.Quota,
			UsedQuota:             row.UsedQuota,
			Level1Count:           level1Count[row.Id],
			Level2Count:           level2Count[row.Id],
			TotalQuota:            totalQuota,
			TransferredQuota:      transferred,
			DeductedQuota:         deducted,
			AvailableQuota:        commissionRemaining(totalQuota, transferred, deducted),
			Level1CommissionQuota: level1,
			Level2CommissionQuota: level2,
			Level1Rate:            level1Rate,
			Level2Rate:            level2Rate,
			Level1Custom:          level1Custom,
			Level2Custom:          level2Custom,
			AppliedAt:             row.AgentAppliedAt,
			ReviewedAt:            row.AgentReviewedAt,
		})
	}
	return ApprovedAgentPage{
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		Items:    items,
	}, nil
}

func countInviteesByAgent(ids []int, level int) (map[int]int64, error) {
	counts := make(map[int]int64, len(ids))
	if len(ids) == 0 {
		return counts, nil
	}
	column := "second_inviter_id"
	query := DB.Model(&User{}).Where("id <> second_inviter_id AND second_inviter_id IN ? AND inviter_id <> second_inviter_id", ids)
	if level == 1 {
		column = "inviter_id"
		query = DB.Model(&User{}).Where("id <> inviter_id AND inviter_id IN ?", ids)
	}
	var rows []struct {
		AgentId int
		Count   int64
	}
	err := query.Select(column + " AS agent_id, COUNT(id) AS count").Group(column).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		counts[row.AgentId] = row.Count
	}
	return counts, nil
}

func commissionByAgent(ids []int, level int) (map[int]int, error) {
	result := make(map[int]int, len(ids))
	if len(ids) == 0 {
		return result, nil
	}
	column := "top_ups.level2_commission_percent"
	fallback := agentLevel2CommissionPercent
	agentColumn := "users.second_inviter_id"
	relation := "users.inviter_id <> users.second_inviter_id AND "
	if level == 1 {
		column = "top_ups.level1_commission_percent"
		fallback = agentLevel1CommissionPercent
		agentColumn = "users.inviter_id"
		relation = ""
	}
	percentSQL := agentStoredPercentSQL(column, fallback)
	sql := `
SELECT ` + agentColumn + ` AS agent_id,
       ` + percentSQL + ` AS percent,
       COALESCE(SUM(top_ups.amount), 0) AS amount
FROM top_ups
JOIN users ON users.id = top_ups.user_id AND users.deleted_at IS NULL
WHERE top_ups.status = ?
  AND users.id <> ` + agentColumn + `
  AND ` + relation + agentColumn + ` IN ?
GROUP BY ` + agentColumn + `, ` + percentSQL
	var rows []struct {
		AgentId int
		Percent int
		Amount  int64
	}
	if err := DB.Raw(sql, common.TopUpStatusSuccess, ids).Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		part, err := agentCommissionQuota(row.Amount, row.Percent)
		if err != nil {
			return nil, err
		}
		total, err := addAgentQuota(result[row.AgentId], part)
		if err != nil {
			return nil, err
		}
		result[row.AgentId] = total
	}
	return result, nil
}
