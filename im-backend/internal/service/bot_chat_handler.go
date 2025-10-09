package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"

	"zhihang-messenger/im-backend/config"
	"zhihang-messenger/im-backend/internal/model"
)

// BotChatHandler 机器人聊天处理器
type BotChatHandler struct {
	db             *gorm.DB
	botService     *BotService
	messageService *MessageService
}

// NewBotChatHandler 创建机器人聊天处理器
func NewBotChatHandler() *BotChatHandler {
	return &BotChatHandler{
		db:             config.DB,
		botService:     NewBotService(),
		messageService: NewMessageService(),
	}
}

// BotCommand 机器人命令
type BotCommand struct {
	Command string            // 命令名称
	Args    map[string]string // 命令参数
	UserID  uint              // 发起用户ID
	ChatID  uint              // 聊天ID
}

// HandleMessage 处理用户发送给机器人的消息
func (h *BotChatHandler) HandleMessage(ctx context.Context, message *model.Message) error {
	// 1. 检查消息接收者是否是机器人用户
	if message.ReceiverID == nil {
		return nil // 群聊消息，暂不处理
	}

	botUser, err := h.getBotUser(ctx, *message.ReceiverID)
	if err != nil {
		return nil // 不是发给机器人的消息，忽略
	}

	// 2. 检查发送者是否有权限使用机器人
	if !h.checkUserPermission(ctx, message.SenderID) {
		return h.sendReply(ctx, message, "❌ 您没有权限使用机器人功能。请联系管理员授权。")
	}

	// 3. 解析命令
	cmd, err := h.parseCommand(message.Content)
	if err != nil {
		return h.sendReply(ctx, message, fmt.Sprintf("❌ 命令格式错误: %s\n\n使用 /help 查看帮助", err.Error()))
	}

	cmd.UserID = message.SenderID
	if message.ChatID != nil {
		cmd.ChatID = *message.ChatID
	}

	// 4. 执行命令
	response := h.executeCommand(ctx, &botUser.BotID, cmd)

	// 5. 发送回复
	return h.sendReply(ctx, message, response)
}

// getBotUser 获取机器人用户信息
func (h *BotChatHandler) getBotUser(ctx context.Context, userID uint) (*model.BotUser, error) {
	var botUser model.BotUser
	err := h.db.WithContext(ctx).Where("user_id = ? AND is_active = ?", userID, true).First(&botUser).Error
	return &botUser, err
}

// checkUserPermission 检查用户是否有权限使用机器人
func (h *BotChatHandler) checkUserPermission(ctx context.Context, userID uint) bool {
	var permission model.BotUserPermission
	err := h.db.WithContext(ctx).
		Where("user_id = ? AND is_active = ?", userID, true).
		First(&permission).Error
	return err == nil
}

// parseCommand 解析命令
func (h *BotChatHandler) parseCommand(content string) (*BotCommand, error) {
	content = strings.TrimSpace(content)

	// 检查是否以 / 开头
	if !strings.HasPrefix(content, "/") {
		return nil, fmt.Errorf("命令必须以 / 开头")
	}

	// 分割命令和参数
	parts := strings.Fields(content)
	if len(parts) == 0 {
		return nil, fmt.Errorf("空命令")
	}

	cmd := &BotCommand{
		Command: strings.ToLower(parts[0]),
		Args:    make(map[string]string),
	}

	// 解析参数（格式: key=value 或 key:value）
	for i := 1; i < len(parts); i++ {
		part := parts[i]

		// 支持 key=value 格式
		if strings.Contains(part, "=") {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) == 2 {
				cmd.Args[kv[0]] = kv[1]
			}
		} else if strings.Contains(part, ":") {
			// 支持 key:value 格式
			kv := strings.SplitN(part, ":", 2)
			if len(kv) == 2 {
				cmd.Args[kv[0]] = kv[1]
			}
		}
	}

	return cmd, nil
}

// executeCommand 执行命令
func (h *BotChatHandler) executeCommand(ctx context.Context, botID *uint, cmd *BotCommand) string {
	if botID == nil {
		return "❌ 机器人配置错误"
	}

	// 获取机器人实例
	bot, err := h.botService.GetBotByID(ctx, *botID)
	if err != nil {
		return "❌ 机器人不存在"
	}

	switch cmd.Command {
	case "/help":
		return h.cmdHelp()
	case "/create", "/create_user":
		return h.cmdCreateUser(ctx, bot, cmd)
	case "/delete", "/delete_user":
		return h.cmdDeleteUser(ctx, bot, cmd)
	case "/list", "/list_users":
		return h.cmdListUsers(ctx, bot, cmd)
	case "/info", "/user_info":
		return h.cmdUserInfo(ctx, bot, cmd)
	default:
		return fmt.Sprintf("❌ 未知命令: %s\n\n使用 /help 查看帮助", cmd.Command)
	}
}

