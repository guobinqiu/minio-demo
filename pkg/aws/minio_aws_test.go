package main

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	testEndpoint  = "http://localhost:9000"
	testAccessKey = "minioadmin"
	testSecretKey = "minioadmin"
	testContent   = "This is a test file content"
)

func setupTestFile(path string) error {
	return os.WriteFile(path, []byte(testContent), 0644)
}

func cleanupFile(path string) {
	_ = os.Remove(path)
}

func TestMinioClient(t *testing.T) {
	// 使用唯一前缀避免冲突
	uid := time.Now().UnixNano()
	testBucket := fmt.Sprintf("test-bucket-%d", uid)
	testObjectKey := fmt.Sprintf("test-object-%d.txt", uid)
	testLocalFile := fmt.Sprintf("upload-%d.txt", uid)

	err := setupTestFile(testLocalFile)
	assert.NoError(t, err)
	defer cleanupFile(testLocalFile)

	client := NewMinioClient(testEndpoint, testAccessKey, testSecretKey)

	t.Run("CreateBucket", func(t *testing.T) {
		err := client.CreateBucket(testBucket)
		assert.NoError(t, err)
	})

	t.Run("UploadFile", func(t *testing.T) {
		err := client.UploadFile(testBucket, testObjectKey, testLocalFile)
		assert.NoError(t, err)
	})

	t.Run("ListObjects", func(t *testing.T) {
		objects, err := client.ListObjects(testBucket, false)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(objects))
		assert.Equal(t, testObjectKey, *objects[0].Key)
	})

	t.Run("DownloadFile", func(t *testing.T) {
		downloadedFile := fmt.Sprintf("downloaded-%d.txt", uid)
		defer cleanupFile(downloadedFile)

		err := client.DownloadFile(testBucket, testObjectKey, downloadedFile)
		assert.NoError(t, err)

		content, err := os.ReadFile(downloadedFile)
		assert.NoError(t, err)
		assert.Equal(t, testContent, string(content))
	})

	t.Run("GeneratePresignedURL", func(t *testing.T) {
		url, err := client.GeneratePresignedURL(testBucket, testObjectKey, 3600)
		assert.NoError(t, err)
		assert.NotEmpty(t, url)
	})

	t.Run("DeleteObject", func(t *testing.T) {
		err := client.DeleteObject(testBucket, testObjectKey)
		assert.NoError(t, err)
		objects, err := client.ListObjects(testBucket, false)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(objects))
	})

	t.Run("DeleteBucket", func(t *testing.T) {
		err := client.DeleteBucket(testBucket)
		assert.NoError(t, err)
	})
}
