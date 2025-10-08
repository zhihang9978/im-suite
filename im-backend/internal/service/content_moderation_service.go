package service

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"zhihang-messenger/im-backend/internal/model"

	"gorm.io/gorm"
)

// ContentModerationService 内容审核服务
type ContentModerationService struct {
	db *gorm.DB
}

// NewContentModerationService 创建内容审核服务实例
func NewContentModerationService(db *gorm.DB) *ContentModerationService {
	return &ContentModerationService{
		db: db,
	}
}

// ReportContentRequest 举报内容请求
type ReportContentRequest struct {
	ContentType   string `json:"content_type" binding:"required"` // message, user, chat, file
	ContentID     uint   `json:"content_id" binding:"required"`
	ContentUserID uint   `json:"content_user_id" binding:"required"`
	ReporterID    uint   `json:"reporter_id" binding:"required"`
	ReportReason  string `json:"report_reason" binding:"required"`
	ReportDetail  string `json:"report_detail"`
	ReportEvidence string `json:"report_evidence"`
}

// CheckContentRequest 检查内容请求
type CheckContentRequest struct {
	ContentType string `json:"content_type" binding:"required"`
	ContentID   uint   `json:"content_id" binding:"required"`
	ContentText string `json:"content_text" binding:"required"`
	UserID      uint   `json:"user_id" binding:"required"`
}

// HandleReportRequest 处理举报请求
type HandleReportRequest struct {
	ReportID      uint   `json:"report_id" binding:"required"`
	HandlerID     uint   `json:"handler_id" binding:"required"`
	HandleAction  string `json:"handle_action" binding:"required"` // warn, delete, ban, ignore
	HandleComment string `json:"handle_comment"`
}

// CreateFilterRequest 创建过滤规则请求
type CreateFilterRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	RuleType    string  `json:"rule_type" binding:"required"` // keyword, regex, url
	RuleContent string  `json:"rule_content" binding:"required"`
	Category    string  `json:"category" binding:"required"` // spam, porn, violence, politics
	Severity    string  `json:"severity" default:"normal"`
	Action      string  `json:"action" default:"report"`
	Threshold   float64 `json:"threshold" default:"0.8"`
	AutoReport  bool    `json:"auto_report" default:"true"`
	CreatorID   uint    `json:"creator_id" binding:"required"`
}

// ReportContent 举报内容
func (s *ContentModerationService) ReportContent(req ReportContentRequest) (*model.ContentReport, error) {
	// 验证举报原因
	validReasons := []string{"spam", "porn", "violence", "politics", "harassment", "fraud", "other"}
	if !contains(validReasons, req.ReportReason) {
		return nil, errors.New("无效的举报原因")
	}

	// 获取内容文本
	contentText := ""
	if req.ContentType == "message" {
		var message model.Message
		if err := s.db.First(&message, req.ContentID).Error; err == nil {
			contentText = message.Content
		}
	}

	// 创建举报记录
	report := model.ContentReport{
		ContentType:   req.ContentType,
		ContentID:     req.ContentID,
		ContentText:   contentText,
		ContentUserID: req.ContentUserID,
		ReporterID:    req.ReporterID,
		ReportReason:  req.ReportReason,
		ReportDetail:  req.ReportDetail,
		ReportEvidence: req.ReportEvidence,
		Status:        "pending",
		Priority:      s.calculatePriority(req.ReportReason),
		AutoDetected:  false,
	}

	if err := s.db.Create(&report).Error; err != nil {
		return nil, fmt.Errorf("创建举报记录失败: %v", err)
	}

	// 记录审核日志
	s.logModeration("report_created", req.ContentType, req.ContentID, req.ReporterID, 
		fmt.Sprintf("举报原因: %s", req.ReportReason), "")

	// 更新统计
	s.updateStatistics("manual_reported")

	return &report, nil
}

