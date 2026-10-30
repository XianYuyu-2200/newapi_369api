package model

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestGetAgentSystemSeparatesTwoLevels(t *testing.T) {
	truncateTables(t)

	root := User{
		Username:    "agent-root",
		Password:    "password-pass",
		AffCode:     "root-code",
		DisplayName: "Root",
		Email:       "root@example.com",
		Role:        common.RoleCommonUser,
		Status:      common.UserStatusEnabled,
		Group:       "default",
		AgentStatus: AgentStatusApproved,
	}
	require.NoError(t, DB.Create(&root).Error)

	direct := User{
		Username:    "agent-direct",
		Password:    "password-pass",
		DisplayName: "Direct",
		Email:       "direct@example.com",
		Role:        common.RoleCommonUser,
		Status:      common.UserStatusEnabled,
		Group:       "default",
	}
	require.NoError(t, direct.Insert(root.Id))

	child := User{
		Username:    "agent-child",
		Password:    "password-pass",
		DisplayName: "Child",
		Email:       "child@example.com",
		Role:        common.RoleCommonUser,
		Status:      common.UserStatusDisabled,
		Group:       "default",
	}
	require.NoError(t, child.Insert(direct.Id))

	grandchild := User{
		Username: "agent-grand",
		Password: "password-pass",
		Email:    "grand@example.com",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
	}
	require.NoError(t, grandchild.Insert(child.Id))

	stranger := User{
		Username: "agent-other",
		Password: "password-pass",
		Email:    "other@example.com",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
	}
	require.NoError(t, stranger.Insert(0))

	report, err := GetAgentSystem(root.Id)
	require.NoError(t, err)
	require.True(t, report.CommissionEnabled)
	require.Equal(t, "root-code", report.InviteCode)
	require.Equal(t, int64(1), report.Level1Count)
	require.Equal(t, int64(1), report.Level2Count)
	require.False(t, report.UsersTruncated)

	levels := map[int]int{}
	for _, user := range report.Users {
		levels[user.Id] = user.Level
		require.NotEmpty(t, user.Username)
	}
	require.Equal(t, 1, levels[direct.Id])
	require.Equal(t, 2, levels[child.Id])
	require.Equal(t, "direct@example.com", userByID(report.Users, direct.Id).Email)
	require.Equal(t, common.UserStatusDisabled, userByID(report.Users, child.Id).Status)
	_, hasGrand := levels[grandchild.Id]
	_, hasStranger := levels[stranger.Id]
	_, hasRoot := levels[root.Id]
	require.False(t, hasGrand)
	require.False(t, hasStranger)
	require.False(t, hasRoot)

	directReport, err := GetAgentSystem(direct.Id)
	require.NoError(t, err)
	require.Equal(t, int64(1), directReport.Level1Count)
	require.Equal(t, int64(1), directReport.Level2Count)
	directLevels := map[int]int{}
	for _, user := range directReport.Users {
		directLevels[user.Id] = user.Level
	}
	require.Equal(t, 1, directLevels[child.Id])
	require.Equal(t, 2, directLevels[grandchild.Id])
	_, hasRoot = directLevels[root.Id]
	require.False(t, hasRoot)

	limited, err := listAgentInvitees(root.Id, 1)
	require.NoError(t, err)
	require.Len(t, limited, 1)
	level1, level2, err := countAgentInvitees(root.Id)
	require.NoError(t, err)
	require.Equal(t, int64(1), level1)
	require.Equal(t, int64(1), level2)
	require.Greater(t, level1+level2, int64(len(limited)))
}

func TestEnsureUserAffCodeCreatesAndKeepsCodeForAnyUser(t *testing.T) {
	truncateTables(t)

	blank := User{
		Username: "agent-blank",
		Password: "password-pass",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
	}
	require.NoError(t, DB.Create(&blank).Error)
	code, err := EnsureUserAffCode(blank.Id)
	require.NoError(t, err)
	require.Len(t, code, 4)
	again, err := EnsureUserAffCode(blank.Id)
	require.NoError(t, err)
	require.Equal(t, code, again)

	kept := User{
		Username:    "agent-kept",
		Password:    "password-pass",
		AffCode:     "keep",
		Role:        common.RoleCommonUser,
		Status:      common.UserStatusEnabled,
		Group:       "default",
		AgentStatus: AgentStatusApproved,
	}
	require.NoError(t, DB.Create(&kept).Error)
	keptCode, err := EnsureUserAffCode(kept.Id)
	require.NoError(t, err)
	require.Equal(t, "keep", keptCode)

	_, err = EnsureUserAffCode(0)
	require.Error(t, err)
}

func userByID(users []AgentInvitee, id int) AgentInvitee {
	for _, user := range users {
		if user.Id == id {
			return user
		}
	}
	return AgentInvitee{}
}

