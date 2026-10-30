package controller

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"

	"github.com/gin-gonic/gin"
)

func GetGroupStatus(c *gin.Context) {
	userID := c.GetInt("id")
	role := c.GetInt("role")
	userGroup, err := model.GetUserGroup(userID, false)
	if err != nil {
		logger.LogError(c, common.LogText("failed to load user group for group status: %s", err.Error()))
		common.ApiErrorT(c, "Failed to load group status")
		return
	}
	modelNames, err := model.ListEnabledModelsByGroup()
	if err != nil {
		logger.LogError(c, common.LogText("failed to load group models for group status: %s", err.Error()))
		common.ApiErrorT(c, "Failed to load group status")
		return
	}

	names := visibleGroupStatusNames(role, userGroup)
	now := common.GetTimestamp()
	start, end := model.GroupStatusBounds(now)
	logs, err := model.ListGroupStatusLogs(start, end, names)
	if err != nil {
		logger.LogError(c, common.LogText("failed to load group status logs: %s", err.Error()))
		common.ApiErrorT(c, "Failed to load group status")
		return
	}

	usable := service.GetUserUsableGroups(userGroup)
	inputs := make([]model.GroupStatusInput, 0, len(names))
	for _, name := range names {
		description := usable[name]
		if description == "" {
			description = setting.GetUsableGroupDescription(name)
		}
		inputs = append(inputs, model.GroupStatusInput{
			Name:        name,
			Description: description,
			Ratio:       service.GetUserGroupRatio(userGroup, name),
			Models:      modelNames[name],
		})
	}
	common.ApiSuccess(c, model.AggregateGroupStatus(inputs, logs, now))
}

func visibleGroupStatusNames(role int, userGroup string) []string {
	names := make([]string, 0)
	if role >= common.RoleAdminUser {
		for name := range ratio_setting.GetGroupRatioCopy() {
			if name == "" || name == "auto" {
				continue
			}
			names = append(names, name)
		}
		return names
	}
	for name := range service.GetUserUsableGroups(userGroup) {
		if service.IsUserSelectableGroup(userGroup, name) {
			names = append(names, name)
		}
	}
	return names
}