// CheckContent 自动检测内容（不拦截，仅上报）
func (s *ContentModerationService) CheckContent(req CheckContentRequest) (*model.ContentReport, error) {
	// 获取所有启用的过滤规则
	var filters []model.ContentFilter
	if err := s.db.Where("is_enabled = ?", true).Find(&filters).Error; err != nil {
		return nil, fmt.Errorf("获取过滤规则失败: %v", err)
	}

	// 检测内容
	for _, filter := range filters {
		matched, score, keywords := s.matchFilter(filter, req.ContentText)
		
		if matched && score >= filter.Threshold {
			// 更新过滤规则统计
			now := time.Now()
			s.db.Model(&filter).Updates(map[string]interface{}{
				"hit_count":     gorm.Expr("hit_count + 1"),
				"last_hit_time": now,
			})

			// 如果启用了自动上报，创建举报记录
			if filter.AutoReport {
				report := model.ContentReport{
					ContentType:       req.ContentType,
					ContentID:         req.ContentID,
					ContentText:       req.ContentText,
					ContentUserID:     req.UserID,
					ReporterID:        0, // 系统自动检测
					ReportReason:      filter.Category,
					ReportDetail:      fmt.Sprintf("自动检测匹配规则: %s", filter.Name),
					AutoDetected:      true,
					DetectionType:     filter.RuleType,
					DetectionScore:    score,
					DetectionKeywords: keywords,
					Status:            "pending",
					Priority:          s.mapSeverityToPriority(filter.Severity),
				}

				if err := s.db.Create(&report).Error; err != nil {
					return nil, fmt.Errorf("创建自动检测举报失败: %v", err)
				}

				// 记录审核日志
				s.logModeration("auto_detected", req.ContentType, req.ContentID, 0,
					fmt.Sprintf("自动检测匹配规则: %s, 分数: %.2f", filter.Name, score), "")

				// 更新统计
				s.updateStatistics("auto_detected")

				return &report, nil
			}
		}
	}

	return nil, nil // 未检测到违规内容
}

// HandleReport 处理举报（管理员操作）
func (s *ContentModerationService) HandleReport(req HandleReportRequest) error {
	// 验证处理动作
	validActions := []string{"warn", "delete", "ban", "ignore"}
	if !contains(validActions, req.HandleAction) {
		return errors.New("无效的处理动作")
	}

	// 获取举报记录
	var report model.ContentReport
	if err := s.db.First(&report, req.ReportID).Error; err != nil {
		return errors.New("举报记录不存在")
	}

	if report.Status != "pending" && report.Status != "reviewing" {
		return errors.New("该举报已处理")
	}

	// 更新举报状态
	now := time.Now()
	updates := map[string]interface{}{
		"status":         "resolved",
		"handler_id":     req.HandlerID,
		"handle_time":    now,
		"handle_action":  req.HandleAction,
		"handle_comment": req.HandleComment,
	}

	if err := s.db.Model(&report).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新举报状态失败: %v", err)
	}

	// 执行处理动作（仅记录，不实际拦截）
	switch req.HandleAction {
	case "warn":
		// 发出警告
		warning := model.UserWarning{
			UserID:       report.ContentUserID,
			WarningType:  report.ReportReason,
			WarningLevel: s.mapPriorityToLevel(report.Priority),
			Reason:       req.HandleComment,
			Evidence:     report.ContentText,
			ReportID:     &report.ID,
			IssuedBy:     req.HandlerID,
		}
		s.db.Create(&warning)
		s.updateStatistics("warnings_issued")

	case "delete":
		// 记录删除动作（实际删除由管理员手动执行）
		s.logModeration("content_marked_delete", report.ContentType, report.ContentID, 
			req.HandlerID, req.HandleComment, "")
		s.updateStatistics("content_deleted")

	case "ban":
		// 记录封禁动作（实际封禁由管理员手动执行）
		s.logModeration("user_marked_ban", "user", report.ContentUserID, 
			req.HandlerID, req.HandleComment, "")
		s.updateStatistics("users_banned")

	case "ignore":
		// 更新状态为已拒绝
		s.db.Model(&report).Update("status", "rejected")
		s.updateStatistics("rejected_reports")
	}

	// 记录处理日志
	s.logModeration("report_handled", report.ContentType, report.ContentID, 
		req.HandlerID, fmt.Sprintf("处理动作: %s", req.HandleAction), req.HandleComment)

	// 更新统计
	s.updateStatistics("resolved_reports")

	return nil
}

