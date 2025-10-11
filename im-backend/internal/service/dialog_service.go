package service

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"zhihang-messenger/im-backend/config"
	"zhihang-messenger/im-backend/internal/model"
)

// DialogService 会话服务
type DialogService struct {
	db *gorm.DB
}

// NewDialogService 创建会话服务
func NewDialogService() *DialogService {
	return &DialogService{
		db: config.DB,
	}
}

// DialogResponse 会话响应（对应Telegram的TL_dialog）
type DialogResponse struct {
	PeerID           uint   `json:"peer_id"`            // 对话方ID
	PeerType         string `json:"peer_type"`          // user/group/channel
	TopMessageID     uint   `json:"top_message_id"`     // 最新消息ID
	UnreadCount      int    `json:"unread_count"`       // 未读数
	Pinned           bool   `json:"pinned"`             // 是否置顶
	Muted            bool   `json:"muted"`              // 是否静音
	LastMessageDate  int64  `json:"last_message_date"`  // 最后消息时间（Unix时间戳）
	Draft            string `json:"draft,omitempty"`    // 草稿
	MuteUntil        int64  `json:"mute_until"`         // 静音到期时间
}

// DialogsResponse 会话列表响应（对应Telegram的TL_messages_dialogs）
type DialogsResponse struct {
	Dialogs  []DialogResponse   `json:"dialogs"`  // 会话列表
	Messages []model.Message    `json:"messages"` // 最新消息列表
	Users    []DialogUserInfo   `json:"users"`    // 用户信息列表
	Groups   []DialogGroupInfo  `json:"groups"`   // 群组信息列表
	Total    int                `json:"total"`    // 总会话数
}

