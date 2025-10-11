package model

import (
	"gorm.io/gorm"
	"time"
)

// File 文件模型
type File struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// 基本信息
	FileName string `json:"file_name" gorm:"type:varchar(255);not null"`                            // 原始文件名
	FileSize int64  `json:"file_size" gorm:"not null"`                                              // 文件大小
	FileType string `json:"file_type" gorm:"type:varchar(50);not null"`                             // 文件类型
	MimeType string `json:"mime_type" gorm:"type:varchar(100);not null"`                            // MIME类型
	FileHash string `json:"file_hash" gorm:"type:varchar(64);index:idx_files_hash,unique;not null"` // 文件哈希值 (SHA256)

	// 存储信息
	StoragePath string `json:"storage_path" gorm:"type:varchar(500);not null"` // 存储路径
	StorageURL  string `json:"storage_url" gorm:"type:varchar(500);not null"`  // 访问URL
	Thumbnail   string `json:"thumbnail" gorm:"type:varchar(500)"`             // 缩略图URL
	Preview     string `json:"preview" gorm:"type:varchar(500)"`               // 预览URL

	// 加密信息
	IsEncrypted   bool   `json:"is_encrypted" gorm:"default:false"`       // 是否加密
	EncryptionKey string `json:"encryption_key" gorm:"type:varchar(255)"` // 加密密钥(加密存储)

	// 版本控制
	Version  int   `json:"version" gorm:"default:1"`      // 文件版本
	ParentID *uint `json:"parent_id"`                     // 父文件ID(用于版本控制)
	IsLatest bool  `json:"is_latest" gorm:"default:true"` // 是否最新版本

	// 分片上传信息
	ChunkSize   int64  `json:"chunk_size" gorm:"default:0"`        // 分片大小
	TotalChunks int    `json:"total_chunks" gorm:"default:0"`      // 总分片数
	UploadID    string `json:"upload_id" gorm:"type:varchar(100)"` // 上传ID

	// 权限信息
	OwnerID     uint   `json:"owner_id" gorm:"not null"`              // 文件所有者
	IsPublic    bool   `json:"is_public" gorm:"default:false"`        // 是否公开
	AccessToken string `json:"access_token" gorm:"type:varchar(255)"` // 访问令牌

	// 统计信息
	DownloadCount int64 `json:"download_count" gorm:"default:0"` // 下载次数
	ViewCount     int64 `json:"view_count" gorm:"default:0"`     // 查看次数

	// 关联关系
	Owner    User   `json:"owner" gorm:"foreignKey:OwnerID"`
	Parent   *File  `json:"parent" gorm:"foreignKey:ParentID"`
	Versions []File `json:"versions" gorm:"foreignKey:ParentID"`
}

// FileChunk 文件分片模型
type FileChunk struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// 分片信息
	UploadID   string `json:"upload_id" gorm:"not null"`   // 上传ID
	ChunkIndex int    `json:"chunk_index" gorm:"not null"` // 分片索引
	ChunkSize  int64  `json:"chunk_size" gorm:"not null"`  // 分片大小
	ChunkHash  string `json:"chunk_hash" gorm:"not null"`  // 分片哈希值

	// 存储信息
	StoragePath string `json:"storage_path" gorm:"not null"`     // 存储路径
	IsUploaded  bool   `json:"is_uploaded" gorm:"default:false"` // 是否已上传
	IsVerified  bool   `json:"is_verified" gorm:"default:false"` // 是否已验证

	// 关联关系
	File *File `json:"file" gorm:"foreignKey:UploadID;references:UploadID"`
}

// FilePreview 文件预览模型
type FilePreview struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// 预览信息
	FileID       uint   `json:"file_id" gorm:"not null"`      // 文件ID
	PreviewType  string `json:"preview_type" gorm:"not null"` // 预览类型: image/video/audio/document
	PreviewURL   string `json:"preview_url" gorm:"not null"`  // 预览URL
	ThumbnailURL string `json:"thumbnail_url"`                // 缩略图URL

	// 预览配置
	Width    int    `json:"width"`    // 宽度
	Height   int    `json:"height"`   // 高度
	Duration int    `json:"duration"` // 时长(秒)
	Quality  string `json:"quality"`  // 质量

	// 关联关系
	File File `json:"file" gorm:"foreignKey:FileID"`
}

// FileAccess 文件访问记录模型
type FileAccess struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// 访问信息
	FileID     uint   `json:"file_id" gorm:"not null"`     // 文件ID
	UserID     *uint  `json:"user_id"`                     // 用户ID(可为空，表示匿名访问)
	AccessType string `json:"access_type" gorm:"not null"` // 访问类型: view/download/preview
	IPAddress  string `json:"ip_address"`                  // IP地址
	UserAgent  string `json:"user_agent"`                  // 用户代理

	// 关联关系
	File File  `json:"file" gorm:"foreignKey:FileID"`
	User *User `json:"user" gorm:"foreignKey:UserID"`
}