func TestTransferAgentCommissionUsesSuccessfulRechargesOnly(t *testing.T) {
	truncateTables(t)
	previous := common.QuotaPerUnit
	common.QuotaPerUnit = 500000
	t.Cleanup(func() { common.QuotaPerUnit = previous })

	root := User{
		Username: "commission-root",
		Password: "password-pass",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
	}
	require.NoError(t, DB.Create(&root).Error)
	direct := User{
		Username: "commission-direct",
		Password: "password-pass",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
	}
	require.NoError(t, direct.Insert(root.Id))
	child := User{
		Username: "commission-child",
		Password: "password-pass",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
	}
	require.NoError(t, child.Insert(direct.Id))
	stranger := User{
		Username: "commission-stranger",
		Password: "password-pass",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
	}
	require.NoError(t, stranger.Insert(0))

	require.NoError(t, DB.Create(&TopUp{UserId: direct.Id, Amount: 10, Money: 10, TradeNo: "direct-success", Status: common.TopUpStatusSuccess}).Error)
	require.NoError(t, DB.Create(&TopUp{UserId: direct.Id, Amount: 1000, Money: 1000, TradeNo: "direct-pending", Status: common.TopUpStatusPending}).Error)
	require.NoError(t, DB.Create(&TopUp{UserId: child.Id, Amount: 20, Money: 20, TradeNo: "child-success", Status: common.TopUpStatusSuccess}).Error)
	require.NoError(t, DB.Create(&TopUp{UserId: stranger.Id, Amount: 50, Money: 50, TradeNo: "stranger-success", Status: common.TopUpStatusSuccess}).Error)

	// 10 * 500000 * 15% = 750000; 20 * 500000 * 5% = 500000; total 1250000.
	report, err := GetAgentSystem(root.Id)
	require.NoError(t, err)
	require.Equal(t, 750000, report.Level1CommissionQuota)
	require.Equal(t, 500000, report.Level2CommissionQuota)
	require.Equal(t, 1250000, report.AvailableQuota)
	require.Equal(t, int64(1), report.Level1OrderCount)
	require.Equal(t, int64(1), report.Level2OrderCount)
	require.Equal(t, 5000000, report.Level1RechargeQuota)
	require.Equal(t, 10000000, report.Level2RechargeQuota)

	_, err = TransferAgentCommission(root.Id, 1000)
	require.Error(t, err)
	result, err := TransferAgentCommission(root.Id, 500000)
	require.NoError(t, err)
	require.Equal(t, 750000, result.AvailableQuota)
	require.Equal(t, 500000, result.BalanceQuota)

	_, err = TransferAgentCommission(root.Id, 1250000)
	require.Error(t, err)
	result, err = TransferAgentCommission(root.Id, 750000)
	require.NoError(t, err)
	require.Equal(t, 0, result.AvailableQuota)
	require.Equal(t, 1250000, result.BalanceQuota)
	require.Equal(t, 1250000, result.TransferredQuota)

	_, err = TransferAgentCommission(root.Id, 500000)
	require.Error(t, err)
	var saved User
	require.NoError(t, DB.Select("quota", "agent_commission_transferred").First(&saved, root.Id).Error)
	require.Equal(t, 1250000, saved.Quota)
	require.Equal(t, 1250000, saved.AgentCommissionTransferred)
	var transfers int64
	require.NoError(t, DB.Model(&AgentCommissionTransfer{}).Where("user_id = ?", root.Id).Count(&transfers).Error)
	require.Equal(t, int64(2), transfers)
}

func TestDeductAgentCommissionDoesNotIncreaseQuota(t *testing.T) {
	truncateTables(t)
	previous := common.QuotaPerUnit
	common.QuotaPerUnit = 500000
	t.Cleanup(func() { common.QuotaPerUnit = previous })

	root := User{
		Username:    "deduct-root",
		DisplayName: "Deduct Root",
		Password:    "password-pass",
		AffCode:     "deduct-root-code",
		Role:        common.RoleCommonUser,
		Status:      common.UserStatusEnabled,
		Group:       "default",
	}
	require.NoError(t, DB.Create(&root).Error)
	admin := User{
		Username: "deduct-admin",
		Password: "password-pass",
		AffCode:  "deduct-admin-code",
		Role:     common.RoleAdminUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
	}
	require.NoError(t, DB.Create(&admin).Error)
	higher := User{
		Username: "deduct-higher",
		Password: "password-pass",
		AffCode:  "deduct-higher-code",
		Role:     common.RoleRootUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
	}
	require.NoError(t, DB.Create(&higher).Error)
	direct := User{
		Username: "deduct-direct",
		Password: "password-pass",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
	}
	require.NoError(t, direct.Insert(root.Id))
	require.NoError(t, DB.Create(&TopUp{UserId: direct.Id, Amount: 10, Money: 10, TradeNo: "deduct-success", Status: common.TopUpStatusSuccess}).Error)

	_, err := DeductAgentCommission(admin.Id, root.Id, 200000, "   ")
	require.Error(t, err)
	_, err = DeductAgentCommission(admin.Id, root.Id, 0, "微信已转账")
	require.Error(t, err)

	// 10 * 500000 * 15% = 750000. Deducting does not move quota.
	result, err := DeductAgentCommission(admin.Id, root.Id, 200000, "  微信已转账  ")
	require.NoError(t, err)
	require.Equal(t, 550000, result.AvailableQuota)
	require.Equal(t, 200000, result.DeductedQuota)
	require.Equal(t, 0, result.BalanceQuota)

	transfer, err := TransferAgentCommission(root.Id, 500000)
	require.NoError(t, err)
	require.Equal(t, 50000, transfer.AvailableQuota)
	require.Equal(t, 500000, transfer.BalanceQuota)

	_, err = DeductAgentCommission(admin.Id, root.Id, 50001, "超出可核销")
	require.Error(t, err)
	result, err = DeductAgentCommission(admin.Id, root.Id, 50000, "支付宝已转账")
	require.NoError(t, err)
	require.Equal(t, 0, result.AvailableQuota)
	require.Equal(t, 250000, result.DeductedQuota)
	require.Equal(t, 500000, result.TransferredQuota)
	require.Equal(t, 500000, result.BalanceQuota)

	_, err = TransferAgentCommission(root.Id, 500000)
	require.Error(t, err)

	var saved User
	require.NoError(t, DB.Select("quota", "agent_commission_transferred", "agent_commission_deducted").First(&saved, root.Id).Error)
	require.Equal(t, 500000, saved.Quota)
	require.Equal(t, 500000, saved.AgentCommissionTransferred)
	require.Equal(t, 250000, saved.AgentCommissionDeducted)

	report, err := GetAgentSystem(root.Id)
	require.NoError(t, err)
	require.Equal(t, 0, report.AvailableQuota)
	require.Equal(t, 250000, report.DeductedQuota)
	require.Equal(t, 500000, report.TransferredQuota)

	accounts, err := SearchAgentCommissionAccounts("deduct-root", common.RoleAdminUser)
	require.NoError(t, err)
	require.Len(t, accounts, 1)
	require.Equal(t, root.Id, accounts[0].Id)
	require.Equal(t, 0, accounts[0].AvailableQuota)

	hidden, err := SearchAgentCommissionAccounts("deduct-higher", common.RoleAdminUser)
	require.NoError(t, err)
	require.Empty(t, hidden)
	visible, err := SearchAgentCommissionAccounts(strconv.Itoa(higher.Id), common.RoleRootUser)
	require.NoError(t, err)
	require.Len(t, visible, 1)

	rows, err := ListAgentCommissionDeductions(20)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	require.Equal(t, "支付宝已转账", rows[0].Note)
	require.Equal(t, admin.Id, rows[0].AdminId)
	require.Equal(t, "deduct-admin", rows[0].AdminUsername)
	require.Equal(t, "deduct-root", rows[0].Username)
}