// DialogUserInfo 会话中的用户信息
type DialogUserInfo struct {
	ID        uint   `json:"id"`
	Phone     string `json:"phone"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Online    bool   `json:"online"`
	LastSeen  int64  `json:"last_seen"` // Unix时间戳
}

// DialogGroupInfo 会话中的群组信息
type DialogGroupInfo struct {
	ID            uint   `json:"id"`
	Title         string `json:"title"`
	Photo         string `json:"photo"`
	MembersCount  int    `json:"members_count"`
	Type          string `json:"type"` // group/channel
}

// GetDialogs 获取会话列表
func (s *DialogService) GetDialogs(userID uint, limit, offset int) (*DialogsResponse, error) {
	response := &DialogsResponse{
		Dialogs:  []DialogResponse{},
		Messages: []model.Message{},
		Users:    []DialogUserInfo{},
		Groups:   []DialogGroupInfo{},
	}

	// 1. 获取该用户所有相关的消息（私聊 + 群聊）
	var messages []model.Message
	
	// 查询私聊消息（作为发送者或接收者）
	err := s.db.Where("sender_id = ? OR receiver_id = ?", userID, userID).
		Order("created_at DESC").
		Find(&messages).Error
	if err != nil {
		return nil, err
	}

	// 2. 按对话方分组，提取最新消息
	dialogMap := make(map[string]*DialogResponse) // key: "user_123" 或 "group_456"
	messageMap := make(map[uint]*model.Message)   // 存储最新消息
	userIDSet := make(map[uint]bool)              // 收集所有用户ID
	groupIDSet := make(map[uint]bool)             // 收集所有群组ID

	for i := range messages {
		msg := &messages[i]
		var peerID uint
		var peerType string
		var dialogKey string

		// 判断对话方
		if msg.ChatID != nil && *msg.ChatID > 0 {
			// 群聊
			peerID = *msg.ChatID
			peerType = "group"
			dialogKey = "group_" + string(rune(peerID))
			groupIDSet[peerID] = true
		} else if msg.ReceiverID != nil {
			// 私聊
			if msg.SenderID == userID {
				// 我发送的消息，对话方是接收者
				peerID = *msg.ReceiverID
			} else {
				// 我接收的消息，对话方是发送者
				peerID = msg.SenderID
			}
			peerType = "user"
			dialogKey = "user_" + string(rune(peerID))
			userIDSet[peerID] = true
		} else {
			continue // 跳过无效消息
		}

		// 如果这个对话还没有记录，或者当前消息更新，则更新
		if dialog, exists := dialogMap[dialogKey]; !exists {
			// 计算未读数
			unreadCount := s.getUnreadCount(userID, peerID, peerType)
			
			dialogMap[dialogKey] = &DialogResponse{
				PeerID:          peerID,
				PeerType:        peerType,
				TopMessageID:    msg.ID,
				UnreadCount:     unreadCount,
				LastMessageDate: msg.CreatedAt.Unix(),
				Pinned:          false, // TODO: 从用户设置表读取
				Muted:           false, // TODO: 从用户设置表读取
			}
			messageMap[msg.ID] = msg
		} else if msg.CreatedAt.After(messageMap[dialog.TopMessageID].CreatedAt) {
			// 更新为更新的消息
			dialog.TopMessageID = msg.ID
			dialog.LastMessageDate = msg.CreatedAt.Unix()
			messageMap[msg.ID] = msg
		}
	}

	// 3. 转换为切片并排序（按最后消息时间倒序）
	dialogs := make([]DialogResponse, 0, len(dialogMap))
	for _, dialog := range dialogMap {
		dialogs = append(dialogs, *dialog)
	}

	// 排序：置顶的在前，然后按时间倒序
	for i := 0; i < len(dialogs)-1; i++ {
		for j := i + 1; j < len(dialogs); j++ {
			if dialogs[i].Pinned == dialogs[j].Pinned {
				if dialogs[i].LastMessageDate < dialogs[j].LastMessageDate {
					dialogs[i], dialogs[j] = dialogs[j], dialogs[i]
				}
			} else if !dialogs[i].Pinned && dialogs[j].Pinned {
				dialogs[i], dialogs[j] = dialogs[j], dialogs[i]
			}
		}
	}

	// 4. 应用分页
	total := len(dialogs)
	if offset >= total {
		response.Total = total
		return response, nil
	}
	
	end := offset + limit
	if end > total {
		end = total
	}
	dialogs = dialogs[offset:end]

	// 5. 收集最新消息
	topMessages := make([]model.Message, 0, len(dialogs))
	for _, dialog := range dialogs {
		if msg, exists := messageMap[dialog.TopMessageID]; exists {
			topMessages = append(topMessages, *msg)
		}
	}

	// 6. 获取用户信息
	users := make([]DialogUserInfo, 0, len(userIDSet))
	if len(userIDSet) > 0 {
		userIDs := make([]uint, 0, len(userIDSet))
		for id := range userIDSet {
			if id != userID { // 不包含自己
				userIDs = append(userIDs, id)
			}
		}

		var dbUsers []model.User
		if err := s.db.Where("id IN ?", userIDs).Find(&dbUsers).Error; err == nil {
			for _, u := range dbUsers {
				users = append(users, DialogUserInfo{
					ID:       u.ID,
					Phone:    u.Phone,
					Username: u.Username,
					Nickname: u.Nickname,
					Avatar:   u.Avatar,
					Online:   u.Online,
					LastSeen: u.LastSeenAt.Unix(),
				})
			}
		}
	}

	// 7. 获取群组信息
	groups := make([]DialogGroupInfo, 0, len(groupIDSet))
	if len(groupIDSet) > 0 {
		groupIDs := make([]uint, 0, len(groupIDSet))
		for id := range groupIDSet {
			groupIDs = append(groupIDs, id)
		}

		var dbChats []model.Chat
		if err := s.db.Where("id IN ?", groupIDs).Find(&dbChats).Error; err == nil {
			for _, c := range dbChats {
				groups = append(groups, DialogGroupInfo{
					ID:           c.ID,
					Title:        c.Name,
					Photo:        c.Avatar,
					MembersCount: c.MembersCount,
					Type:         c.Type,
				})
			}
		}
	}

	// 8. 组装响应
	response.Dialogs = dialogs
	response.Messages = topMessages
	response.Users = users
	response.Groups = groups
	response.Total = total

	return response, nil
}

// GetDialogByPeer 获取与指定用户/群组的会话
func (s *DialogService) GetDialogByPeer(userID, peerID uint, peerType string) (*DialogResponse, error) {
	// 查找最新一条消息
	var message model.Message
	var err error

	if peerType == "user" {
		// 私聊：查找userID和peerID之间的最新消息
		err = s.db.Where(
			"(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			userID, peerID, peerID, userID,
		).Order("created_at DESC").First(&message).Error
	} else if peerType == "group" {
		// 群聊：查找chat_id的最新消息
		err = s.db.Where("chat_id = ?", peerID).
			Order("created_at DESC").First(&message).Error
	} else {
		return nil, errors.New("无效的peer_type")
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 没有消息历史，返回空会话
			return &DialogResponse{
				PeerID:          peerID,
				PeerType:        peerType,
				TopMessageID:    0,
				UnreadCount:     0,
				LastMessageDate: time.Now().Unix(),
			}, nil
		}
		return nil, err
	}

	// 计算未读数
	unreadCount := s.getUnreadCount(userID, peerID, peerType)

	return &DialogResponse{
		PeerID:          peerID,
		PeerType:        peerType,
		TopMessageID:    message.ID,
		UnreadCount:     unreadCount,
		LastMessageDate: message.CreatedAt.Unix(),
		Pinned:          false, // TODO: 从用户设置表读取
		Muted:           false,
	}, nil
}

// SetDialogPin 设置会话置顶
func (s *DialogService) SetDialogPin(userID, peerID uint, peerType string, pinned bool) error {
	// TODO: 需要创建一个user_dialog_settings表来存储置顶、静音等设置
	// 这里暂时只返回成功，实际数据库操作需要后续实现
	return nil
}

// SetDialogMute 设置会话静音
func (s *DialogService) SetDialogMute(userID, peerID uint, peerType string, muted bool, muteUntil int64) error {
	// TODO: 同上，需要user_dialog_settings表
	return nil
}

// getUnreadCount 计算未读消息数
func (s *DialogService) getUnreadCount(userID, peerID uint, peerType string) int {
	var count int64

	if peerType == "user" {
		// 私聊：对方发给我的、我还没读的消息
		s.db.Model(&model.Message{}).
			Where("sender_id = ? AND receiver_id = ? AND read_at IS NULL", peerID, userID).
			Count(&count)
	} else if peerType == "group" {
		// 群聊：群里的消息、不是我发的、我还没读的
		s.db.Model(&model.Message{}).
			Where("chat_id = ? AND sender_id != ? AND read_at IS NULL", peerID, userID).
			Count(&count)
	}

	return int(count)
}

