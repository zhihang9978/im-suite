package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"zhihang-messenger/im-backend/internal/service"
)

// ChatManagementController 群组管理控制器
type ChatManagementController struct {
	permissionService    *service.ChatPermissionService
	announcementService  *service.ChatAnnouncementService
	statisticsService    *service.ChatStatisticsService
	backupService        *service.ChatBackupService
}

// NewChatManagementController 创建群组管理控制器
func NewChatManagementController(
	permissionService *service.ChatPermissionService,
	announcementService *service.ChatAnnouncementService,
	statisticsService *service.ChatStatisticsService,
	backupService *service.ChatBackupService,
) *ChatManagementController {
	return &ChatManagementController{
		permissionService:   permissionService,
		announcementService: announcementService,
		statisticsService:   statisticsService,
		backupService:       backupService,
	}
}

// 权限管理相关接口

// SetChatPermissions 设置群组权限
func (c *ChatManagementController) SetChatPermissions(ctx *gin.Context) {
	var req service.SetPermissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.GetUint("user_id")
	if err := c.permissionService.SetChatPermissions(ctx, userID, &req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "权限设置成功"})
}

// GetChatPermissions 获取群组权限
func (c *ChatManagementController) GetChatPermissions(ctx *gin.Context) {
	chatIDStr := ctx.Param("chat_id")
	chatID, err := strconv.ParseUint(chatIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的群组ID"})
		return
	}

	userID := ctx.GetUint("user_id")
	permissions, err := c.permissionService.GetChatPermissions(ctx, uint(chatID), userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": permissions})
}

// MuteMember 禁言成员
func (c *ChatManagementController) MuteMember(ctx *gin.Context) {
	var req service.MuteMemberRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.GetUint("user_id")
	if err := c.permissionService.MuteMember(ctx, userID, &req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "成员禁言成功"})
}

// UnmuteMember 解除禁言
func (c *ChatManagementController) UnmuteMember(ctx *gin.Context) {
	chatIDStr := ctx.Param("chat_id")
	userIDStr := ctx.Param("user_id")
	
	chatID, err := strconv.ParseUint(chatIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的群组ID"})
		return
	}

	targetUserID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	userID := ctx.GetUint("user_id")
	if err := c.permissionService.UnmuteMember(ctx, userID, uint(chatID), uint(targetUserID)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "解除禁言成功"})
}

// BanMember 踢出成员
func (c *ChatManagementController) BanMember(ctx *gin.Context) {
	var req service.BanMemberRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.GetUint("user_id")
	if err := c.permissionService.BanMember(ctx, userID, &req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "成员踢出成功"})
}

// UnbanMember 解除封禁
func (c *ChatManagementController) UnbanMember(ctx *gin.Context) {
	chatIDStr := ctx.Param("chat_id")
	userIDStr := ctx.Param("user_id")
	
	chatID, err := strconv.ParseUint(chatIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的群组ID"})
		return
	}

	targetUserID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	userID := ctx.GetUint("user_id")
	if err := c.permissionService.UnbanMember(ctx, userID, uint(chatID), uint(targetUserID)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "解除封禁成功"})
}

// PromoteMember 提升成员权限
func (c *ChatManagementController) PromoteMember(ctx *gin.Context) {
	var req service.PromoteMemberRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.GetUint("user_id")
	if err := c.permissionService.PromoteMember(ctx, userID, &req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "成员权限提升成功"})
}

// DemoteMember 降级成员权限
func (c *ChatManagementController) DemoteMember(ctx *gin.Context) {
	chatIDStr := ctx.Param("chat_id")
	userIDStr := ctx.Param("user_id")
	
	chatID, err := strconv.ParseUint(chatIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的群组ID"})
		return
	}

	targetUserID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	userID := ctx.GetUint("user_id")
	if err := c.permissionService.DemoteMember(ctx, userID, uint(chatID), uint(targetUserID)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "成员权限降级成功"})
}

// GetChatMembers 获取群组成员列表
func (c *ChatManagementController) GetChatMembers(ctx *gin.Context) {
	chatIDStr := ctx.Param("chat_id")
	chatID, err := strconv.ParseUint(chatIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的群组ID"})
		return
	}

	userID := ctx.GetUint("user_id")
	members, err := c.permissionService.GetChatMembers(ctx, uint(chatID), userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": members})
}

// 公告和规则管理相关接口

// CreateAnnouncement 创建群组公告
func (c *ChatManagementController) CreateAnnouncement(ctx *gin.Context) {
	var req service.CreateAnnouncementRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.GetUint("user_id")
	announcement, err := c.announcementService.CreateAnnouncement(ctx, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": announcement, "message": "公告创建成功"})
}

// UpdateAnnouncement 更新群组公告
func (c *ChatManagementController) UpdateAnnouncement(ctx *gin.Context) {
	var req service.UpdateAnnouncementRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.GetUint("user_id")
	if err := c.announcementService.UpdateAnnouncement(ctx, userID, &req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "公告更新成功"})
}

// DeleteAnnouncement 删除群组公告
func (c *ChatManagementController) DeleteAnnouncement(ctx *gin.Context) {
	announcementIDStr := ctx.Param("announcement_id")
	announcementID, err := strconv.ParseUint(announcementIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的公告ID"})
		return
	}

	userID := ctx.GetUint("user_id")
	if err := c.announcementService.DeleteAnnouncement(ctx, userID, uint(announcementID)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "公告删除成功"})
}

