package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
)

// FileEncryptionService 文件加密服务
type FileEncryptionService struct{}

// NewFileEncryptionService 创建文件加密服务实例
func NewFileEncryptionService() *FileEncryptionService {
	return &FileEncryptionService{}
}

// EncryptFileRequest 文件加密请求
type EncryptFileRequest struct {
	FilePath    string `json:"file_path" binding:"required"`
	OutputPath  string `json:"output_path" binding:"required"`
	Key         string `json:"key,omitempty"` // 如果不提供，将生成新密钥
	Algorithm   string `json:"algorithm"`     // 加密算法，默认AES-256-GCM
}

// EncryptFileResponse 文件加密响应
type EncryptFileResponse struct {
	EncryptedPath string `json:"encrypted_path"`
	Key           string `json:"key"`           // 加密密钥
	Algorithm     string `json:"algorithm"`     // 使用的算法
	OriginalSize  int64  `json:"original_size"` // 原始文件大小
	EncryptedSize int64  `json:"encrypted_size"` // 加密后文件大小
}

// DecryptFileRequest 文件解密请求
type DecryptFileRequest struct {
	FilePath   string `json:"file_path" binding:"required"`
	OutputPath string `json:"output_path" binding:"required"`
	Key        string `json:"key" binding:"required"`
	Algorithm  string `json:"algorithm"` // 解密算法，默认AES-256-GCM
}

// DecryptFileResponse 文件解密响应
type DecryptFileResponse struct {
	DecryptedPath string `json:"decrypted_path"`
	OriginalSize  int64  `json:"original_size"` // 原始文件大小
	DecryptedSize int64  `json:"decrypted_size"` // 解密后文件大小
}

// EncryptFile 加密文件
func (s *FileEncryptionService) EncryptFile(req EncryptFileRequest) (*EncryptFileResponse, error) {
	// 读取原始文件
	fileData, err := os.ReadFile(req.FilePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}
	
	originalSize := int64(len(fileData))
	
	// 生成或使用提供的密钥
	var key []byte
	if req.Key != "" {
		key, err = hex.DecodeString(req.Key)
		if err != nil {
			return nil, fmt.Errorf("密钥格式错误: %v", err)
		}
	} else {
		key = s.generateKey()
	}
	
	// 选择加密算法
	algorithm := req.Algorithm
	if algorithm == "" {
		algorithm = "AES-256-GCM"
	}
	
	// 执行加密
	var encryptedData []byte
	switch algorithm {
	case "AES-256-GCM":
		encryptedData, err = s.encryptAESGCM(fileData, key)
	case "AES-256-CBC":
		encryptedData, err = s.encryptAESCBC(fileData, key)
	default:
		return nil, fmt.Errorf("不支持的加密算法: %s", algorithm)
	}
	
	if err != nil {
		return nil, fmt.Errorf("文件加密失败: %v", err)
	}
	
	// 写入加密文件
	if err := os.WriteFile(req.OutputPath, encryptedData, 0644); err != nil {
		return nil, fmt.Errorf("写入加密文件失败: %v", err)
	}
	
	encryptedSize := int64(len(encryptedData))
	
	return &EncryptFileResponse{
		EncryptedPath: req.OutputPath,
		Key:           hex.EncodeToString(key),
		Algorithm:     algorithm,
		OriginalSize:  originalSize,
		EncryptedSize: encryptedSize,
	}, nil
}

// DecryptFile 解密文件
func (s *FileEncryptionService) DecryptFile(req DecryptFileRequest) (*DecryptFileResponse, error) {
	// 读取加密文件
	encryptedData, err := os.ReadFile(req.FilePath)
	if err != nil {
		return nil, fmt.Errorf("读取加密文件失败: %v", err)
	}
	
	originalSize := int64(len(encryptedData))
	
	// 解析密钥
	key, err := hex.DecodeString(req.Key)
	if err != nil {
		return nil, fmt.Errorf("密钥格式错误: %v", err)
	}
	
	// 选择解密算法
	algorithm := req.Algorithm
	if algorithm == "" {
		algorithm = "AES-256-GCM"
	}
	
	// 执行解密
	var decryptedData []byte
	switch algorithm {
	case "AES-256-GCM":
		decryptedData, err = s.decryptAESGCM(encryptedData, key)
	case "AES-256-CBC":
		decryptedData, err = s.decryptAESCBC(encryptedData, key)
	default:
		return nil, fmt.Errorf("不支持的解密算法: %s", algorithm)
	}
	
	if err != nil {
		return nil, fmt.Errorf("文件解密失败: %v", err)
	}
	
	// 写入解密文件
	if err := os.WriteFile(req.OutputPath, decryptedData, 0644); err != nil {
		return nil, fmt.Errorf("写入解密文件失败: %v", err)
	}
	
	decryptedSize := int64(len(decryptedData))
	
	return &DecryptFileResponse{
		DecryptedPath: req.OutputPath,
		OriginalSize:  originalSize,
		DecryptedSize: decryptedSize,
	}, nil
}

