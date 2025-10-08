package controller

import (
	"net/http"
	"strconv"

	"zhihang-messenger/im-backend/config"
	"zhihang-messenger/im-backend/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// FileController 文件控制器
type FileController struct {
	fileService *service.FileService
}

// NewFileController 创建文件控制器实例
func NewFileController() *FileController {
	return &FileController{
		fileService: service.NewFileService(),
	}
}

// UploadFile 上传文件
// @Summary 上传文件
// @Description 支持单文件和分片上传
// @Tags 文件管理
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "文件"
// @Param is_encrypted formData bool false "是否加密"
// @Success 200 {object} service.UploadResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/files/upload [post]
func (c *FileController) UploadFile(ctx *gin.Context) {
	// 获取用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 获取上传的文件
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "获取文件失败: " + err.Error()})
		return
	}

	// 读取文件内容
	src, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "打开文件失败: " + err.Error()})
		return
	}
	defer src.Close()

	// 读取文件数据
	fileData := make([]byte, file.Size)
	if _, err := src.Read(fileData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "读取文件失败: " + err.Error()})
		return
	}

	// 获取其他参数
	isEncrypted := ctx.PostForm("is_encrypted") == "true"

	// 构建上传请求
	uploadReq := service.UploadRequest{
		FileName:    file.Filename,
		FileSize:    file.Size,
		FileType:    getFileType(file.Filename),
		MimeType:    getMimeType(file.Filename),
		FileData:    fileData,
		IsEncrypted: isEncrypted,
		OwnerID:     userID.(uint),
	}

	// 执行上传
	response, err := c.fileService.UploadFile(uploadReq)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// UploadChunk 上传文件分片
