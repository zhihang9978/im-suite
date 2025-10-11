package service

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"zhihang-messenger/im-backend/config"
	"zhihang-messenger/im-backend/internal/model"

	"gorm.io/gorm"
)

// FileService 文件服务
type FileService struct {
	db *gorm.DB
}

// NewFileService 创建文件服务实例
func NewFileService() *FileService {
	return &FileService{
		db: config.DB,
	}
}

// UploadRequest 文件上传请求
type UploadRequest struct {
	FileName    string `json:"file_name" binding:"required"`
	FileSize    int64  `json:"file_size" binding:"required"`
	FileType    string `json:"file_type" binding:"required"`
	MimeType    string `json:"mime_type" binding:"required"`
	FileData    []byte `json:"file_data"`
	ChunkIndex  int    `json:"chunk_index"`
	TotalChunks int    `json:"total_chunks"`
	UploadID    string `json:"upload_id"`
	IsEncrypted bool   `json:"is_encrypted"`
	OwnerID     uint   `json:"owner_id" binding:"required"`
}

// UploadResponse 文件上传响应
type UploadResponse struct {
	FileID       uint   `json:"file_id"`
	FileName     string `json:"file_name"`
	FileURL      string `json:"file_url"`
	ThumbnailURL string `json:"thumbnail_url"`
	PreviewURL   string `json:"preview_url"`
	UploadID     string `json:"upload_id"`
	ChunkIndex   int    `json:"chunk_index"`
	IsComplete   bool   `json:"is_complete"`
}

// PreviewRequest 文件预览请求
type PreviewRequest struct {
	FileID      uint   `json:"file_id" binding:"required"`
	PreviewType string `json:"preview_type"` // image/video/audio/document
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Quality     string `json:"quality"`
}

// VersionRequest 文件版本请求
type VersionRequest struct {
	FileID      uint   `json:"file_id" binding:"required"`
	Version     int    `json:"version"`
	NewFileName string `json:"new_file_name"`
	OwnerID     uint   `json:"owner_id" binding:"required"`
}

// UploadFile 上传文件
func (s *FileService) UploadFile(req UploadRequest) (*UploadResponse, error) {
	// 生成文件哈希值
	fileHash := s.calculateFileHash(req.FileData)

	// 检查文件是否已存在
	var existingFile model.File
	err := s.db.Where("file_hash = ? AND owner_id = ?", fileHash, req.OwnerID).First(&existingFile).Error
	if err == nil {
		// 文件已存在，返回现有文件信息
		return &UploadResponse{
			FileID:       existingFile.ID,
			FileName:     existingFile.FileName,
			FileURL:      existingFile.StorageURL,
			ThumbnailURL: existingFile.Thumbnail,
			PreviewURL:   existingFile.Preview,
			UploadID:     req.UploadID,
			ChunkIndex:   req.ChunkIndex,
			IsComplete:   true,
		}, nil
	}

	// 如果是分片上传
	if req.TotalChunks > 1 {
		return s.handleChunkedUpload(req, fileHash)
	}

	// 单文件上传
	return s.handleSingleUpload(req, fileHash)
}

// handleSingleUpload 处理单文件上传
func (s *FileService) handleSingleUpload(req UploadRequest, fileHash string) (*UploadResponse, error) {
	// 生成存储路径
	storagePath := s.generateStoragePath(req.FileName)

	// 保存文件到存储
	if err := s.saveFileToStorage(storagePath, req.FileData); err != nil {
		return nil, fmt.Errorf("保存文件失败: %v", err)
	}

	// 生成访问URL
	fileURL := s.generateFileURL(storagePath)

	// 创建文件记录
	file := model.File{
		FileName:    req.FileName,
		FileSize:    req.FileSize,
		FileType:    req.FileType,
		MimeType:    req.MimeType,
		FileHash:    fileHash,
		StoragePath: storagePath,
		StorageURL:  fileURL,
		IsEncrypted: req.IsEncrypted,
		Version:     1,
		IsLatest:    true,
		OwnerID:     req.OwnerID,
		IsPublic:    false,
		UploadID:    req.UploadID,
	}

	// 如果是加密文件，生成加密密钥
	if req.IsEncrypted {
		encryptionKey, err := s.generateEncryptionKey()
		if err != nil {
			return nil, fmt.Errorf("生成加密密钥失败: %v", err)
		}
		file.EncryptionKey = encryptionKey
	}

	// 保存到数据库
	if err := s.db.Create(&file).Error; err != nil {
		return nil, fmt.Errorf("创建文件记录失败: %v", err)
	}

	// 生成缩略图和预览
	thumbnailURL, previewURL := s.generatePreview(&file)

	return &UploadResponse{
		FileID:       file.ID,
		FileName:     file.FileName,
		FileURL:      file.StorageURL,
		ThumbnailURL: thumbnailURL,
		PreviewURL:   previewURL,
		UploadID:     req.UploadID,
		ChunkIndex:   0,
		IsComplete:   true,
	}, nil
}

