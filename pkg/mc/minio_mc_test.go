package main

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	testEndpoint  = "localhost:9000"
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
	uid := time.Now().UnixNano()
	bucket := fmt.Sprintf("test-bucket-%d", uid)
	object := fmt.Sprintf("test-object-%d.txt", uid)
	localFile := fmt.Sprintf("upload-%d.txt", uid)
	downloadedFile := fmt.Sprintf("downloaded-%d.txt", uid)

	err := setupTestFile(localFile)
	assert.NoError(t, err)
	defer cleanupFile(localFile)

	client := NewMinioClient(testEndpoint, testAccessKey, testSecretKey)

	t.Run("CreateBucket", func(t *testing.T) {
		err := client.CreateBucket(bucket)
		assert.NoError(t, err)
	})

	t.Run("UploadFile", func(t *testing.T) {
		err := client.UploadFile(bucket, object, localFile)
		assert.NoError(t, err)
	})

	t.Run("ListObjects", func(t *testing.T) {
		objects, err := client.ListObjects(bucket, false)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(objects))
		assert.Equal(t, object, objects[0].Key)
	})

	t.Run("DownloadFile", func(t *testing.T) {
		defer cleanupFile(downloadedFile)
		err := client.DownloadFile(bucket, object, downloadedFile)
		assert.NoError(t, err)

		data, err := os.ReadFile(downloadedFile)
		assert.NoError(t, err)
		assert.Equal(t, testContent, string(data))
	})

	t.Run("GeneratePresignedURL", func(t *testing.T) {
		url, err := client.GeneratePresignedURL(bucket, object, 3600)
		assert.NoError(t, err)
		assert.NotEmpty(t, url)
	})

	t.Run("DeleteObject", func(t *testing.T) {
		err := client.DeleteObject(bucket, object)
		assert.NoError(t, err)

		objects, err := client.ListObjects(bucket, false)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(objects))
	})

	t.Run("DeleteBucket", func(t *testing.T) {
		err := client.DeleteBucket(bucket)
		assert.NoError(t, err)
	})
}