// GetChatAnnouncements 获取群组公告列表
func (c *ChatManagementController) GetChatAnnouncements(ctx *gin.Context) {
	chatIDStr := ctx.Param("chat_id")
	chatID, err := strconv.ParseUint(chatIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的群组ID"})
		return
	}

	userID := ctx.GetUint("user_id")
	announcements, err := c.announcementService.GetChatAnnouncements(ctx, uint(chatID), userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": announcements})
}

// GetPinnedAnnouncement 获取置顶公告
func (c *ChatManagementController) GetPinnedAnnouncement(ctx *gin.Context) {
	chatIDStr := ctx.Param("chat_id")
	chatID, err := strconv.ParseUint(chatIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的群组ID"})
		return
	}

	userID := ctx.GetUint("user_id")
	announcement, err := c.announcementService.GetPinnedAnnouncement(ctx, uint(chatID), userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": announcement})
}

// PinAnnouncement 置顶公告
func (c *ChatManagementController) PinAnnouncement(ctx *gin.Context) {
	announcementIDStr := ctx.Param("announcement_id")
	announcementID, err := strconv.ParseUint(announcementIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的公告ID"})
		return
	}

	userID := ctx.GetUint("user_id")
	if err := c.announcementService.PinAnnouncement(ctx, userID, uint(announcementID)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "公告置顶成功"})
}

// UnpinAnnouncement 取消置顶公告
func (c *ChatManagementController) UnpinAnnouncement(ctx *gin.Context) {
	announcementIDStr := ctx.Param("announcement_id")
	announcementID, err := strconv.ParseUint(announcementIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的公告ID"})
		return
	}

	userID := ctx.GetUint("user_id")
	if err := c.announcementService.UnpinAnnouncement(ctx, userID, uint(announcementID)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "取消置顶成功"})
}

// CreateRule 创建群组规则
func (c *ChatManagementController) CreateRule(ctx *gin.Context) {
	var req service.CreateRuleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.GetUint("user_id")
	rule, err := c.announcementService.CreateRule(ctx, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": rule, "message": "规则创建成功"})
}

// UpdateRule 更新群组规则
func (c *ChatManagementController) UpdateRule(ctx *gin.Context) {
	var req service.UpdateRuleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.GetUint("user_id")
	if err := c.announcementService.UpdateRule(ctx, userID, &req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "规则更新成功"})
}

// DeleteRule 删除群组规则
func (c *ChatManagementController) DeleteRule(ctx *gin.Context) {
	ruleIDStr := ctx.Param("rule_id")
	ruleID, err := strconv.ParseUint(ruleIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的规则ID"})
		return
	}

	userID := ctx.GetUint("user_id")
	if err := c.announcementService.DeleteRule(ctx, userID, uint(ruleID)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "规则删除成功"})
}

// GetChatRules 获取群组规则列表
func (c *ChatManagementController) GetChatRules(ctx *gin.Context) {
	chatIDStr := ctx.Param("chat_id")
	chatID, err := strconv.ParseUint(chatIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的群组ID"})
		return
	}

	userID := ctx.GetUint("user_id")
	rules, err := c.announcementService.GetChatRules(ctx, uint(chatID), userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": rules})
}

// 统计和分析相关接口

// GetChatStatistics 获取群组统计信息
func (c *ChatManagementController) GetChatStatistics(ctx *gin.Context) {
	chatIDStr := ctx.Param("chat_id")
	chatID, err := strconv.ParseUint(chatIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的群组ID"})
		return
	}

	var req service.StatisticsRequest
	req.ChatID = uint(chatID)

	// 从查询参数获取统计条件
	if dateFromStr := ctx.Query("date_from"); dateFromStr != "" {
		if dateFrom, err := time.Parse("2006-01-02", dateFromStr); err == nil {
			req.DateFrom = &dateFrom
		}
	}
	if dateToStr := ctx.Query("date_to"); dateToStr != "" {
		if dateTo, err := time.Parse("2006-01-02", dateToStr); err == nil {
			req.DateTo = &dateTo
		}
	}
	req.GroupBy = ctx.Query("group_by")

	userID := ctx.GetUint("user_id")
	statistics, err := c.statisticsService.GetChatStatistics(ctx, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": statistics})
}

// 备份和恢复相关接口

// CreateBackup 创建群组备份
func (c *ChatManagementController) CreateBackup(ctx *gin.Context) {
	var req service.CreateBackupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.GetUint("user_id")
	backup, err := c.backupService.CreateBackup(ctx, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": backup, "message": "备份创建成功"})
}

// RestoreBackup 恢复群组备份
func (c *ChatManagementController) RestoreBackup(ctx *gin.Context) {
	var req service.RestoreBackupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.GetUint("user_id")
	if err := c.backupService.RestoreBackup(ctx, userID, &req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "备份恢复成功"})
}

// GetBackupList 获取备份列表
func (c *ChatManagementController) GetBackupList(ctx *gin.Context) {
	chatIDStr := ctx.Param("chat_id")
	chatID, err := strconv.ParseUint(chatIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的群组ID"})
		return
	}

	userID := ctx.GetUint("user_id")
	backups, err := c.backupService.GetBackupList(ctx, uint(chatID), userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": backups})
}

// DeleteBackup 删除备份
func (c *ChatManagementController) DeleteBackup(ctx *gin.Context) {
	backupIDStr := ctx.Param("backup_id")
	backupID, err := strconv.ParseUint(backupIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的备份ID"})
		return
	}

	userID := ctx.GetUint("user_id")
	if err := c.backupService.DeleteBackup(ctx, userID, uint(backupID)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "备份删除成功"})
}