// handleChunkedUpload 处理分片上传
func (s *FileService) handleChunkedUpload(req UploadRequest, fileHash string) (*UploadResponse, error) {
	// 检查是否为新的上传ID
	if req.UploadID == "" {
		req.UploadID = s.generateUploadID()
	}

	// 创建或更新分片记录
	chunk := model.FileChunk{
		UploadID:   req.UploadID,
		ChunkIndex: req.ChunkIndex,
		ChunkSize:  int64(len(req.FileData)),
		ChunkHash:  s.calculateFileHash(req.FileData),
	}

	// 生成分片存储路径
	chunkPath := s.generateChunkPath(req.UploadID, req.ChunkIndex)
	chunk.StoragePath = chunkPath

	// 保存分片文件
	if err := s.saveFileToStorage(chunkPath, req.FileData); err != nil {
		return nil, fmt.Errorf("保存分片失败: %v", err)
	}

	chunk.IsUploaded = true
	chunk.IsVerified = true

	// 保存分片记录
	if err := s.db.Create(&chunk).Error; err != nil {
		return nil, fmt.Errorf("创建分片记录失败: %v", err)
	}

	// 检查是否所有分片都已上传
	var uploadedChunks int64
	s.db.Model(&model.FileChunk{}).Where("upload_id = ? AND is_uploaded = ?", req.UploadID, true).Count(&uploadedChunks)

	if uploadedChunks == int64(req.TotalChunks) {
		// 所有分片都已上传，合并文件
		return s.mergeChunks(req, fileHash)
	}

	return &UploadResponse{
		UploadID:   req.UploadID,
		ChunkIndex: req.ChunkIndex,
		IsComplete: false,
	}, nil
}

// mergeChunks 合并分片文件
func (s *FileService) mergeChunks(req UploadRequest, fileHash string) (*UploadResponse, error) {
	// 获取所有分片
	var chunks []model.FileChunk
	if err := s.db.Where("upload_id = ? AND is_uploaded = ?", req.UploadID, true).
		Order("chunk_index ASC").Find(&chunks).Error; err != nil {
		return nil, fmt.Errorf("获取分片失败: %v", err)
	}

	// 生成最终文件路径
	finalPath := s.generateStoragePath(req.FileName)
	finalFile, err := os.Create(finalPath)
	if err != nil {
		return nil, fmt.Errorf("创建最终文件失败: %v", err)
	}
	defer finalFile.Close()

	// 合并所有分片
	for _, chunk := range chunks {
		chunkFile, err := os.Open(chunk.StoragePath)
		if err != nil {
			return nil, fmt.Errorf("打开分片文件失败: %v", err)
		}

		_, err = io.Copy(finalFile, chunkFile)
		chunkFile.Close()

		if err != nil {
			return nil, fmt.Errorf("合并分片失败: %v", err)
		}
	}

	// 生成访问URL
	fileURL := s.generateFileURL(finalPath)

	// 创建文件记录
	file := model.File{
		FileName:    req.FileName,
		FileSize:    req.FileSize,
		FileType:    req.FileType,
		MimeType:    req.MimeType,
		FileHash:    fileHash,
		StoragePath: finalPath,
		StorageURL:  fileURL,
		IsEncrypted: req.IsEncrypted,
		Version:     1,
		IsLatest:    true,
		OwnerID:     req.OwnerID,
		IsPublic:    false,
		UploadID:    req.UploadID,
		ChunkSize:   int64(len(chunks)),
		TotalChunks: len(chunks),
	}

	// 保存到数据库
	if err := s.db.Create(&file).Error; err != nil {
		return nil, fmt.Errorf("创建文件记录失败: %v", err)
	}

	// 清理分片文件
	go s.cleanupChunks(req.UploadID)

	// 生成缩略图和预览
	thumbnailURL, previewURL := s.generatePreview(&file)

	return &UploadResponse{
		FileID:       file.ID,
		FileName:     file.FileName,
		FileURL:      file.StorageURL,
		ThumbnailURL: thumbnailURL,
		PreviewURL:   previewURL,
		UploadID:     req.UploadID,
		ChunkIndex:   req.TotalChunks,
		IsComplete:   true,
	}, nil
}