// cmdHelp 帮助命令
func (h *BotChatHandler) cmdHelp() string {
	return `🤖 **机器人命令帮助**

**用户管理命令：**

📝 **/create** - 创建新用户
格式: /create phone=手机号 username=用户名 password=密码 [nickname=昵称]
示例: /create phone=13800138000 username=testuser password=Pass123! nickname=测试用户

🗑️ **/delete** - 删除用户
格式: /delete user_id=用户ID reason=删除原因
示例: /delete user_id=123 reason=测试完成

📋 **/list** - 查看已创建的用户
格式: /list [limit=数量]
示例: /list limit=10

ℹ️ **/info** - 查看用户信息
格式: /info user_id=用户ID
示例: /info user_id=123

❓ **/help** - 显示此帮助信息

**注意事项：**
⚠️ 只能创建普通用户（role=user）
⚠️ 只能删除机器人创建的用户
⚠️ 所有操作都会被记录

**命令格式说明：**
- 参数使用 key=value 或 key:value 格式
- 多个参数用空格分隔
- 参数值如有空格请用引号包裹（暂不支持）`
}

// cmdCreateUser 创建用户命令
func (h *BotChatHandler) cmdCreateUser(ctx context.Context, bot *model.Bot, cmd *BotCommand) string {
	// 验证必填参数
	phone := cmd.Args["phone"]
	username := cmd.Args["username"]
	password := cmd.Args["password"]
	nickname := cmd.Args["nickname"]

	if phone == "" {
		return "❌ 缺少必填参数: phone\n示例: /create phone=13800138000 username=testuser password=Pass123!"
	}
	if username == "" {
		return "❌ 缺少必填参数: username\n示例: /create phone=13800138000 username=testuser password=Pass123!"
	}
	if password == "" {
		return "❌ 缺少必填参数: password\n示例: /create phone=13800138000 username=testuser password=Pass123!"
	}

	// 验证手机号格式
	phoneRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)
	if !phoneRegex.MatchString(phone) {
		return "❌ 手机号格式错误，应为11位数字，以1开头"
	}

	// 验证密码强度
	if len(password) < 6 {
		return "❌ 密码长度至少6位"
	}

	// 调用Bot服务创建用户
	req := &BotCreateUserRequest{
		Phone:    phone,
		Username: username,
		Password: password,
		Nickname: nickname,
	}

	user, err := h.botService.BotCreateUser(ctx, bot, req)
	if err != nil {
		return fmt.Sprintf("❌ 创建用户失败: %s", err.Error())
	}

	return fmt.Sprintf(`✅ **用户创建成功！**

👤 **用户信息：**
- ID: %d
- 用户名: %s
- 手机号: %s
- 昵称: %s
- 角色: %s
- 状态: %s
- 创建时间: %s

⚠️ 请妥善保管用户的登录凭证`,
		user.ID,
		user.Username,
		user.Phone,
		user.Nickname,
		user.Role,
		func() string {
			if user.IsActive {
				return "激活"
			}
			return "未激活"
		}(),
		user.CreatedAt.Format("2006-01-02 15:04:05"),
	)
}

// cmdDeleteUser 删除用户命令
func (h *BotChatHandler) cmdDeleteUser(ctx context.Context, bot *model.Bot, cmd *BotCommand) string {
	// 验证必填参数
	userIDStr := cmd.Args["user_id"]
	reason := cmd.Args["reason"]

	if userIDStr == "" {
		return "❌ 缺少必填参数: user_id\n示例: /delete user_id=123 reason=测试完成"
	}
	if reason == "" {
		return "❌ 缺少必填参数: reason\n示例: /delete user_id=123 reason=测试完成"
	}

	// 解析用户ID
	var userID uint
	if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
		return "❌ 用户ID格式错误，应为数字"
	}

	// 获取用户信息（用于显示）
	var user model.User
	if err := h.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return "❌ 用户不存在"
	}

	// 调用Bot服务删除用户
	req := &BotDeleteUserRequest{
		UserID: userID,
		Reason: reason,
	}

	if err := h.botService.BotDeleteUser(ctx, bot, req); err != nil {
		return fmt.Sprintf("❌ 删除用户失败: %s", err.Error())
	}

	return fmt.Sprintf(`✅ **用户删除成功！**

👤 **已删除用户：**
- ID: %d
- 用户名: %s
- 手机号: %s
- 删除原因: %s
- 删除时间: %s`,
		user.ID,
		user.Username,
		user.Phone,
		reason,
		time.Now().Format("2006-01-02 15:04:05"),
	)
}