func TestAgentSystemStatsAndRecordsFollowSuccessfulRecharges(t *testing.T) {
	truncateTables(t)
	previous := common.QuotaPerUnit
	common.QuotaPerUnit = 500000
	t.Cleanup(func() { common.QuotaPerUnit = previous })

	loc := time.FixedZone("Asia/Shanghai", 8*3600)
	now := time.Date(2026, 10, 24, 15, 4, 0, 0, loc)
	todayLate := time.Date(2026, 10, 24, 11, 0, 0, 0, loc).Unix()
	todayEarly := time.Date(2026, 10, 24, 10, 0, 0, 0, loc).Unix()
	yesterday := time.Date(2026, 10, 23, 23, 0, 0, 0, loc).Unix()
	outsideWeek := time.Date(2026, 10, 17, 12, 0, 0, 0, loc).Unix()
	createdLongAgo := time.Date(2026, 9, 1, 8, 0, 0, 0, loc).Unix()

	root := User{Username: "stats-root", Password: "password-pass", AffCode: "stats-root", Role: common.RoleCommonUser, Status: common.UserStatusEnabled, Group: "default"}
	require.NoError(t, DB.Create(&root).Error)
	direct := User{Username: "stats-direct", Password: "password-pass", AffCode: "stats-direct", DisplayName: "Direct Name", Email: "direct-stats@example.com", Role: common.RoleCommonUser, Status: common.UserStatusEnabled, Group: "default", InviterId: root.Id}
	require.NoError(t, DB.Create(&direct).Error)
	child := User{Username: "stats-child", Password: "password-pass", AffCode: "stats-child", Email: "child-stats@example.com", Role: common.RoleCommonUser, Status: common.UserStatusEnabled, Group: "default", InviterId: direct.Id, SecondInviterId: root.Id}
	require.NoError(t, DB.Create(&child).Error)
	stranger := User{Username: "stats-stranger", Password: "password-pass", AffCode: "stats-stranger", Role: common.RoleCommonUser, Status: common.UserStatusEnabled, Group: "default"}
	require.NoError(t, DB.Create(&stranger).Error)
	deleted := User{Username: "stats-deleted", Password: "password-pass", AffCode: "stats-deleted", Role: common.RoleCommonUser, Status: common.UserStatusEnabled, Group: "default", InviterId: root.Id}
	require.NoError(t, DB.Create(&deleted).Error)
	require.NoError(t, DB.Delete(&deleted).Error)

	require.NoError(t, DB.Create(&TopUp{UserId: direct.Id, Amount: 10, Money: 10, TradeNo: "direct-yesterday", PaymentMethod: "stripe", Status: common.TopUpStatusSuccess, CreateTime: yesterday, CompleteTime: yesterday}).Error)
	require.NoError(t, DB.Create(&TopUp{UserId: child.Id, Amount: 20, Money: 20, TradeNo: "child-today", PaymentProvider: "epay", Status: common.TopUpStatusSuccess, CreateTime: todayEarly, CompleteTime: todayEarly}).Error)
	require.NoError(t, DB.Create(&TopUp{UserId: direct.Id, Amount: 4, Money: 4, TradeNo: "late-success", PaymentMethod: "stripe", Status: common.TopUpStatusSuccess, CreateTime: createdLongAgo, CompleteTime: todayLate}).Error)
	require.NoError(t, DB.Create(&TopUp{UserId: direct.Id, Amount: 8, Money: 8, TradeNo: "old-success", PaymentMethod: "stripe", Status: common.TopUpStatusSuccess, CreateTime: outsideWeek, CompleteTime: outsideWeek}).Error)
	require.NoError(t, DB.Create(&TopUp{UserId: direct.Id, Amount: 1000, Money: 1000, TradeNo: "direct-pending", PaymentMethod: "stripe", Status: common.TopUpStatusPending, CreateTime: todayEarly}).Error)
	require.NoError(t, DB.Create(&TopUp{UserId: direct.Id, Amount: 7, Money: 7, TradeNo: "direct-failed", PaymentMethod: "stripe", Status: common.TopUpStatusFailed, CreateTime: todayEarly}).Error)
	require.NoError(t, DB.Create(&TopUp{UserId: stranger.Id, Amount: 50, Money: 50, TradeNo: "stranger-success", Status: common.TopUpStatusSuccess, CreateTime: todayEarly, CompleteTime: todayEarly}).Error)
	require.NoError(t, DB.Create(&TopUp{UserId: deleted.Id, Amount: 999, Money: 999, TradeNo: "deleted-success", Status: common.TopUpStatusSuccess, CreateTime: todayEarly, CompleteTime: todayEarly}).Error)

	_, err := getAgentSystemStatsAt(root.Id, 6, now)
	require.Error(t, err)

	// 7 days exclude Oct 17. Pending and failed add orders but not commission.
	// Level 1 success amounts inside the week: 10 + 4. Level 2: 20.
	stats, err := getAgentSystemStatsAt(root.Id, 7, now)
	require.NoError(t, err)
	require.Equal(t, int64(5), stats.OrderCount)
	require.Equal(t, int64(3), stats.SuccessCount)
	require.Equal(t, int64(1), stats.PendingCount)
	require.Equal(t, int64(1), stats.FailedCount)
	require.InDelta(t, 0.6, stats.SuccessRate, 0.000001)
	require.Equal(t, 17000000, stats.RechargeQuota)
	require.Equal(t, 1550000, stats.CommissionQuota)
	require.Len(t, stats.Points, 7)
	require.Equal(t, "2026-10-18", stats.Points[0].Date)
	require.Equal(t, "10-24", stats.Points[6].Label)
	require.Equal(t, int64(0), stats.Points[0].Orders)
	require.Equal(t, int64(1), pointByDate(stats.Points, "2026-10-23").Success)
	require.Equal(t, 5000000, pointByDate(stats.Points, "2026-10-23").RechargeQuota)
	require.Equal(t, 750000, pointByDate(stats.Points, "2026-10-23").CommissionQuota)
	today := pointByDate(stats.Points, "2026-10-24")
	require.Equal(t, int64(4), today.Orders)
	require.Equal(t, int64(2), today.Success)
	require.Equal(t, int64(1), today.Pending)
	require.Equal(t, int64(1), today.Failed)
	require.Equal(t, 12000000, today.RechargeQuota)
	require.Equal(t, 800000, today.CommissionQuota)
	require.Len(t, stats.Payments, 2)
	require.Equal(t, "epay", stats.Payments[0].Method)
	require.Equal(t, int64(1), stats.Payments[0].Count)
	require.Equal(t, 10000000, stats.Payments[0].RechargeQuota)
	require.Equal(t, "stripe", stats.Payments[1].Method)
	require.Equal(t, int64(2), stats.Payments[1].Count)
	require.Equal(t, 7000000, stats.Payments[1].RechargeQuota)

	wider, err := getAgentSystemStatsAt(root.Id, 15, now)
	require.NoError(t, err)
	require.Equal(t, int64(6), wider.OrderCount)
	require.Equal(t, int64(4), wider.SuccessCount)
	require.Equal(t, 2150000, wider.CommissionQuota)

	overview, err := GetAgentSystem(root.Id)
	require.NoError(t, err)
	month, err := getAgentSystemStatsAt(root.Id, 30, now)
	require.NoError(t, err)
	require.Equal(t, overview.Level1CommissionQuota+overview.Level2CommissionQuota, month.CommissionQuota)

	records, err := listAgentSystemRecordsAt(root.Id, "commission", "", 1, 2, now)
	require.NoError(t, err)
	require.Equal(t, int64(4), records.Total)
	require.Equal(t, 800000, records.TodayCommissionQuota)
	require.Equal(t, 750000, records.YesterdayCommissionQuota)
	require.Equal(t, 1550000, records.WeekCommissionQuota)
	require.Len(t, records.Records, 2)
	require.Equal(t, "late-success", records.Records[0].TradeNo)
	require.Equal(t, 1, records.Records[0].Level)
	require.Equal(t, 300000, records.Records[0].CommissionQuota)
	require.Equal(t, "stripe", records.Records[0].PaymentMethod)
	require.Equal(t, "child-today", records.Records[1].TradeNo)
	require.Equal(t, 2, records.Records[1].Level)
	require.Equal(t, "epay", records.Records[1].PaymentMethod)
	require.Equal(t, 500000, records.Records[1].CommissionQuota)

	second, err := listAgentSystemRecordsAt(root.Id, "commission", "", 2, 2, now)
	require.NoError(t, err)
	require.Equal(t, "direct-yesterday", second.Records[0].TradeNo)
	require.Equal(t, "old-success", second.Records[1].TradeNo)

	found, err := listAgentSystemRecordsAt(root.Id, "recharge", "direct-pending", 1, 10, now)
	require.NoError(t, err)
	require.Equal(t, int64(1), found.Total)
	require.Equal(t, common.TopUpStatusPending, found.Records[0].Status)
	require.Equal(t, 0, found.Records[0].CommissionQuota)
	require.Equal(t, 500000000, found.Records[0].RechargeQuota)
	require.Equal(t, 800000, found.TodayCommissionQuota)

	byName, err := listAgentSystemRecordsAt(root.Id, "recharge", "child-stats@example.com", 1, 10, now)
	require.NoError(t, err)
	require.Equal(t, int64(1), byName.Total)
	require.Equal(t, "stats-child", byName.Records[0].Username)

	recharge, err := listAgentSystemRecordsAt(root.Id, "recharge", "", 1, 10, now)
	require.NoError(t, err)
	require.Equal(t, int64(6), recharge.Total)
	for _, row := range recharge.Records {
		require.NotEqual(t, "stranger-success", row.TradeNo)
		require.NotEqual(t, "deleted-success", row.TradeNo)
	}

	_, err = listAgentSystemRecordsAt(root.Id, "commission", strings.Repeat("a", 65), 1, 10, now)
	require.Error(t, err)
	_, err = listAgentSystemRecordsAt(root.Id, "other", "", 1, 10, now)
	require.Error(t, err)
}

