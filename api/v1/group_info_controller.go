package v1

import (
	"GoLinko/internal/dto/request"
	gormss "GoLinko/internal/service/gorms"
	constants "GoLinko/pkg/const"
	"GoLinko/pkg/zlog"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CreateGroup 创建群组
func CreateGroup(c *gin.Context) {
	createGroupReq := request.CreateGroupRequest{}
	if err := c.ShouldBindJSON(&createGroupReq); err != nil {
		zlog.GetLogger().Error("创建群组参数绑定失败", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constants.SYSTEM_ERROR,
		})
		return
	}
	// 强制使用 Token 中的 userId 作为群主 ID
	userId := c.GetString("userId")
	//调用 service 层的创建群组逻辑
	msg, groupId, ret := gormss.GroupInfoService.CreateGroup(c.Request.Context(), &createGroupReq, userId)

	JsonBack(c, msg, ret, gin.H{"group_id": groupId})
}

// LoadMyGroups 加载我的群组列表
func LoadMyGroups(c *gin.Context) {

	userId := c.GetString("userId")
	if userId == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}
	msg, groups, ret := gormss.GroupInfoService.LoadMyGroups(c.Request.Context(), userId)
	JsonBack(c, msg, ret, groups)
}

// 检查加群方式
func CheckGroupAddMode(c *gin.Context) {
	groupId := c.Query("group_id")
	if groupId == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "缺少群ID",
		})
		return
	}
	msg, addMode, ret := gormss.GroupInfoService.CheckGroupAddMode(c.Request.Context(), groupId)
	JsonBack(c, msg, ret, gin.H{"add_mode": addMode})
}

// EnterGroupDirectly 直接入群
func EnterGroupDirectly(c *gin.Context) {
	req := request.EnterGroupDirectlyRequest{}
	req.GroupId = c.Query("group_id")
	if req.GroupId == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "缺少群ID",
		})
		return
	}
	userId := c.GetString("userId")
	if userId == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}
	msg, ret := gormss.GroupInfoService.EnterGroupDirectly(c.Request.Context(), req.GroupId, userId)
	JsonBack(c, msg, ret, nil)
}

func LeaveGroup(c *gin.Context) {
	req := request.LeaveGroupRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.GetLogger().Error("离开群组参数绑定失败", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constants.SYSTEM_ERROR,
		})
		return
	}
	userId := c.GetString("userId")
	if userId == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}
	msg, ret := gormss.GroupInfoService.LeaveGroup(c.Request.Context(), req.GroupId, userId)
	JsonBack(c, msg, ret, nil)
}

// 解散群组
func DismissGroup(c *gin.Context) {
	req := request.DismissGroupRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.GetLogger().Error("解散群组参数绑定失败", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constants.SYSTEM_ERROR,
		})
		return
	}
	userId := c.GetString("userId")
	if userId == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}
	msg, ret := gormss.GroupInfoService.DismissGroup(c.Request.Context(), req.GroupId, userId)
	JsonBack(c, msg, ret, nil)
}

// GetGroupInfo 获取群信息
func GetGroupInfo(c *gin.Context) {
	groupId := c.Query("group_id")
	if groupId == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "缺少群ID",
		})
		return
	}
	msg, groupInfo, ret := gormss.GroupInfoService.GetGroupInfo(c.Request.Context(), groupId)
	JsonBack(c, msg, ret, groupInfo)
}

// UpdateGroupInfo 更新群信息
func UpdateGroupInfo(c *gin.Context) {
	req := request.UpdateGroupInfoRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.GetLogger().Error("更新群组信息参数绑定失败", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constants.SYSTEM_ERROR,
		})
		return
	}
	operatorId := c.GetString("userId")
	if operatorId == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}
	msg, ret := gormss.GroupInfoService.UpdateGroupInfo(c.Request.Context(), req, operatorId)
	JsonBack(c, msg, ret, nil)
}

// GetGroupMembers 获取群成员列表
func GetGroupMembers(c *gin.Context) {
	groupId := c.Query("group_id")
	if groupId == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "缺少群ID",
		})
		return
	}
	msg, members, ret := gormss.GroupInfoService.GetGroupMembers(c.Request.Context(), groupId)
	JsonBack(c, msg, ret, members)
}

// RemoveGroupMember 移除群成员
func RemoveGroupMember(c *gin.Context) {

	var req request.RemoveGroupMemberRequest
	// 兼容两种调用方式：
	// 1) DELETE /v1/remove_group_member?group_id=...&user_id=...
	// 2) DELETE body: {"group_id":"...","user_id":"..."}
	if err := c.ShouldBindQuery(&req); err != nil || req.GroupId == "" || req.UserId == "" {
		if err2 := c.ShouldBindJSON(&req); err2 != nil {
			zlog.GetLogger().Error("移除群成员参数绑定失败", zap.Error(err2))
			c.JSON(http.StatusOK, gin.H{
				"code":    500,
				"message": constants.SYSTEM_ERROR,
			})
			return
		}
	}
	operatorId := c.GetString("userId")
	if operatorId == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}
	msg, ret := gormss.GroupInfoService.RemoveGroupMember(c.Request.Context(), req.GroupId, req.UserId, operatorId)
	JsonBack(c, msg, ret, nil)
}
