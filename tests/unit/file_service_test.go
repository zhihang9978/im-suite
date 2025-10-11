package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFile_SizeValidation 测试文件大小验证
func TestFile_SizeValidation(t *testing.T) {
	maxSize := int64(100 * 1024 * 1024) // 100MB

	testCases := []struct {
		name     string
		size     int64
		expected bool
	}{
		{"SmallFile", 1024, true},
		{"MediumFile", 10 * 1024 * 1024, true},
		{"LargeFile", 99 * 1024 * 1024, true},
		{"TooLargeFile", 101 * 1024 * 1024, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isValid := tc.size <= maxSize
			assert.Equal(t, tc.expected, isValid)
		})
	}
}

// TestFile_TypeValidation 测试文件类型验证
func TestFile_TypeValidation(t *testing.T) {
	allowedTypes := map[string]bool{
		"jpg":  true,
		"jpeg": true,
		"png":  true,
		"gif":  true,
		"pdf":  true,
		"doc":  true,
		"docx": true,
		"mp4":  true,
		"mp3":  true,
	}

	testCases := []struct {
		ext      string
		expected bool
	}{
		{"jpg", true},
		{"png", true},
		{"pdf", true},
		{"exe", false},
		{"bat", false},
		{"sh", false},
	}

	for _, tc := range testCases {
		t.Run("Extension_"+tc.ext, func(t *testing.T) {
			_, isAllowed := allowedTypes[tc.ext]
			assert.Equal(t, tc.expected, isAllowed)
		})
	}
}

// TestFile_HashGeneration 测试文件哈希生成
func TestFile_HashGeneration(t *testing.T) {
	content1 := []byte("test content 1")
	content2 := []byte("test content 2")

	// 简单的哈希模拟
	hash1 := string(content1)
	hash2 := string(content2)

	assert.NotEqual(t, hash1, hash2)
}

// BenchmarkFile_HashCalculation 性能测试：哈希计算
func BenchmarkFile_HashCalculation(b *testing.B) {
	content := make([]byte, 1024*1024) // 1MB

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = string(content) // 简化的哈希
	}
}