func TestListAgentInviteesPagePassesTwoHundred(t *testing.T) {
	truncateTables(t)
	previous := common.QuotaPerUnit
	common.QuotaPerUnit = 500000
	t.Cleanup(func() { common.QuotaPerUnit = previous })

	root := User{Username: "page-root", Password: "password-pass", AffCode: "page-root", Role: common.RoleCommonUser, Status: common.UserStatusEnabled, Group: "default"}
	require.NoError(t, DB.Create(&root).Error)
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	var oldest User
	for i := 0; i <= 200; i++ {
		user := User{
			Username:    "page-user-" + strconv.Itoa(i),
			Password:    "password-pass",
			AffCode:     "p" + strconv.Itoa(i),
			DisplayName: "U" + strconv.Itoa(i),
			Email:       "p" + strconv.Itoa(i) + "@example.com",
			Role:        common.RoleCommonUser,
			Status:      common.UserStatusEnabled,
			Group:       "default",
			InviterId:   root.Id,
			CreatedAt:   base + int64(i),
		}
		require.NoError(t, DB.Create(&user).Error)
		if i == 0 {
			oldest = user
		}
	}
	level2 := User{
		Username:        "page-level2",
		Password:        "password-pass",
		AffCode:         "page-level2",
		DisplayName:     "Level Two",
		Email:           "level2@example.com",
		Role:            common.RoleCommonUser,
		Status:          common.UserStatusEnabled,
		Group:           "default",
		InviterId:       oldest.Id,
		SecondInviterId: root.Id,
		CreatedAt:       base + 1000,
	}
	require.NoError(t, DB.Create(&level2).Error)
	decoy := User{
		Username:    "other-extra",
		Password:    "password-pass",
		AffCode:     "page-decoy",
		DisplayName: "not-a-percent",
		Email:       "decoy@example.com",
		Role:        common.RoleCommonUser,
		Status:      common.UserStatusEnabled,
		Group:       "default",
		InviterId:   root.Id,
		CreatedAt:   base + 50,
	}
	require.NoError(t, DB.Create(&decoy).Error)
	literal := User{
		Username:  "agent_100%",
		Password:  "password-pass",
		AffCode:   "page-literal",
		Email:     "literal@example.com",
		Role:      common.RoleCommonUser,
		Status:    common.UserStatusEnabled,
		Group:     "default",
		InviterId: root.Id,
		CreatedAt: base + 60,
	}
	require.NoError(t, DB.Create(&literal).Error)

	require.NoError(t, DB.Create(&TopUp{UserId: oldest.Id, Amount: 10, Money: 10, TradeNo: "oldest-success", Status: common.TopUpStatusSuccess}).Error)
	require.NoError(t, DB.Create(&TopUp{UserId: level2.Id, Amount: 20, Money: 20, TradeNo: "level2-success", Status: common.TopUpStatusSuccess}).Error)
	require.NoError(t, DB.Create(&TopUp{UserId: oldest.Id, Amount: 1000, Money: 1000, TradeNo: "oldest-pending", Status: common.TopUpStatusPending}).Error)

	report, err := GetAgentSystem(root.Id)
	require.NoError(t, err)
	require.True(t, report.UsersTruncated)
	require.Len(t, report.Users, AgentInviteeListLimit)
	require.Equal(t, int64(203), report.Level1Count)
	require.Equal(t, int64(1), report.Level2Count)

	first, err := ListAgentInviteesPage(root.Id, "", 1, 50)
	require.NoError(t, err)
	require.Equal(t, int64(204), first.Total)
	require.Len(t, first.Users, 50)
	require.Equal(t, "page-level2", first.Users[0].Username)
	require.Equal(t, 2, first.Users[0].Level)
	require.Equal(t, 10000000, first.Users[0].RechargeQuota)
	require.Equal(t, 500000, first.Users[0].CommissionQuota)

	last, err := ListAgentInviteesPage(root.Id, "", 5, 50)
	require.NoError(t, err)
	require.Len(t, last.Users, 4)
	var foundOldest bool
	for _, user := range last.Users {
		if user.Username == "page-user-0" {
			foundOldest = true
			require.Equal(t, 1, user.Level)
			require.Equal(t, int64(1), user.RechargeCount)
			require.Equal(t, 5000000, user.RechargeQuota)
			require.Equal(t, 750000, user.CommissionQuota)
		}
	}
	require.True(t, foundOldest)

	exact, err := ListAgentInviteesPage(root.Id, "page-user-200", 1, 10)
	require.NoError(t, err)
	require.Equal(t, int64(1), exact.Total)
	require.Equal(t, "page-user-200", exact.Users[0].Username)

	wildcard, err := ListAgentInviteesPage(root.Id, "100%", 1, 10)
	require.NoError(t, err)
	require.Equal(t, int64(1), wildcard.Total)
	require.Equal(t, "agent_100%", wildcard.Users[0].Username)

	_, err = ListAgentInviteesPage(root.Id, "", 0, 10)
	require.Error(t, err)
	_, err = ListAgentInviteesPage(root.Id, strings.Repeat("字", 65), 1, 10)
	require.Error(t, err)
}