// CreateFilter 创建过滤规则
func (s *ContentModerationService) CreateFilter(req CreateFilterRequest) (*model.ContentFilter, error) {
	// 验证规则类型
	validTypes := []string{"keyword", "regex", "url"}
	if !contains(validTypes, req.RuleType) {
		return nil, errors.New("无效的规则类型")
	}

	// 验证正则表达式
	if req.RuleType == "regex" {
		if _, err := regexp.Compile(req.RuleContent); err != nil {
			return nil, fmt.Errorf("无效的正则表达式: %v", err)
		}
	}

	filter := model.ContentFilter{
		Name:        req.Name,
		Description: req.Description,
		RuleType:    req.RuleType,
		RuleContent: req.RuleContent,
		Category:    req.Category,
		Severity:    req.Severity,
		IsEnabled:   true,
		Action:      req.Action,
		Threshold:   req.Threshold,
		AutoReport:  req.AutoReport,
		CreatorID:   req.CreatorID,
	}

	if err := s.db.Create(&filter).Error; err != nil {
		return nil, fmt.Errorf("创建过滤规则失败: %v", err)
	}

	return &filter, nil
}

// GetPendingReports 获取待处理举报列表
func (s *ContentModerationService) GetPendingReports(limit, offset int, priority string) ([]model.ContentReport, int64, error) {
	var reports []model.ContentReport
	var total int64

	query := s.db.Preload("Reporter").Preload("ContentUser").Where("status IN ?", []string{"pending", "reviewing"})

	if priority != "" {
		query = query.Where("priority = ?", priority)
	}

	// 获取总数
	query.Model(&model.ContentReport{}).Count(&total)

	// 获取分页数据
	if err := query.Order("priority DESC, created_at DESC").
		Limit(limit).Offset(offset).Find(&reports).Error; err != nil {
		return nil, 0, fmt.Errorf("获取待处理举报失败: %v", err)
	}

	return reports, total, nil
}

// GetReportDetail 获取举报详情
func (s *ContentModerationService) GetReportDetail(reportID uint) (*model.ContentReport, error) {
	var report model.ContentReport
	if err := s.db.Preload("Reporter").Preload("ContentUser").Preload("Handler").
		First(&report, reportID).Error; err != nil {
		return nil, errors.New("举报记录不存在")
	}
	return &report, nil
}

// GetUserWarnings 获取用户警告记录
func (s *ContentModerationService) GetUserWarnings(userID uint, limit, offset int) ([]model.UserWarning, int64, error) {
	var warnings []model.UserWarning
	var total int64

	query := s.db.Preload("IssuedByUser").Preload("Report").Where("user_id = ?", userID)

	// 获取总数
	query.Model(&model.UserWarning{}).Count(&total)

	// 获取分页数据
	if err := query.Order("created_at DESC").
		Limit(limit).Offset(offset).Find(&warnings).Error; err != nil {
		return nil, 0, fmt.Errorf("获取用户警告失败: %v", err)
	}

	return warnings, total, nil
}

// GetStatistics 获取审核统计
func (s *ContentModerationService) GetStatistics(startDate, endDate time.Time) ([]model.ContentStatistics, error) {
	var stats []model.ContentStatistics
	if err := s.db.Where("date BETWEEN ? AND ?", startDate, endDate).
		Order("date DESC").Find(&stats).Error; err != nil {
		return nil, fmt.Errorf("获取统计数据失败: %v", err)
	}
	return stats, nil
}