// EncryptData 加密数据
func (s *FileEncryptionService) EncryptData(data []byte, key string) ([]byte, error) {
	keyBytes, err := hex.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("密钥格式错误: %v", err)
	}
	
	return s.encryptAESGCM(data, keyBytes)
}

// DecryptData 解密数据
func (s *FileEncryptionService) DecryptData(encryptedData []byte, key string) ([]byte, error) {
	keyBytes, err := hex.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("密钥格式错误: %v", err)
	}
	
	return s.decryptAESGCM(encryptedData, keyBytes)
}

// GenerateKey 生成加密密钥
func (s *FileEncryptionService) GenerateKey() string {
	key := s.generateKey()
	return hex.EncodeToString(key)
}

// GenerateKeyFromPassword 从密码生成密钥
func (s *FileEncryptionService) GenerateKeyFromPassword(password, salt string) string {
	hash := sha256.Sum256([]byte(password + salt))
	return hex.EncodeToString(hash[:])
}

// 私有方法

// generateKey 生成随机密钥
func (s *FileEncryptionService) generateKey() []byte {
	key := make([]byte, 32) // AES-256需要32字节密钥
	if _, err := rand.Read(key); err != nil {
		panic(err)
	}
	return key
}

// encryptAESGCM 使用AES-GCM加密
func (s *FileEncryptionService) encryptAESGCM(data, key []byte) ([]byte, error) {
	// 创建AES密码块
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	
	// 创建GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	
	// 生成随机nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	
	// 加密数据
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	
	return ciphertext, nil
}

// decryptAESGCM 使用AES-GCM解密
func (s *FileEncryptionService) decryptAESGCM(data, key []byte) ([]byte, error) {
	// 创建AES密码块
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	
	// 创建GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	
	// 检查数据长度
	if len(data) < gcm.NonceSize() {
		return nil, errors.New("加密数据太短")
	}
	
	// 提取nonce和密文
	nonce := data[:gcm.NonceSize()]
	ciphertext := data[gcm.NonceSize():]
	
	// 解密数据
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	
	return plaintext, nil
}

// encryptAESCBC 使用AES-CBC加密
func (s *FileEncryptionService) encryptAESCBC(data, key []byte) ([]byte, error) {
	// 创建AES密码块
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	
	// 填充数据到块大小的倍数
	paddedData := s.pkcs7Padding(data, aes.BlockSize)
	
	// 生成随机IV
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	
	// 创建CBC模式加密器
	mode := cipher.NewCBCEncrypter(block, iv)
	
	// 加密数据
	ciphertext := make([]byte, len(paddedData))
	mode.CryptBlocks(ciphertext, paddedData)
	
	// 将IV和密文组合
	result := make([]byte, len(iv)+len(ciphertext))
	copy(result, iv)
	copy(result[len(iv):], ciphertext)
	
	return result, nil
}

// decryptAESCBC 使用AES-CBC解密
func (s *FileEncryptionService) decryptAESCBC(data, key []byte) ([]byte, error) {
	// 创建AES密码块
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	
	// 检查数据长度
	if len(data) < aes.BlockSize {
		return nil, errors.New("加密数据太短")
	}
	
	// 提取IV和密文
	iv := data[:aes.BlockSize]
	ciphertext := data[aes.BlockSize:]
	
	// 创建CBC模式解密器
	mode := cipher.NewCBCDecrypter(block, iv)
	
	// 解密数据
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)
	
	// 去除填充
	return s.pkcs7UnPadding(plaintext), nil
}

// pkcs7Padding PKCS7填充
func (s *FileEncryptionService) pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := make([]byte, padding)
	for i := range padtext {
		padtext[i] = byte(padding)
	}
	return append(data, padtext...)
}

// pkcs7UnPadding PKCS7去填充
func (s *FileEncryptionService) pkcs7UnPadding(data []byte) []byte {
	length := len(data)
	if length == 0 {
		return data
	}
	unpadding := int(data[length-1])
	if unpadding > length {
		return data
	}
	return data[:(length - unpadding)]
}