func pointByDate(points []AgentStatsPoint, date string) AgentStatsPoint {
	for _, point := range points {
		if point.Date == date {
			return point
		}
	}
	return AgentStatsPoint{}
}

func TestAgentSearchTreatsSpecialCharactersLiterally(t *testing.T) {
	truncateTables(t)
	owner := User{Username: "escape-owner", Password: "password-pass", AffCode: "escape-owner", Role: common.RoleCommonUser, Status: common.UserStatusEnabled, Group: "default"}
	require.NoError(t, DB.Create(&owner).Error)
	usernames := []string{"literal!mark", "literal_mark", "literal%mark", `literal\mark`, "literalXmark"}
	for index, username := range usernames {
		user := User{Username: username, Password: "password-pass", AffCode: "escape-" + strconv.Itoa(index), Role: common.RoleCommonUser, Status: common.UserStatusEnabled, Group: "default", InviterId: owner.Id}
		require.NoError(t, DB.Create(&user).Error)
		require.NoError(t, DB.Create(&TopUp{UserId: user.Id, Amount: 1, Money: 1, TradeNo: "escape-order-" + strconv.Itoa(index), Status: common.TopUpStatusSuccess}).Error)
	}
	for _, username := range usernames[:4] {
		t.Run(username, func(t *testing.T) {
			invitees, err := ListAgentInviteesPage(owner.Id, username, 1, 10)
			require.NoError(t, err)
			require.Equal(t, int64(1), invitees.Total)
			require.Len(t, invitees.Users, 1)
			require.Equal(t, username, invitees.Users[0].Username)
			records, err := ListAgentSystemRecords(owner.Id, "recharge", username, 1, 10)
			require.NoError(t, err)
			require.Equal(t, int64(1), records.Total)
			require.Len(t, records.Records, 1)
			require.Equal(t, username, records.Records[0].Username)
			if !strings.ContainsAny(username, "%\\") {
				accounts, err := SearchAgentCommissionAccounts(username, common.RoleRootUser)
				require.NoError(t, err)
				require.Len(t, accounts, 1)
				require.Equal(t, username, accounts[0].Username)
			}
		})
	}
}

func TestEveryUserGetsAnInviteCodeWithoutApplying(t *testing.T) {
	truncateTables(t)

	user := User{
		Username: "agent-hidden",
		Password: "password-pass",
		AffCode:  "hide",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
	}
	require.NoError(t, DB.Create(&user).Error)
	id, err := GetUserIdByAffCode("hide")
	require.NoError(t, err)
	require.Equal(t, user.Id, id)

	report, err := GetAgentSystem(user.Id)
	require.NoError(t, err)
	require.Equal(t, "approved", report.AgentStatus)
	require.Equal(t, "hide", report.InviteCode)

	var stored User
	require.NoError(t, DB.Select("agent_status").First(&stored, user.Id).Error)
	require.Equal(t, AgentStatusApproved, stored.AgentStatus)
	code, err := EnsureUserAffCode(user.Id)
	require.NoError(t, err)
	require.Equal(t, "hide", code)
	report, err = GetAgentSystem(user.Id)
	require.NoError(t, err)
	require.Equal(t, "approved", report.AgentStatus)
	require.Equal(t, "hide", report.InviteCode)
	require.Equal(t, 15, report.Level1Rate)
	require.False(t, report.Level1RateCustom)
	require.Equal(t, 5, report.Level2Rate)
}

