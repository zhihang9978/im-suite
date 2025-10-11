package controller

import (
	"net/http"
	"strconv"
	"zhihang-messenger/im-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// GroupManagementController 群组管理控制器
type GroupManagementController struct {
	groupMgmtService *service.GroupManagementService
}

// NewGroupManagementController 创建群组管理控制器实例
func NewGroupManagementController(groupMgmtService *service.GroupManagementService) *GroupManagementController {
	return &GroupManagementController{
		groupMgmtService: groupMgmtService,
	}
}

// CreateInvite 创建邀请链接
func (c *GroupManagementController) CreateInvite(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")

	var req service.CreateInviteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	req.CreatorID = userID.(uint)

	invite, err := c.groupMgmtService.CreateInvite(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "邀请创建成功",
		"invite":  invite,
	})
}

// UseInvite 使用邀请
func (c *GroupManagementController) UseInvite(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	inviteCode := ctx.Param("invite_code")
	ipAddress := ctx.ClientIP()

	if err := c.groupMgmtService.UseInvite(inviteCode, userID.(uint), ipAddress); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "加入群组成功"})
}

// RevokeInvite 撤销邀请
func (c *GroupManagementController) RevokeInvite(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	inviteIDStr := ctx.Param("invite_id")
	inviteID, _ := strconv.ParseUint(inviteIDStr, 10, 32)

	var req struct {
		Reason string `json:"reason"`
	}
	ctx.ShouldBindJSON(&req)

	if err := c.groupMgmtService.RevokeInvite(uint(inviteID), userID.(uint), req.Reason); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "邀请已撤销"})
}

// GetChatInvites 获取群组邀请列表
func (c *GroupManagementController) GetChatInvites(ctx *gin.Context) {
	chatIDStr := ctx.Param("chat_id")
	chatID, _ := strconv.ParseUint(chatIDStr, 10, 32)
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	invites, total, err := c.groupMgmtService.GetChatInvites(uint(chatID), limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"invites": invites,
		"total":   total,
	})
}

// ApproveJoinRequest 审批入群申请
func (c *GroupManagementController) ApproveJoinRequest(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")

	var req service.ApproveJoinRequestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	req.ReviewerID = userID.(uint)

	if err := c.groupMgmtService.ApproveJoinRequest(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "申请已处理"})
}

// GetPendingJoinRequests 获取待审批的入群申请
func (c *GroupManagementController) GetPendingJoinRequests(ctx *gin.Context) {
	chatIDStr := ctx.Param("chat_id")
	chatID, _ := strconv.ParseUint(chatIDStr, 10, 32)
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	requests, total, err := c.groupMgmtService.GetPendingJoinRequests(uint(chatID), limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"requests": requests,
		"total":    total,
	})
}

// PromoteMember 提升管理员
func (c *GroupManagementController) PromoteMember(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")

	var req service.PromoteAdminRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	req.PromotedBy = userID.(uint)

	if err := c.groupMgmtService.PromoteMember(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "提升管理员成功"})
}

// DemoteMember 降级管理员
func (c *GroupManagementController) DemoteMember(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	chatIDStr := ctx.Param("chat_id")
	targetUserIDStr := ctx.Param("user_id")

	chatID, _ := strconv.ParseUint(chatIDStr, 10, 32)
	targetUserID, _ := strconv.ParseUint(targetUserIDStr, 10, 32)

	if err := c.groupMgmtService.DemoteMember(uint(chatID), uint(targetUserID), userID.(uint)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "降级管理员成功"})
}

// GetChatAdmins 获取群组管理员列表
func (c *GroupManagementController) GetChatAdmins(ctx *gin.Context) {
	chatIDStr := ctx.Param("chat_id")
	chatID, _ := strconv.ParseUint(chatIDStr, 10, 32)

	admins, err := c.groupMgmtService.GetChatAdmins(uint(chatID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"admins": admins})
}

// GetAuditLogs 获取审计日志
func (c *GroupManagementController) GetAuditLogs(ctx *gin.Context) {
	chatIDStr := ctx.Param("chat_id")
	chatID, _ := strconv.ParseUint(chatIDStr, 10, 32)
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	logs, total, err := c.groupMgmtService.GetAuditLogs(uint(chatID), limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"total": total,
	})
}