// cmdListUsers 列出用户命令
func (h *BotChatHandler) cmdListUsers(ctx context.Context, bot *model.Bot, cmd *BotCommand) string {
	// 解析limit参数
	limit := 10
	if limitStr, ok := cmd.Args["limit"]; ok {
		fmt.Sscanf(limitStr, "%d", &limit)
	}
	if limit > 50 {
		limit = 50 // 最多显示50个
	}

	// 查询机器人创建的用户
	var users []model.User
	err := h.db.WithContext(ctx).
		Where("created_by_bot_id = ? AND deleted_at IS NULL", bot.ID).
		Order("created_at DESC").
		Limit(limit).
		Find(&users).Error

	if err != nil {
		return "❌ 查询用户列表失败"
	}

	if len(users) == 0 {
		return "📋 暂无用户\n\n使用 /create 创建新用户"
	}

	// 构建列表
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("📋 **用户列表** (共 %d 个)\n\n", len(users)))

	for i, user := range users {
		builder.WriteString(fmt.Sprintf("%d. **%s** (ID:%d)\n", i+1, user.Username, user.ID))
		builder.WriteString(fmt.Sprintf("   - 手机: %s\n", user.Phone))
		builder.WriteString(fmt.Sprintf("   - 昵称: %s\n", user.Nickname))
		builder.WriteString(fmt.Sprintf("   - 状态: %s\n", func() string {
			if user.IsActive {
				return "✅ 激活"
			}
			return "❌ 未激活"
		}()))
		builder.WriteString(fmt.Sprintf("   - 创建: %s\n\n", user.CreatedAt.Format("2006-01-02 15:04")))
	}

	builder.WriteString("💡 使用 /info user_id=ID 查看详细信息")

	return builder.String()
}

// cmdUserInfo 用户信息命令
func (h *BotChatHandler) cmdUserInfo(ctx context.Context, bot *model.Bot, cmd *BotCommand) string {
	// 验证必填参数
	userIDStr := cmd.Args["user_id"]
	if userIDStr == "" {
		return "❌ 缺少必填参数: user_id\n示例: /info user_id=123"
	}

	// 解析用户ID
	var userID uint
	if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
		return "❌ 用户ID格式错误，应为数字"
	}

	// 查询用户
	var user model.User
	if err := h.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return "❌ 用户不存在"
	}

	// 检查是否是本机器人创建的
	isManagedByBot := user.CreatedByBotID != nil && *user.CreatedByBotID == bot.ID

	return fmt.Sprintf(`ℹ️ **用户详细信息**

**基本信息：**
- ID: %d
- 用户名: %s
- 手机号: %s
- 昵称: %s
- 角色: %s

**状态信息：**
- 账户状态: %s
- 在线状态: %s
- 最后在线: %s

**管理信息：**
- 由本机器人创建: %s
- 可被机器人管理: %s
- 注册时间: %s

**操作提示：**
%s`,
		user.ID,
		user.Username,
		user.Phone,
		user.Nickname,
		user.Role,
		func() string {
			if user.IsActive {
				return "✅ 激活"
			}
			return "❌ 未激活"
		}(),
		func() string {
			if user.Online {
				return "🟢 在线"
			}
			return "⚪ 离线"
		}(),
		user.LastSeen.Format("2006-01-02 15:04:05"),
		func() string {
			if isManagedByBot {
				return "✅ 是"
			}
			return "❌ 否"
		}(),
		func() string {
			if user.BotManageable {
				return "✅ 是"
			}
			return "❌ 否"
		}(),
		user.CreatedAt.Format("2006-01-02 15:04:05"),
		func() string {
			if isManagedByBot {
				return "💡 您可以删除此用户: /delete user_id=" + userIDStr + " reason=删除原因"
			}
			return "⚠️ 此用户不是由本机器人创建，无法删除"
		}(),
	)
}

// sendReply 发送回复消息
func (h *BotChatHandler) sendReply(ctx context.Context, originalMessage *model.Message, content string) error {
	reply := &model.Message{
		ChatID:      originalMessage.ChatID,
		SenderID:    *originalMessage.ReceiverID, // 机器人作为发送者
		ReceiverID:  &originalMessage.SenderID,   // 原发送者作为接收者
		MessageType: "text",
		Content:     content,
		Status:      "sent",
	}

	return h.db.WithContext(ctx).Create(reply).Error
}