func TestBackfillApprovedAgentsEnablesEveryUserAndCreatesMissingCodes(t *testing.T) {
	truncateTables(t)

	inviter := User{Username: "backfill-inviter", Password: "password-pass", AffCode: "bf-inviter", Role: common.RoleCommonUser, Status: common.UserStatusEnabled, Group: "default"}
	require.NoError(t, DB.Create(&inviter).Error)
	child := User{Username: "backfill-child", Password: "password-pass", AffCode: "bf-child", Role: common.RoleCommonUser, Status: common.UserStatusEnabled, Group: "default", InviterId: inviter.Id}
	require.NoError(t, DB.Create(&child).Error)
	counted := User{Username: "backfill-count", Password: "password-pass", AffCode: "bf-count", AffCount: 2, Role: common.RoleCommonUser, Status: common.UserStatusEnabled, Group: "default"}
	require.NoError(t, DB.Create(&counted).Error)
	transferred := User{Username: "backfill-paid", Password: "password-pass", AffCode: "bf-paid", AgentCommissionTransferred: 10, Role: common.RoleCommonUser, Status: common.UserStatusEnabled, Group: "default"}
	require.NoError(t, DB.Create(&transferred).Error)
	deducted := User{Username: "backfill-deducted", Password: "password-pass", AffCode: "bf-deducted", AgentCommissionDeducted: 10, Role: common.RoleCommonUser, Status: common.UserStatusEnabled, Group: "default"}
	require.NoError(t, DB.Create(&deducted).Error)
	unused := User{Username: "backfill-unused", Password: "password-pass", AffCode: "bf-unused", Role: common.RoleCommonUser, Status: common.UserStatusEnabled, Group: "default"}
	require.NoError(t, DB.Create(&unused).Error)
	missingCode := User{Username: "backfill-no-code", Password: "password-pass", Role: common.RoleCommonUser, Status: common.UserStatusEnabled, Group: "default"}
	require.NoError(t, DB.Create(&missingCode).Error)
	secondOnly := User{Username: "backfill-second", Password: "password-pass", AffCode: "bf-second", Role: common.RoleCommonUser, Status: common.UserStatusEnabled, Group: "default"}
	require.NoError(t, DB.Create(&secondOnly).Error)
	grandchild := User{Username: "backfill-grand", Password: "password-pass", AffCode: "bf-grand", Role: common.RoleCommonUser, Status: common.UserStatusEnabled, Group: "default", InviterId: child.Id, SecondInviterId: secondOnly.Id}
	require.NoError(t, DB.Create(&grandchild).Error)

	require.NoError(t, BackfillApprovedAgents())

	status := func(id int) int {
		var user User
		require.NoError(t, DB.Select("agent_status").First(&user, id).Error)
		return user.AgentStatus
	}
	require.Equal(t, AgentStatusApproved, status(inviter.Id))
	require.Equal(t, AgentStatusApproved, status(child.Id))
	require.Equal(t, AgentStatusApproved, status(counted.Id))
	require.Equal(t, AgentStatusApproved, status(transferred.Id))
	require.Equal(t, AgentStatusApproved, status(deducted.Id))
	require.Equal(t, AgentStatusApproved, status(unused.Id))
	require.Equal(t, AgentStatusApproved, status(secondOnly.Id))
	require.Equal(t, AgentStatusApproved, status(missingCode.Id))
	id, err := GetUserIdByAffCode("bf-unused")
	require.NoError(t, err)
	require.Equal(t, unused.Id, id)
	var migrated User
	require.NoError(t, DB.Select("aff_code").First(&migrated, unused.Id).Error)
	require.Equal(t, "bf-unused", migrated.AffCode)
	// Startup backfill is safe to run repeatedly and preserves existing codes.
	require.NoError(t, BackfillApprovedAgents())
	var migratedAgain User
	require.NoError(t, DB.Select("aff_code").First(&migratedAgain, unused.Id).Error)
	require.Equal(t, migrated.AffCode, migratedAgain.AffCode)
	require.NoError(t, DB.Select("aff_code").First(&migrated, missingCode.Id).Error)
	require.Len(t, migrated.AffCode, 4)
	id, err = GetUserIdByAffCode("bf-inviter")
	require.NoError(t, err)
	require.Equal(t, inviter.Id, id)
}

