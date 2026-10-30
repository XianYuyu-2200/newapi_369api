package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestBindInviteRelationStopsAtSecondLevel(t *testing.T) {
	truncateTables(t)

	root := User{
		Username: "root-agent",
		Password: "password-pass",
		AffCode:  "root-aff",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
	}
	require.NoError(t, DB.Create(&root).Error)

	direct := User{
		Username:  "direct-agent",
		Password:  "password-pass",
		AffCode:   "direct-aff",
		InviterId: root.Id,
		Role:      common.RoleCommonUser,
		Status:    common.UserStatusEnabled,
		Group:     "default",
	}
	require.NoError(t, DB.Create(&direct).Error)

	child := User{
		Username: "new-user",
		Password: "password-pass",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
	}
	require.NoError(t, child.Insert(direct.Id))

	stored := loadInviteRelation(t, child.Id)
	require.Equal(t, direct.Id, stored.InviterId)
	require.Equal(t, root.Id, stored.SecondInviterId)

	grandchild := User{
		Username: "grand-user",
		Password: "password-pass",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
	}
	require.NoError(t, grandchild.Insert(child.Id))
	storedGrand := loadInviteRelation(t, grandchild.Id)
	require.Equal(t, child.Id, storedGrand.InviterId)
	require.Equal(t, direct.Id, storedGrand.SecondInviterId)

	require.NoError(t, DB.Model(&User{}).Where("id = ?", child.Id).Update("second_inviter_id", 0).Error)
	require.NoError(t, BackfillSecondInviterIDs())
	stored = loadInviteRelation(t, child.Id)
	require.Equal(t, root.Id, stored.SecondInviterId)

	require.NoError(t, DB.Model(&User{}).Where("id = ?", child.Id).Update("second_inviter_id", 99).Error)
	require.NoError(t, BackfillSecondInviterIDs())
	stored = loadInviteRelation(t, child.Id)
	require.Equal(t, 99, stored.SecondInviterId)

	orphan := User{
		Username: "orphan-user",
		Password: "password-pass",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
	}
	require.NoError(t, orphan.Insert(999999))
	stored = loadInviteRelation(t, orphan.Id)
	require.Zero(t, stored.InviterId)
	require.Zero(t, stored.SecondInviterId)

	loop := User{
		Username: "loop-user",
		Password: "password-pass",
		AffCode:  "loop-aff",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
	}
	require.NoError(t, DB.Create(&loop).Error)
	require.NoError(t, DB.Model(&User{}).Where("id = ?", loop.Id).Update("inviter_id", loop.Id).Error)
	viaLoop := User{
		Username: "via-loop",
		Password: "password-pass",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
	}
	require.NoError(t, viaLoop.Insert(loop.Id))
	stored = loadInviteRelation(t, viaLoop.Id)
	require.Equal(t, loop.Id, stored.InviterId)
	require.Zero(t, stored.SecondInviterId)

	oauthUser := User{
		Username: "oauth-child",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
	}
	require.NoError(t, DB.Transaction(func(tx *gorm.DB) error {
		return oauthUser.InsertWithTx(tx, direct.Id)
	}))
	stored = loadInviteRelation(t, oauthUser.Id)
	require.Equal(t, direct.Id, stored.InviterId)
	require.Equal(t, root.Id, stored.SecondInviterId)
}

func loadInviteRelation(t *testing.T, id int) User {
	t.Helper()
	var user User
	require.NoError(t, DB.Where("id = ?", id).Select("id", "inviter_id", "second_inviter_id").Take(&user).Error)
	return user
}
