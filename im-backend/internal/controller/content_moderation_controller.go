package controller

import (
	"net/http"
	"strconv"
	"time"
	"zhihang-messenger/im-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// ContentModerationController 内容审核控制器
type ContentModerationController struct {
	moderationService *service.ContentModerationService
}

// NewContentModerationController 创建内容审核控制器实例
func NewContentModerationController(moderationService *service.ContentModerationService) *ContentModerationController {
	return &ContentModerationController{
		moderationService: moderationService,
	}
}

// ReportContent 举报内容
// @Summary 举报内容
// @Description 用户举报违规内容
// @Tags 内容审核
// @Accept json
// @Produce json
// @Param request body service.ReportContentRequest true "举报请求"
// @Success 200 {object} model.ContentReport
// @Failure 400 {object} map[string]interface{}
// @Router /api/moderation/report [post]
func (c *ContentModerationController) ReportContent(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req service.ReportContentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	req.ReporterID = userID.(uint)

	report, err := c.moderationService.ReportContent(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "举报已提交，我们会尽快处理",
		"report":  report,
	})
}

// CheckContent 检查内容（自动检测）
// @Summary 检查内容
// @Description 自动检测内容是否违规（仅上报，不拦截）
// @Tags 内容审核
// @Accept json
// @Produce json
// @Param request body service.CheckContentRequest true "检查请求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/moderation/check [post]
func (c *ContentModerationController) CheckContent(ctx *gin.Context) {
	var req service.CheckContentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	report, err := c.moderationService.CheckContent(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if report != nil {
		// 检测到违规内容，已自动上报
		ctx.JSON(http.StatusOK, gin.H{
			"detected": true,
			"message":  "检测到可疑内容，已自动上报至管理员",
			"report":   report,
		})
	} else {
		// 未检测到违规
		ctx.JSON(http.StatusOK, gin.H{
			"detected": false,
			"message":  "内容检测通过",
		})
	}
}

// GetPendingReports 获取待处理举报列表
// @Summary 获取待处理举报
// @Description 管理员获取待处理的举报列表
// @Tags 内容审核
// @Produce json
// @Param limit query int false "每页数量" default(20)
// @Param offset query int false "偏移量" default(0)
// @Param priority query string false "优先级筛选"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/moderation/reports/pending [get]
func (c *ContentModerationController) GetPendingReports(ctx *gin.Context) {
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))
	priority := ctx.Query("priority")

	reports, total, err := c.moderationService.GetPendingReports(limit, offset, priority)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"reports": reports,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// GetReportDetail 获取举报详情
// @Summary 获取举报详情
// @Description 管理员查看举报的详细信息
// @Tags 内容审核
// @Produce json
// @Param report_id path int true "举报ID"
// @Success 200 {object} model.ContentReport
// @Failure 400 {object} map[string]interface{}
// @Router /api/moderation/reports/{report_id} [get]
func (c *ContentModerationController) GetReportDetail(ctx *gin.Context) {
	reportIDStr := ctx.Param("report_id")
	reportID, err := strconv.ParseUint(reportIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "举报ID格式错误"})
		return
	}

	report, err := c.moderationService.GetReportDetail(uint(reportID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, report)
}

// HandleReport 处理举报
// @Summary 处理举报
// @Description 管理员处理举报内容
// @Tags 内容审核
// @Accept json
// @Produce json
// @Param request body service.HandleReportRequest true "处理请求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/moderation/reports/handle [post]
func (c *ContentModerationController) HandleReport(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req service.HandleReportRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	req.HandlerID = userID.(uint)

	if err := c.moderationService.HandleReport(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "举报处理成功",
	})
}

// CreateFilter 创建过滤规则
// @Summary 创建过滤规则
// @Description 管理员创建内容过滤规则
// @Tags 内容审核
// @Accept json
// @Produce json
// @Param request body service.CreateFilterRequest true "规则请求"
// @Success 200 {object} model.ContentFilter
// @Failure 400 {object} map[string]interface{}
// @Router /api/moderation/filters [post]
func (c *ContentModerationController) CreateFilter(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req service.CreateFilterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	req.CreatorID = userID.(uint)

	filter, err := c.moderationService.CreateFilter(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "过滤规则创建成功",
		"filter":  filter,
	})
}

// GetUserWarnings 获取用户警告记录
// @Summary 获取用户警告
// @Description 查看用户的警告记录
// @Tags 内容审核
// @Produce json
// @Param user_id path int true "用户ID"
// @Param limit query int false "每页数量" default(20)
// @Param offset query int false "偏移量" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/moderation/users/{user_id}/warnings [get]
func (c *ContentModerationController) GetUserWarnings(ctx *gin.Context) {
	userIDStr := ctx.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "用户ID格式错误"})
		return
	}

	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	warnings, total, err := c.moderationService.GetUserWarnings(uint(userID), limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"warnings": warnings,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

// GetStatistics 获取审核统计
// @Summary 获取审核统计
// @Description 获取内容审核的统计数据
// @Tags 内容审核
// @Produce json
// @Param start_date query string true "开始日期" format(2006-01-02)
// @Param end_date query string true "结束日期" format(2006-01-02)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/moderation/statistics [get]
func (c *ContentModerationController) GetStatistics(ctx *gin.Context) {
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请提供开始日期和结束日期"})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "开始日期格式错误"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "结束日期格式错误"})
		return
	}

	stats, err := c.moderationService.GetStatistics(startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"statistics": stats,
		"start_date": startDateStr,
		"end_date":   endDateStr,
	})
}