func TestCustomAgentRateAppliesOnlyToLaterOrders(t *testing.T) {
	truncateTables(t)
	previous := common.QuotaPerUnit
	common.QuotaPerUnit = 500000
	t.Cleanup(func() { common.QuotaPerUnit = previous })

	root := User{Username: "rate-root", Password: "password-pass", AffCode: "rate-root", Role: common.RoleCommonUser, Status: common.UserStatusEnabled, Group: "default", AgentStatus: AgentStatusApproved}
	require.NoError(t, DB.Create(&root).Error)
	direct := User{Username: "rate-direct", Password: "password-pass", AffCode: "rate-direct", Role: common.RoleCommonUser, Status: common.UserStatusEnabled, Group: "default", InviterId: root.Id}
	require.NoError(t, DB.Create(&direct).Error)
	child := User{Username: "rate-child", Password: "password-pass", AffCode: "rate-child", Role: common.RoleCommonUser, Status: common.UserStatusEnabled, Group: "default", InviterId: direct.Id, SecondInviterId: root.Id}
	require.NoError(t, DB.Create(&child).Error)

	require.NoError(t, DB.Create(&TopUp{UserId: direct.Id, Amount: 10, Money: 10, TradeNo: "historical-l1", Status: common.TopUpStatusSuccess}).Error)
	require.NoError(t, DB.Create(&TopUp{UserId: child.Id, Amount: 20, Money: 20, TradeNo: "historical-l2", Status: common.TopUpStatusSuccess}).Error)

	report, err := GetAgentSystem(root.Id)
	require.NoError(t, err)
	require.Equal(t, 750000, report.Level1CommissionQuota)
	require.Equal(t, 500000, report.Level2CommissionQuota)
	require.Equal(t, 15, report.Level1Rate)
	require.False(t, report.Level1RateCustom)

	_, err = SetAgentCommissionRates(direct.Id, 30, 5)
	require.NoError(t, err)
	updated, err := SetAgentCommissionRates(root.Id, 20, 10)
	require.NoError(t, err)
	require.Equal(t, 20, updated.Level1Rate)
	require.True(t, updated.Level1Custom)
	require.Equal(t, 10, updated.Level2Rate)
	require.True(t, updated.Level2Custom)

	var historical TopUp
	require.NoError(t, DB.Where("trade_no = ?", "historical-l1").First(&historical).Error)
	require.Nil(t, historical.Level1CommissionPercent)
	report, err = GetAgentSystem(root.Id)
	require.NoError(t, err)
	require.Equal(t, 750000, report.Level1CommissionQuota)
	require.Equal(t, 500000, report.Level2CommissionQuota)

	saveSnapshotted := func(tradeNo string, userID int, amount int64) TopUp {
		t.Helper()
		order := &TopUp{UserId: userID, Amount: amount, Money: float64(amount), TradeNo: tradeNo, Status: common.TopUpStatusPending}
		require.NoError(t, DB.Create(order).Error)
		require.NoError(t, DB.Transaction(func(tx *gorm.DB) error {
			if err := snapshotAgentCommissionRates(tx, order); err != nil {
				return err
			}
			return tx.Model(&TopUp{}).Where("id = ?", order.Id).Updates(map[string]any{
				"status":                    common.TopUpStatusSuccess,
				"level1_commission_percent": order.Level1CommissionPercent,
				"level2_commission_percent": order.Level2CommissionPercent,
			}).Error
		}))
		var stored TopUp
		require.NoError(t, DB.First(&stored, order.Id).Error)
		return stored
	}

	fresh := saveSnapshotted("fresh-l1", direct.Id, 10)
	require.NotNil(t, fresh.Level1CommissionPercent)
	require.Equal(t, 20, *fresh.Level1CommissionPercent)
	freshChild := saveSnapshotted("fresh-l2", child.Id, 20)
	require.NotNil(t, freshChild.Level2CommissionPercent)
	require.Equal(t, 10, *freshChild.Level2CommissionPercent)

	report, err = GetAgentSystem(root.Id)
	require.NoError(t, err)
	// Historical NULL stays 15%/5%. New rows use 20% and 10%.
	require.Equal(t, 1750000, report.Level1CommissionQuota)
	require.Equal(t, 1500000, report.Level2CommissionQuota)

	_, err = SetAgentCommissionRates(root.Id, 0, -1)
	require.NoError(t, err)
	zero := saveSnapshotted("zero-l1", direct.Id, 10)
	require.NotNil(t, zero.Level1CommissionPercent)
	require.Equal(t, 0, *zero.Level1CommissionPercent)
	require.NotNil(t, zero.Level2CommissionPercent)
	require.Equal(t, 0, *zero.Level2CommissionPercent)
	report, err = GetAgentSystem(root.Id)
	require.NoError(t, err)
	require.Equal(t, 0, report.Level1Rate)
	require.True(t, report.Level1RateCustom)
	require.Equal(t, 5, report.Level2Rate)
	require.False(t, report.Level2RateCustom)
	require.Equal(t, 1750000, report.Level1CommissionQuota)

	pending := User{Username: "rate-agent", Password: "password-pass", AffCode: "rate-agent", Role: common.RoleCommonUser, Status: common.UserStatusEnabled, Group: "default", AgentStatus: AgentStatusApproved}
	require.NoError(t, DB.Create(&pending).Error)
	payer := User{Username: "rate-payer", Password: "password-pass", AffCode: "rate-payer", Role: common.RoleCommonUser, Status: common.UserStatusEnabled, Group: "default", InviterId: pending.Id}
	require.NoError(t, DB.Create(&payer).Error)
	newAgentOrder := saveSnapshotted("automatic-agent-order", payer.Id, 10)
	require.NotNil(t, newAgentOrder.Level1CommissionPercent)
	require.Equal(t, 15, *newAgentOrder.Level1CommissionPercent)
}

func TestAgentCommissionFloorsOncePerStoredPercent(t *testing.T) {
	truncateTables(t)
	previous := common.QuotaPerUnit
	common.QuotaPerUnit = 1
	t.Cleanup(func() { common.QuotaPerUnit = previous })

	root := User{Username: "floor-root", Password: "password-pass", AffCode: "floor-root", Role: common.RoleCommonUser, Status: common.UserStatusEnabled, Group: "default", AgentStatus: AgentStatusApproved}
	require.NoError(t, DB.Create(&root).Error)
	direct := User{Username: "floor-direct", Password: "password-pass", AffCode: "floor-direct", Role: common.RoleCommonUser, Status: common.UserStatusEnabled, Group: "default", InviterId: root.Id}
	require.NoError(t, DB.Create(&direct).Error)
	now := time.Now().Unix()
	require.NoError(t, DB.Create(&TopUp{UserId: direct.Id, Amount: 10, Money: 10, TradeNo: "floor-a", Status: common.TopUpStatusSuccess, CreateTime: now, CompleteTime: now}).Error)
	require.NoError(t, DB.Create(&TopUp{UserId: direct.Id, Amount: 10, Money: 10, TradeNo: "floor-b", Status: common.TopUpStatusSuccess, CreateTime: now, CompleteTime: now}).Error)

	commission, amount, count, err := sumAgentLevelCommission(DB, root.Id, 1, nil, nil)
	require.NoError(t, err)
	require.Equal(t, 3, commission)
	require.Equal(t, int64(20), amount)
	require.Equal(t, int64(2), count)

	stats, err := getAgentSystemStatsAt(root.Id, 7, time.Unix(now, 0))
	require.NoError(t, err)
	require.Equal(t, 3, stats.CommissionQuota)
}