// GetFilePreview 获取文件预览
func (s *FileService) GetFilePreview(req PreviewRequest) (*model.FilePreview, error) {
	// 查找文件
	var file model.File
	if err := s.db.Where("id = ?", req.FileID).First(&file).Error; err != nil {
		return nil, errors.New("文件不存在")
	}

	// 检查是否已有预览
	var preview model.FilePreview
	err := s.db.Where("file_id = ? AND preview_type = ?", req.FileID, req.PreviewType).First(&preview).Error
	if err == nil {
		return &preview, nil
	}

	// 生成新的预览
	previewURL := s.generatePreviewURL(&file, req.PreviewType, req.Width, req.Height, req.Quality)
	thumbnailURL := s.generateThumbnailURL(&file)

	preview = model.FilePreview{
		FileID:       req.FileID,
		PreviewType:  req.PreviewType,
		PreviewURL:   previewURL,
		ThumbnailURL: thumbnailURL,
		Width:        req.Width,
		Height:       req.Height,
		Quality:      req.Quality,
	}

	// 保存预览记录
	if err := s.db.Create(&preview).Error; err != nil {
		return nil, fmt.Errorf("创建预览记录失败: %v", err)
	}

	return &preview, nil
}

// CreateFileVersion 创建文件版本
func (s *FileService) CreateFileVersion(req VersionRequest) (*model.File, error) {
	// 查找原文件
	var originalFile model.File
	if err := s.db.Where("id = ? AND owner_id = ?", req.FileID, req.OwnerID).First(&originalFile).Error; err != nil {
		return nil, errors.New("文件不存在或无权限")
	}

	// 将原文件标记为非最新版本
	originalFile.IsLatest = false
	s.db.Save(&originalFile)

	// 创建新版本
	newFile := originalFile
	newFile.ID = 0 // 重置ID
	newFile.ParentID = &originalFile.ID
	newFile.Version = originalFile.Version + 1
	newFile.IsLatest = true
	newFile.CreatedAt = time.Now()
	newFile.UpdatedAt = time.Now()

	if req.NewFileName != "" {
		newFile.FileName = req.NewFileName
	}

	// 保存新版本
	if err := s.db.Create(&newFile).Error; err != nil {
		return nil, fmt.Errorf("创建文件版本失败: %v", err)
	}

	return &newFile, nil
}

// GetFileVersions 获取文件版本列表
func (s *FileService) GetFileVersions(fileID uint, ownerID uint) ([]model.File, error) {
	// 查找原文件
	var originalFile model.File
	if err := s.db.Where("id = ? AND owner_id = ?", fileID, ownerID).First(&originalFile).Error; err != nil {
		return nil, errors.New("文件不存在或无权限")
	}

	// 获取所有版本
	var versions []model.File
	query := s.db.Where("parent_id = ? OR id = ?", originalFile.ID, originalFile.ID).
		Order("version DESC")

	if err := query.Find(&versions).Error; err != nil {
		return nil, fmt.Errorf("获取文件版本失败: %v", err)
	}

	return versions, nil
}