// matchFilter 匹配过滤规则
func (s *ContentModerationService) matchFilter(filter model.ContentFilter, content string) (bool, float64, string) {
	content = strings.ToLower(content)

	switch filter.RuleType {
	case "keyword":
		keywords := strings.Split(filter.RuleContent, ",")
		matchedKeywords := []string{}
		matchCount := 0

		for _, keyword := range keywords {
			keyword = strings.TrimSpace(strings.ToLower(keyword))
			if strings.Contains(content, keyword) {
				matchCount++
				matchedKeywords = append(matchedKeywords, keyword)
			}
		}

		if matchCount > 0 {
			score := float64(matchCount) / float64(len(keywords))
			return true, score, strings.Join(matchedKeywords, ", ")
		}

	case "regex":
		re, err := regexp.Compile(filter.RuleContent)
		if err == nil && re.MatchString(content) {
			matches := re.FindAllString(content, -1)
			return true, 1.0, strings.Join(matches, ", ")
		}

	case "url":
		urlPattern := `https?://[^\s]+`
		re := regexp.MustCompile(urlPattern)
		if re.MatchString(content) {
			urls := re.FindAllString(content, -1)
			return true, 1.0, strings.Join(urls, ", ")
		}
	}

	return false, 0, ""
}

// calculatePriority 计算优先级
func (s *ContentModerationService) calculatePriority(reason string) string {
	highPriorityReasons := []string{"porn", "violence", "terrorism"}
	urgentPriorityReasons := []string{"child_abuse", "suicide"}

	for _, r := range urgentPriorityReasons {
		if reason == r {
			return "urgent"
		}
	}

	for _, r := range highPriorityReasons {
		if reason == r {
			return "high"
		}
	}

	return "normal"
}

// mapSeverityToPriority 严重程度映射到优先级
func (s *ContentModerationService) mapSeverityToPriority(severity string) string {
	severityMap := map[string]string{
		"critical": "urgent",
		"high":     "high",
		"normal":   "normal",
		"low":      "low",
	}
	if priority, ok := severityMap[severity]; ok {
		return priority
	}
	return "normal"
}

// mapPriorityToLevel 优先级映射到警告级别
func (s *ContentModerationService) mapPriorityToLevel(priority string) string {
	priorityMap := map[string]string{
		"urgent": "severe",
		"high":   "high",
		"normal": "medium",
		"low":    "low",
	}
	if level, ok := priorityMap[priority]; ok {
		return level
	}
	return "medium"
}

// logModeration 记录审核日志
func (s *ContentModerationService) logModeration(action, targetType string, targetID, operatorID uint, reason, details string) {
	log := model.ModerationLog{
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		OperatorID: operatorID,
		Reason:     reason,
		Details:    details,
	}
	s.db.Create(&log)
}

// updateStatistics 更新统计数据
func (s *ContentModerationService) updateStatistics(statType string) {
	today := time.Now().Format("2006-01-02")
	var stat model.ContentStatistics

	// 查找或创建今天的统计记录
	if err := s.db.Where("date = ?", today).First(&stat).Error; err != nil {
		stat = model.ContentStatistics{
			Date: time.Now(),
		}
		s.db.Create(&stat)
	}

	// 更新对应的统计字段
	updates := make(map[string]interface{})
	switch statType {
	case "manual_reported":
		updates["total_reports"] = gorm.Expr("total_reports + 1")
		updates["manual_reported"] = gorm.Expr("manual_reported + 1")
		updates["pending_reports"] = gorm.Expr("pending_reports + 1")
	case "auto_detected":
		updates["total_reports"] = gorm.Expr("total_reports + 1")
		updates["auto_detected"] = gorm.Expr("auto_detected + 1")
		updates["pending_reports"] = gorm.Expr("pending_reports + 1")
	case "resolved_reports":
		updates["resolved_reports"] = gorm.Expr("resolved_reports + 1")
		updates["pending_reports"] = gorm.Expr("pending_reports - 1")
	case "rejected_reports":
		updates["rejected_reports"] = gorm.Expr("rejected_reports + 1")
		updates["pending_reports"] = gorm.Expr("pending_reports - 1")
	case "warnings_issued":
		updates["warnings_issued"] = gorm.Expr("warnings_issued + 1")
	case "content_deleted":
		updates["content_deleted"] = gorm.Expr("content_deleted + 1")
	case "users_banned":
		updates["users_banned"] = gorm.Expr("users_banned + 1")
	}

	if len(updates) > 0 {
		s.db.Model(&model.ContentStatistics{}).Where("date = ?", today).Updates(updates)
	}
}