func TestListApprovedAgentsShowsQuotaAndCommission(t *testing.T) {
	truncateTables(t)
	previous := common.QuotaPerUnit
	common.QuotaPerUnit = 500000
	t.Cleanup(func() { common.QuotaPerUnit = previous })

	level1Rate := 20
	level2Rate := 0
	agent := User{
		Username:                   "agent-main",
		DisplayName:                "Main Agent",
		Password:                   "password-pass",
		AffCode:                    "agent-main-code",
		Role:                       common.RoleCommonUser,
		Status:                     common.UserStatusEnabled,
		Group:                      "default",
		Quota:                      800000,
		UsedQuota:                  123,
		AgentStatus:                AgentStatusApproved,
		AgentLevel1Rate:            &level1Rate,
		AgentLevel2Rate:            &level2Rate,
		AgentCommissionTransferred: 100000,
		AgentCommissionDeducted:    50000,
		AgentAppliedAt:             100,
		AgentReviewedAt:            9000000000,
	}
	require.NoError(t, DB.Create(&agent).Error)
	direct := User{Username: "agent-direct", Password: "password-pass", AffCode: "agent-direct-code", Role: common.RoleCommonUser, Status: common.UserStatusEnabled, Group: "default", InviterId: agent.Id}
	require.NoError(t, DB.Create(&direct).Error)
	second := User{Username: "agent-second", Password: "password-pass", AffCode: "agent-second-code", Role: common.RoleCommonUser, Status: common.UserStatusEnabled, Group: "default", InviterId: direct.Id, SecondInviterId: agent.Id}
	require.NoError(t, DB.Create(&second).Error)
	require.NoError(t, DB.Create(&TopUp{UserId: direct.Id, Amount: 10, Money: 10, TradeNo: "agent-direct-ok", Status: common.TopUpStatusSuccess}).Error)
	require.NoError(t, DB.Create(&TopUp{UserId: direct.Id, Amount: 1000, Money: 1000, TradeNo: "agent-direct-wait", Status: common.TopUpStatusPending}).Error)
	require.NoError(t, DB.Create(&TopUp{UserId: second.Id, Amount: 20, Money: 20, TradeNo: "agent-second-ok", Status: common.TopUpStatusSuccess}).Error)
	zero := 0
	require.NoError(t, DB.Create(&TopUp{UserId: direct.Id, Amount: 4, Money: 4, TradeNo: "agent-direct-zero", Status: common.TopUpStatusSuccess, Level1CommissionPercent: &zero}).Error)

	higher := User{Username: "agent-higher", Password: "password-pass", AffCode: "agent-higher-code", Role: common.RoleRootUser, Status: common.UserStatusEnabled, Group: "default", AgentStatus: AgentStatusApproved}
	require.NoError(t, DB.Create(&higher).Error)
	pending := User{Username: "agent-wait", Password: "password-pass", AffCode: "agent-wait-code", Role: common.RoleCommonUser, Status: common.UserStatusEnabled, Group: "default"}
	require.NoError(t, DB.Create(&pending).Error)
	require.NoError(t, BackfillApprovedAgents())
	rejected := User{Username: "agent-no", Password: "password-pass", AffCode: "agent-no-code", Role: common.RoleCommonUser, Status: common.UserStatusEnabled, Group: "default"}
	require.NoError(t, DB.Create(&rejected).Error)
	_, err := EnsureUserAffCode(rejected.Id)
	require.NoError(t, err)

	fillerIDs := make([]int, 0, 20)
	for i := 0; i < 20; i++ {
		filler := User{
			Username:    "agent-fill-" + strconv.Itoa(i),
			Password:    "password-pass",
			AffCode:     "agent-fill-" + strconv.Itoa(i),
			Role:        common.RoleCommonUser,
			Status:      common.UserStatusEnabled,
			Group:       "default",
			AgentStatus: AgentStatusApproved,
		}
		require.NoError(t, DB.Create(&filler).Error)
		fillerIDs = append(fillerIDs, filler.Id)
	}

	page, err := ListApprovedAgents(common.RoleAdminUser, 1, 20)
	require.NoError(t, err)
	require.Equal(t, int64(25), page.Total)
	require.Len(t, page.Items, 20)
	allItems := append([]ApprovedAgent{}, page.Items...)
	next, err := ListApprovedAgents(common.RoleAdminUser, 2, 20)
	require.NoError(t, err)
	allItems = append(allItems, next.Items...)
	var first ApprovedAgent
	for _, item := range allItems {
		if item.Id == agent.Id {
			first = item
			break
		}
	}
	require.NotZero(t, first.Id)
	require.Equal(t, agent.Id, first.Id)
	require.Equal(t, "agent-main-code", first.InviteCode)
	require.Equal(t, 800000, first.Quota)
	require.Equal(t, 123, first.UsedQuota)
	require.Equal(t, int64(1), first.Level1Count)
	require.Equal(t, int64(1), first.Level2Count)
	require.Equal(t, 750000, first.Level1CommissionQuota)
	require.Equal(t, 500000, first.Level2CommissionQuota)
	require.Equal(t, 1250000, first.TotalQuota)
	require.Equal(t, 100000, first.TransferredQuota)
	require.Equal(t, 50000, first.DeductedQuota)
	require.Equal(t, 1100000, first.AvailableQuota)
	require.Equal(t, 20, first.Level1Rate)
	require.True(t, first.Level1Custom)
	require.Equal(t, 0, first.Level2Rate)
	require.True(t, first.Level2Custom)

	nextPage, err := ListApprovedAgents(common.RoleAdminUser, 2, 20)
	require.NoError(t, err)
	require.Len(t, nextPage.Items, 5)

	rootPage, err := ListApprovedAgents(common.RoleRootUser, 1, 20)
	require.NoError(t, err)
	require.Equal(t, int64(26), rootPage.Total)
	rootNext, err := ListApprovedAgents(common.RoleRootUser, 2, 20)
	require.NoError(t, err)
	foundHigher := false
	for _, item := range append(rootPage.Items, rootNext.Items...) {
		if item.Id == higher.Id {
			foundHigher = true
		}
	}
	require.True(t, foundHigher)
	for _, item := range page.Items {
		require.NotEqual(t, higher.Id, item.Id)
	}

	_, err = ListApprovedAgents(common.RoleAdminUser, 1, 10)
	require.Error(t, err)
	_, err = ListApprovedAgents(common.RoleAdminUser, 0, 20)
	require.Error(t, err)
}
