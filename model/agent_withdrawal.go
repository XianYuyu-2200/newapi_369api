package model

import (
	"errors"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

const (
	AgentWithdrawalPending  = "pending"
	AgentWithdrawalPaid     = "paid"
	AgentWithdrawalRejected = "rejected"
)

type AgentCommissionWithdrawal struct {
	Id            int    `json:"id" gorm:"primaryKey"`
	UserId        int    `json:"userId" gorm:"index;column:user_id"`
	Quota         int    `json:"quota"`
	AlipayAccount string `json:"alipayAccount" gorm:"type:varchar(128);column:alipay_account"`
	Status        string `json:"status" gorm:"type:varchar(16);index"`
	AdminId       int    `json:"adminId" gorm:"column:admin_id"`
	AdminNote     string `json:"adminNote" gorm:"type:varchar(200);column:admin_note"`
	CreatedAt     int64  `json:"createdAt" gorm:"autoCreateTime;column:created_at"`
	ProcessedAt   int64  `json:"processedAt" gorm:"column:processed_at"`
	Username      string `json:"username,omitempty" gorm:"-"`
	AdminUsername string `json:"adminUsername,omitempty" gorm:"-"`
}

type AgentWithdrawalResult struct {
	Withdrawal     AgentCommissionWithdrawal `json:"withdrawal"`
	AvailableQuota int                       `json:"availableQuota"`
}

func UpdateAgentAlipayAccount(userID int, account string) (string, error) {
	account = strings.TrimSpace(account)
	if userID <= 0 {
		return "", errors.New("user id is empty")
	}
	if len([]rune(account)) > 128 {
		return "", common.NewMessage("Alipay account is too long")
	}
	if err := DB.Model(&User{}).Where("id = ?", userID).Update("agent_alipay_account", account).Error; err != nil {
		return "", err
	}
	return account, nil
}

func CreateAgentWithdrawal(userID int, quota int) (AgentWithdrawalResult, error) {
	if userID <= 0 {
		return AgentWithdrawalResult{}, errors.New("user id is empty")
	}
	if quota <= 0 || common.ValidateWalletQuota(quota) != nil {
		return AgentWithdrawalResult{}, common.NewMessage("Invalid withdrawal amount")
	}
	if quota < agentMinimumTransferQuota() {
		return AgentWithdrawalResult{}, common.NewMessage("Withdrawal amount is below the minimum")
	}
	var result AgentWithdrawalResult
	err := DB.Transaction(func(tx *gorm.DB) error {
		var user User
		if err := lockForUpdate(tx).Select("id", "agent_alipay_account", "agent_commission_transferred", "agent_commission_deducted").First(&user, userID).Error; err != nil {
			return err
		}
		if strings.TrimSpace(user.AgentAlipayAccount) == "" {
			return common.NewMessage("Bind an Alipay account before requesting a withdrawal")
		}
		account, err := loadAgentCommissionAccount(tx, userID, user.AgentCommissionTransferred, user.AgentCommissionDeducted)
		if err != nil {
			return err
		}
		if quota > account.AvailableQuota {
			return common.NewMessage("Available commission is insufficient")
		}
		row := AgentCommissionWithdrawal{UserId: userID, Quota: quota, AlipayAccount: user.AgentAlipayAccount, Status: AgentWithdrawalPending}
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
		result.Withdrawal = row
		result.AvailableQuota = account.AvailableQuota - quota
		return nil
	})
	if err != nil {
		return AgentWithdrawalResult{}, err
	}
	return result, nil
}

func ListAgentWithdrawals(userID int, limit int) ([]AgentCommissionWithdrawal, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	var rows []AgentCommissionWithdrawal
	query := DB.Order("id DESC").Limit(limit)
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(rows)*2)
	set := map[int]bool{}
	for _, row := range rows {
		for _, id := range []int{row.UserId, row.AdminId} {
			if id > 0 && !set[id] {
				set[id] = true
				ids = append(ids, id)
			}
		}
	}
	if len(ids) > 0 {
		var users []User
		if err := DB.Select("id", "username").Where("id IN ?", ids).Find(&users).Error; err != nil {
			return nil, err
		}
		names := map[int]string{}
		for _, u := range users {
			names[u.Id] = u.Username
		}
		for i := range rows {
			rows[i].Username = names[rows[i].UserId]
			rows[i].AdminUsername = names[rows[i].AdminId]
		}
	}
	return rows, nil
}

func ProcessAgentWithdrawal(adminID, adminRole, withdrawalID int, paid bool, note string) (AgentCommissionWithdrawal, error) {
	if adminID <= 0 || withdrawalID <= 0 {
		return AgentCommissionWithdrawal{}, common.NewMessage("Invalid parameters")
	}
	note = strings.TrimSpace(note)
	if len([]rune(note)) > 200 {
		return AgentCommissionWithdrawal{}, common.NewMessage("Note is too long")
	}
	var result AgentCommissionWithdrawal
	err := DB.Transaction(func(tx *gorm.DB) error {
		var row AgentCommissionWithdrawal
		if err := lockForUpdate(tx).First(&row, withdrawalID).Error; err != nil {
			return err
		}
		if row.Status != AgentWithdrawalPending {
			return common.NewMessage("Withdrawal request has already been processed")
		}
		var user User
		if err := lockForUpdate(tx).Select("id", "role", "agent_commission_transferred", "agent_commission_deducted").First(&user, row.UserId).Error; err != nil {
			return err
		}
		if adminRole != common.RoleRootUser && user.Role >= adminRole {
			return common.NewMessage("You cannot process withdrawals for this user")
		}
		status := AgentWithdrawalRejected
		if paid {
			status = AgentWithdrawalPaid
			account, err := loadAgentCommissionAccount(tx, row.UserId, user.AgentCommissionTransferred, user.AgentCommissionDeducted)
			if err != nil {
				return err
			}
			remainingBeforeReservations := commissionRemaining(account.TotalQuota, user.AgentCommissionTransferred, user.AgentCommissionDeducted)
			otherPending := account.PendingWithdrawalQuota - row.Quota
			if otherPending < 0 {
				otherPending = 0
			}
			if row.Quota > remainingBeforeReservations-otherPending {
				return common.NewMessage("Available commission is insufficient")
			}
			updated := tx.Model(&User{}).Where("id = ?", row.UserId).Update("agent_commission_deducted", gorm.Expr("COALESCE(agent_commission_deducted,0)+?", row.Quota))
			if updated.Error != nil {
				return updated.Error
			}
			if updated.RowsAffected != 1 {
				return common.NewMessage("User not found")
			}
			if err := tx.Create(&AgentCommissionDeduction{UserId: row.UserId, AdminId: adminID, Quota: row.Quota, Note: note}).Error; err != nil {
				return err
			}
		}
		processedAt := common.GetTimestamp()
		updated := tx.Model(&AgentCommissionWithdrawal{}).Where("id = ? AND status = ?", row.Id, AgentWithdrawalPending).Updates(map[string]any{"status": status, "admin_id": adminID, "admin_note": note, "processed_at": processedAt})
		if updated.Error != nil {
			return updated.Error
		}
		if updated.RowsAffected != 1 {
			return common.NewMessage("Withdrawal request has already been processed")
		}
		row.Status = status
		row.AdminId = adminID
		row.AdminNote = note
		row.ProcessedAt = processedAt
		result = row
		return nil
	})
	if err != nil {
		return AgentCommissionWithdrawal{}, err
	}
	return result, nil
}