// @Summary 上传文件分片
// @Description 用于大文件分片上传
// @Tags 文件管理
// @Accept multipart/form-data
// @Produce json
// @Param chunk formData file true "文件分片"
// @Param upload_id formData string true "上传ID"
// @Param chunk_index formData int true "分片索引"
// @Param total_chunks formData int true "总分片数"
// @Param file_name formData string true "文件名"
// @Param file_size formData int64 true "文件大小"
// @Param is_encrypted formData bool false "是否加密"
// @Success 200 {object} service.UploadResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/files/upload-chunk [post]
func (c *FileController) UploadChunk(ctx *gin.Context) {
	// 获取用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 获取上传的分片
	file, err := ctx.FormFile("chunk")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "获取分片失败: " + err.Error()})
		return
	}

	// 读取分片内容
	src, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "打开分片失败: " + err.Error()})
		return
	}
	defer src.Close()

	// 读取分片数据
	chunkData := make([]byte, file.Size)
	if _, err := src.Read(chunkData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "读取分片失败: " + err.Error()})
		return
	}

	// 获取参数
	uploadID := ctx.PostForm("upload_id")
	if uploadID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "upload_id 不能为空"})
		return
	}

	chunkIndex, err := strconv.Atoi(ctx.PostForm("chunk_index"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "chunk_index 格式错误"})
		return
	}

	totalChunks, err := strconv.Atoi(ctx.PostForm("total_chunks"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "total_chunks 格式错误"})
		return
	}

	fileName := ctx.PostForm("file_name")
	if fileName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file_name 不能为空"})
		return
	}

	fileSize, err := strconv.ParseInt(ctx.PostForm("file_size"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file_size 格式错误"})
		return
	}

	isEncrypted := ctx.PostForm("is_encrypted") == "true"

	// 构建上传请求
	uploadReq := service.UploadRequest{
		FileName:    fileName,
		FileSize:    fileSize,
		FileType:    getFileType(fileName),
		MimeType:    getMimeType(fileName),
		FileData:    chunkData,
		ChunkIndex:  chunkIndex,
		TotalChunks: totalChunks,
		UploadID:    uploadID,
		IsEncrypted: isEncrypted,
		OwnerID:     userID.(uint),
	}

	// 执行分片上传
	response, err := c.fileService.UploadFile(uploadReq)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// GetFile 获取文件信息
// @Summary 获取文件信息
// @Description 获取指定文件的详细信息
// @Tags 文件管理
// @Produce json
// @Param file_id path int true "文件ID"
// @Success 200 {object} model.File
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/files/{file_id} [get]
func (c *FileController) GetFile(ctx *gin.Context) {
	// 获取用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 获取文件ID
	fileIDStr := ctx.Param("file_id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "文件ID格式错误"})
		return
	}

	// 获取文件信息
	file, err := c.fileService.GetFile(uint(fileID), userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, file)
}

// DownloadFile 下载文件
// @Summary 下载文件
// @Description 下载指定文件
// @Tags 文件管理
// @Param file_id path int true "文件ID"
// @Success 200 {file} binary
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/files/{file_id}/download [get]
func (c *FileController) DownloadFile(ctx *gin.Context) {
	// 获取用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 获取文件ID
	fileIDStr := ctx.Param("file_id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "文件ID格式错误"})
		return
	}

	// 获取文件信息
	file, err := c.fileService.GetFile(uint(fileID), userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 设置下载响应头
	ctx.Header("Content-Disposition", "attachment; filename=\""+file.FileName+"\"")
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.File(file.StoragePath)

	// 异步增加下载次数
	go func() {
		config.DB.Model(&file).Update("download_count", gorm.Expr("download_count + 1"))
	}()
}

// GetFilePreview 获取文件预览
// @Summary 获取文件预览
// @Description 获取文件的预览信息
// @Tags 文件管理
// @Produce json
// @Param file_id path int true "文件ID"
// @Param preview_type query string false "预览类型" Enums(image,video,audio,document)
// @Param width query int false "宽度"
// @Param height query int false "高度"
// @Param quality query string false "质量" Enums(low,medium,high)
// @Success 200 {object} model.FilePreview
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/files/{file_id}/preview [get]
func (c *FileController) GetFilePreview(ctx *gin.Context) {
	// 获取用户ID
	_, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 获取文件ID
	fileIDStr := ctx.Param("file_id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "文件ID格式错误"})
		return
	}

	// 获取查询参数
	previewType := ctx.DefaultQuery("preview_type", "image")
	width, _ := strconv.Atoi(ctx.DefaultQuery("width", "800"))
	height, _ := strconv.Atoi(ctx.DefaultQuery("height", "600"))
	quality := ctx.DefaultQuery("quality", "medium")

	// 构建预览请求
	previewReq := service.PreviewRequest{
		FileID:      uint(fileID),
		PreviewType: previewType,
		Width:       width,
		Height:      height,
		Quality:     quality,
	}

	// 获取文件预览
	preview, err := c.fileService.GetFilePreview(previewReq)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, preview)
}

// GetFileVersions 获取文件版本列表
// @Summary 获取文件版本列表
// @Description 获取指定文件的所有版本
// @Tags 文件管理
// @Produce json
// @Param file_id path int true "文件ID"
// @Success 200 {array} model.File
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/files/{file_id}/versions [get]
func (c *FileController) GetFileVersions(ctx *gin.Context) {
	// 获取用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 获取文件ID
	fileIDStr := ctx.Param("file_id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "文件ID格式错误"})
		return
	}

	// 获取文件版本列表
	versions, err := c.fileService.GetFileVersions(uint(fileID), userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, versions)
}

// CreateFileVersion 创建文件版本
// @Summary 创建文件版本
// @Description 为指定文件创建新版本
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param file_id path int true "文件ID"
// @Param request body service.VersionRequest true "版本请求"
// @Success 200 {object} model.File
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/files/{file_id}/versions [post]
func (c *FileController) CreateFileVersion(ctx *gin.Context) {
	// 获取用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 获取文件ID
	fileIDStr := ctx.Param("file_id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "文件ID格式错误"})
		return
	}

	// 解析请求体
	var versionReq service.VersionRequest
	if err := ctx.ShouldBindJSON(&versionReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 设置文件ID和用户ID
	versionReq.FileID = uint(fileID)
	versionReq.OwnerID = userID.(uint)

	// 创建文件版本
	newVersion, err := c.fileService.CreateFileVersion(versionReq)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, newVersion)
}

// DeleteFile 删除文件
// @Summary 删除文件
// @Description 删除指定文件
// @Tags 文件管理
// @Param file_id path int true "文件ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/files/{file_id} [delete]
func (c *FileController) DeleteFile(ctx *gin.Context) {
	// 获取用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 获取文件ID
	fileIDStr := ctx.Param("file_id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "文件ID格式错误"})
		return
	}

	// 删除文件
	if err := c.fileService.DeleteFile(uint(fileID), userID.(uint)); err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "文件删除成功"})
}

// 辅助函数

// getFileType 根据文件名获取文件类型
func getFileType(fileName string) string {
	ext := getFileExtension(fileName)
	switch ext {
	case "jpg", "jpeg", "png", "gif", "bmp", "webp", "svg":
		return "image"
	case "mp4", "avi", "mov", "wmv", "flv", "webm", "mkv":
		return "video"
	case "mp3", "wav", "flac", "aac", "ogg", "wma":
		return "audio"
	case "pdf", "doc", "docx", "xls", "xlsx", "ppt", "pptx", "txt":
		return "document"
	default:
		return "file"
	}
}

// getMimeType 根据文件名获取MIME类型
func getMimeType(fileName string) string {
	// 这里可以使用更完善的MIME类型检测库
	// 简化实现，只处理常见类型
	ext := getFileExtension(fileName)
	switch ext {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "pdf":
		return "application/pdf"
	case "mp4":
		return "video/mp4"
	case "mp3":
		return "audio/mpeg"
	default:
		return "application/octet-stream"
	}
}

// getFileExtension 获取文件扩展名
func getFileExtension(fileName string) string {
	if lastDot := lastIndexOf(fileName, "."); lastDot != -1 {
		return fileName[lastDot+1:]
	}
	return ""
}

// lastIndexOf 查找字符最后出现的位置
func lastIndexOf(s, substr string) int {
	for i := len(s) - len(substr); i >= 0; i-- {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
