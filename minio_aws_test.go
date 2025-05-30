package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testEndpoint  = "http://localhost:9000"
	testAccessKey = "minioadmin"
	testSecretKey = "minioadmin"
	testBucket    = "test-bucket"
	testObjectKey = "test-file.txt"
	testLocalFile = "test-file.txt"
	testContent   = "This is a test file content"
)

func setupTestFile() error {
	return os.WriteFile(testLocalFile, []byte(testContent), 0644)
}

func cleanupTestFile() {
	os.Remove(testLocalFile)
}

func TestMinioClient(t *testing.T) {
	// 创建测试文件
	err := setupTestFile()
	assert.NoError(t, err)
	defer cleanupTestFile()

	// 初始化客户端
	client := NewMinioClient(testEndpoint, testAccessKey, testSecretKey)

	// 测试创建存储桶
	t.Run("CreateBucket", func(t *testing.T) {
		err := client.CreateBucket(testBucket)
		assert.NoError(t, err)
	})

	// 测试上传文件
	t.Run("UploadFile", func(t *testing.T) {
		err := client.UploadFile(testBucket, testObjectKey, testLocalFile)
		assert.NoError(t, err)
	})

	// 测试列出对象
	t.Run("ListObjects", func(t *testing.T) {
		objects, err := client.ListObjects(testBucket, false)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(objects))
		assert.Equal(t, testObjectKey, *objects[0].Key)
	})

	// 测试下载文件
	t.Run("DownloadFile", func(t *testing.T) {
		downloadedFile := "downloaded-" + testLocalFile
		defer os.Remove(downloadedFile)

		err := client.DownloadFile(testBucket, testObjectKey, downloadedFile)
		assert.NoError(t, err)

		content, err := os.ReadFile(downloadedFile)
		assert.NoError(t, err)
		assert.Equal(t, testContent, string(content))
	})

	// 测试生成预签名URL
	t.Run("GeneratePresignedURL", func(t *testing.T) {
		url, err := client.GeneratePresignedURL(testBucket, testObjectKey, 3600)
		assert.NoError(t, err)
		assert.NotEmpty(t, url)
	})

	// 测试删除对象
	t.Run("DeleteObject", func(t *testing.T) {
		err := client.DeleteObject(testBucket, testObjectKey)
		assert.NoError(t, err)

		objects, err := client.ListObjects(testBucket, false)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(objects))
	})

	// 测试删除存储桶
	t.Run("DeleteBucket", func(t *testing.T) {
		err := client.DeleteBucket(testBucket)
		assert.NoError(t, err)
	})
}