// DeleteFile 删除文件
func (s *FileService) DeleteFile(fileID uint, ownerID uint) error {
	// 查找文件
	var file model.File
	if err := s.db.Where("id = ? AND owner_id = ?", fileID, ownerID).First(&file).Error; err != nil {
		return errors.New("文件不存在或无权限")
	}

	// 软删除文件记录
	if err := s.db.Delete(&file).Error; err != nil {
		return fmt.Errorf("删除文件记录失败: %v", err)
	}

	// 异步删除物理文件
	go s.deletePhysicalFile(file.StoragePath)

	return nil
}

// GetFile 获取文件信息
func (s *FileService) GetFile(fileID uint, userID uint) (*model.File, error) {
	var file model.File
	query := s.db.Where("id = ?", fileID)

	// 如果不是公开文件，需要检查权限
	if err := query.First(&file).Error; err != nil {
		return nil, errors.New("文件不存在")
	}

	if !file.IsPublic && file.OwnerID != userID {
		return nil, errors.New("无权限访问该文件")
	}

	// 增加查看次数
	s.db.Model(&file).Update("view_count", gorm.Expr("view_count + 1"))

	return &file, nil
}

// 辅助方法

// calculateFileHash 计算文件哈希值
func (s *FileService) calculateFileHash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// generateStoragePath 生成存储路径
func (s *FileService) generateStoragePath(fileName string) string {
	timestamp := time.Now().Format("2006/01/02")
	hash := md5.Sum([]byte(fileName + time.Now().String()))
	hashStr := hex.EncodeToString(hash[:])
	return fmt.Sprintf("uploads/%s/%s_%s", timestamp, hashStr, fileName)
}

// generateChunkPath 生成分片路径
func (s *FileService) generateChunkPath(uploadID string, chunkIndex int) string {
	return fmt.Sprintf("chunks/%s/chunk_%d", uploadID, chunkIndex)
}

// generateFileURL 生成文件访问URL
func (s *FileService) generateFileURL(storagePath string) string {
	return fmt.Sprintf("/api/files/%s", storagePath)
}

// generatePreviewURL 生成预览URL
func (s *FileService) generatePreviewURL(file *model.File, previewType string, width, height int, quality string) string {
	return fmt.Sprintf("/api/files/%d/preview?type=%s&w=%d&h=%d&q=%s",
		file.ID, previewType, width, height, quality)
}

// generateThumbnailURL 生成缩略图URL
func (s *FileService) generateThumbnailURL(file *model.File) string {
	return fmt.Sprintf("/api/files/%d/thumbnail", file.ID)
}

// generateUploadID 生成上传ID
func (s *FileService) generateUploadID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// generateEncryptionKey 生成加密密钥
func (s *FileService) generateEncryptionKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return hex.EncodeToString(key), nil
}

// saveFileToStorage 保存文件到存储
func (s *FileService) saveFileToStorage(storagePath string, data []byte) error {
	// 创建目录
	dir := filepath.Dir(storagePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 写入文件
	return os.WriteFile(storagePath, data, 0644)
}

// generatePreview 生成预览和缩略图
func (s *FileService) generatePreview(file *model.File) (string, string) {
	thumbnailURL := s.generateThumbnailURL(file)
	previewURL := ""

	// 根据文件类型生成预览URL
	switch {
	case strings.HasPrefix(file.MimeType, "image/"):
		previewURL = s.generatePreviewURL(file, "image", 800, 600, "high")
	case strings.HasPrefix(file.MimeType, "video/"):
		previewURL = s.generatePreviewURL(file, "video", 1280, 720, "medium")
	case strings.HasPrefix(file.MimeType, "audio/"):
		previewURL = s.generatePreviewURL(file, "audio", 0, 0, "high")
	default:
		previewURL = s.generatePreviewURL(file, "document", 0, 0, "high")
	}

	return thumbnailURL, previewURL
}

// cleanupChunks 清理分片文件
func (s *FileService) cleanupChunks(uploadID string) {
	// 删除分片记录
	s.db.Where("upload_id = ?", uploadID).Delete(&model.FileChunk{})

	// 删除物理分片文件
	chunkDir := fmt.Sprintf("chunks/%s", uploadID)
	os.RemoveAll(chunkDir)
}

// deletePhysicalFile 删除物理文件
func (s *FileService) deletePhysicalFile(storagePath string) {
	os.Remove(storagePath)
}
