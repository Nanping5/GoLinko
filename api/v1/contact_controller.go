package v1

import (
	"GoLinko/internal/dto/request"
	gormss "GoLinko/internal/service/gorms"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetContactUserList 获取联系人列表
func GetContactUserList(c *gin.Context) {
	userId := c.GetString("userId")
	if userId == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}
	msg, userList, ret := gormss.ContactInfoService.GetContactUserList(c.Request.Context(), userId)
	JsonBack(c, msg, ret, userList)
}

// 获取联系人详细信息（联系人有群聊和用户）
func GetContactInfo(c *gin.Context) {
	userId := c.GetString("userId")
	if userId == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}
	contactId := c.Query("contact_id")
	msg, userInfo, ret := gormss.ContactInfoService.GetContactInfo(c.Request.Context(), userId, contactId)
	JsonBack(c, msg, ret, userInfo)
}

// LoadMyJoinedGroups 加载我加入的群列表
func LoadMyJoinedGroups(c *gin.Context) {
	userId := c.GetString("userId")
	if userId == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}

	msg, groupList, ret := gormss.ContactInfoService.LoadMyJoinedGroups(c.Request.Context(), userId)
	JsonBack(c, msg, ret, groupList)
}

// DeleteContact 删除联系人
func DeleteContact(c *gin.Context) {
	userId := c.GetString("userId")
	if userId == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}
	contactId := c.Query("contact_id")
	msg, ret := gormss.ContactInfoService.DeleteContact(c.Request.Context(), userId, contactId)
	JsonBack(c, msg, ret, nil)
}

// ApplyAddContact 申请添加联系人或群聊
func ApplyAddContact(c *gin.Context) {
	userId := c.GetString("userId")
	if userId == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}
	var req request.ApplyAddContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}
	msg, ret := gormss.ContactInfoService.ApplyAddContact(c.Request.Context(), userId, req)
	JsonBack(c, msg, ret, nil)
}

// GetContactApplyList 获取新的联系人/群聊申请列表
func GetContactApplyList(c *gin.Context) {
	userId := c.GetString("userId")
	if userId == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}
	msg, applyList, ret := gormss.ContactInfoService.GetContactApplyList(c.Request.Context(), userId)
	JsonBack(c, msg, ret, applyList)
}

// PassContactApply 同意联系人申请
func AcceptContactApply(c *gin.Context) {
	userId := c.GetString("userId")
	if userId == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}
	applyId := c.Query("apply_id")
	if applyId == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "缺少申请ID",
		})
		return
	}
	msg, ret := gormss.ContactInfoService.AcceptAddContact(c.Request.Context(), userId, applyId)
	JsonBack(c, msg, ret, nil)
}

// RejectContactApply 拒绝联系人申请
func RejectContactApply(c *gin.Context) {
	userId := c.GetString("userId")
	if userId == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}
	applyId := c.Query("apply_id")
	if applyId == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "缺少申请ID",
		})
		return
	}
	msg, ret := gormss.ContactInfoService.RejectAddContact(c.Request.Context(), userId, applyId)
	JsonBack(c, msg, ret, nil)
}

// BlackContact 拉黑联系人
func BlackContact(c *gin.Context) {
	userId := c.GetString("userId")
	if userId == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}
	contactId := c.Query("contact_id")
	if contactId == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "缺少联系人ID",
		})
		return
	}
	msg, ret := gormss.ContactInfoService.BlackContact(c.Request.Context(), userId, contactId)
	JsonBack(c, msg, ret, nil)
}

// UnblackContact 解除拉黑联系人
func UnblackContact(c *gin.Context) {
	userId := c.GetString("userId")
	if userId == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}
	contactId := c.Query("contact_id")
	if contactId == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "缺少联系人ID",
		})
		return
	}
	msg, ret := gormss.ContactInfoService.UnblackContact(c.Request.Context(), userId, contactId)
	JsonBack(c, msg, ret, nil)
}
