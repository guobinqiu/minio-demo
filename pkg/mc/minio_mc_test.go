package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"
)

// TestMinioClient 集成测试结构
type TestMinioClient struct {
	client     *MinioClient
	testBucket string
	testObject string
	testFile   string
}

// setupTest 设置测试环境
func setupTest(t *testing.T) *TestMinioClient {
	endpoint := getEnvOrDefault("MINIO_ENDPOINT", "localhost:9000")
	accessKey := getEnvOrDefault("MINIO_ACCESS_KEY", "minioadmin")
	secretKey := getEnvOrDefault("MINIO_SECRET_KEY", "minioadmin")

	client := NewMinioClient(endpoint, accessKey, secretKey)
	testBucket := "test-bucket-" + fmt.Sprintf("%d", time.Now().Unix())
	testObject := "test-object.txt"
	testFile := "/tmp/test-file.txt"

	// 创建测试文件
	err := ioutil.WriteFile(testFile, []byte("test content for minio"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	return &TestMinioClient{
		client:     client,
		testBucket: testBucket,
		testObject: testObject,
		testFile:   testFile,
	}
}

// cleanup 清理测试环境
func (tc *TestMinioClient) cleanup() {
	os.Remove(tc.testFile)
	tc.client.DeleteBucket(tc.testBucket) // 清理可能存在的测试桶
}

func TestNewMinioClient(t *testing.T) {
	client := NewMinioClient("localhost:9000", "testkey", "testsecret")

	if client == nil {
		t.Error("NewMinioClient returned nil")
	}
	if client.Client == nil {
		t.Error("MinIO client is nil")
	}
	if client.Ctx == nil {
		t.Error("Context is nil")
	}
}

func TestCreateBucket(t *testing.T) {
	tc := setupTest(t)
	defer tc.cleanup()

	err := tc.client.CreateBucket(tc.testBucket)
	if err != nil {
		t.Errorf("CreateBucket failed: %v", err)
	}

	// 验证桶是否创建成功
	exists, err := tc.client.Client.BucketExists(tc.client.Ctx, tc.testBucket)
	if err != nil {
		t.Errorf("BucketExists check failed: %v", err)
	}
	if !exists {
		t.Error("Bucket was not created")
	}

	// 测试重复创建（应该不报错）
	err = tc.client.CreateBucket(tc.testBucket)
	if err != nil {
		t.Errorf("Duplicate bucket creation should not fail: %v", err)
	}
}

func TestUploadFile(t *testing.T) {
	tc := setupTest(t)
	defer tc.cleanup()

	// 先创建桶
	err := tc.client.CreateBucket(tc.testBucket)
	if err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}

	// 上传文件
	err = tc.client.UploadFile(tc.testBucket, tc.testObject, tc.testFile)
	if err != nil {
		t.Errorf("UploadFile failed: %v", err)
	}
}

func TestDownloadFile(t *testing.T) {
	tc := setupTest(t)
	defer tc.cleanup()

	// 创建桶并上传文件
	err := tc.client.CreateBucket(tc.testBucket)
	if err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}

	err = tc.client.UploadFile(tc.testBucket, tc.testObject, tc.testFile)
	if err != nil {
		t.Fatalf("UploadFile failed: %v", err)
	}

	// 下载文件
	downloadFile := "/tmp/downloaded-test-file.txt"
	defer os.Remove(downloadFile)

	err = tc.client.DownloadFile(tc.testBucket, tc.testObject, downloadFile)
	if err != nil {
		t.Errorf("DownloadFile failed: %v", err)
	}

	// 验证文件内容
	originalContent, err := ioutil.ReadFile(tc.testFile)
	if err != nil {
		t.Fatalf("Failed to read original file: %v", err)
	}

	downloadedContent, err := ioutil.ReadFile(downloadFile)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(originalContent) != string(downloadedContent) {
		t.Error("Downloaded file content does not match original")
	}
}

func TestListBuckets(t *testing.T) {
	tc := setupTest(t)
	defer tc.cleanup()

	// 创建测试桶
	err := tc.client.CreateBucket(tc.testBucket)
	if err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}

	// 测试列出桶（这里主要测试不报错）
	err = tc.client.ListBuckets()
	if err != nil {
		t.Errorf("ListBuckets failed: %v", err)
	}
}

func TestListObjects(t *testing.T) {
	tc := setupTest(t)
	defer tc.cleanup()

	// 创建桶并上传文件
	err := tc.client.CreateBucket(tc.testBucket)
	if err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}

	err = tc.client.UploadFile(tc.testBucket, tc.testObject, tc.testFile)
	if err != nil {
		t.Fatalf("UploadFile failed: %v", err)
	}

	// 测试列出对象
	err = tc.client.ListObjects(tc.testBucket, false)
	if err != nil {
		t.Errorf("ListObjects failed: %v", err)
	}

	// 测试递归列出对象
	err = tc.client.ListObjects(tc.testBucket, true)
	if err != nil {
		t.Errorf("ListObjects recursive failed: %v", err)
	}
}

func TestDeleteObject(t *testing.T) {
	tc := setupTest(t)
	defer tc.cleanup()

	// 创建桶并上传文件
	err := tc.client.CreateBucket(tc.testBucket)
	if err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}

	err = tc.client.UploadFile(tc.testBucket, tc.testObject, tc.testFile)
	if err != nil {
		t.Fatalf("UploadFile failed: %v", err)
	}

	// 删除对象
	err = tc.client.DeleteObject(tc.testBucket, tc.testObject)
	if err != nil {
		t.Errorf("DeleteObject failed: %v", err)
	}

	// 验证对象是否被删除（尝试下载应该失败）
	downloadFile := "/tmp/should-not-exist.txt"
	err = tc.client.DownloadFile(tc.testBucket, tc.testObject, downloadFile)
	if err == nil {
		t.Error("Download should fail after object deletion")
		os.Remove(downloadFile) // 清理可能创建的文件
	}
}

func TestGeneratePresignedURL(t *testing.T) {
	tc := setupTest(t)
	defer tc.cleanup()

	// 创建桶并上传文件
	err := tc.client.CreateBucket(tc.testBucket)
	if err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}

	err = tc.client.UploadFile(tc.testBucket, tc.testObject, tc.testFile)
	if err != nil {
		t.Fatalf("UploadFile failed: %v", err)
	}

	// 生成预签名URL
	url, err := tc.client.GeneratePresignedURL(tc.testBucket, tc.testObject, 3600)
	if err != nil {
		t.Errorf("GeneratePresignedURL failed: %v", err)
		return
	}

	if url == "" {
		t.Error("Generated URL is empty")
		return
	}

	if !strings.HasPrefix(url, "http") {
		t.Error("Generated URL should start with http")
	}

	// 验证URL包含必要的参数
	if !strings.Contains(url, tc.testBucket) {
		t.Error("URL should contain bucket name")
	}
	if !strings.Contains(url, tc.testObject) {
		t.Error("URL should contain object name")
	}
	if !strings.Contains(url, "X-Amz-Expires") {
		t.Error("URL should contain expiration parameter")
	}
}

func TestDeleteBucket(t *testing.T) {
	tc := setupTest(t)
	defer tc.cleanup()

	// 创建桶并上传文件
	err := tc.client.CreateBucket(tc.testBucket)
	if err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}

	err = tc.client.UploadFile(tc.testBucket, tc.testObject, tc.testFile)
	if err != nil {
		t.Fatalf("UploadFile failed: %v", err)
	}

	// 删除桶（包括所有内容）
	err = tc.client.DeleteBucket(tc.testBucket)
	if err != nil {
		t.Errorf("DeleteBucket failed: %v", err)
	}

	// 验证桶是否被删除
	exists, err := tc.client.Client.BucketExists(tc.client.Ctx, tc.testBucket)
	if err != nil {
		t.Errorf("BucketExists check failed: %v", err)
	}
	if exists {
		t.Error("Bucket should be deleted")
	}
}

// 错误情况测试
func TestCreateBucketWithInvalidEndpoint(t *testing.T) {
	// 注意：这个测试可能会因为NewMinioClient中的log.Fatalf而退出程序
	// 在实际项目中，建议修改NewMinioClient返回error而不是直接Fatal
	client := NewMinioClient("invalid-endpoint:9000", "testkey", "testsecret")
	err := client.CreateBucket("test-bucket")
	if err == nil {
		t.Log("Expected error but got none - this might indicate the endpoint is actually reachable")
	}
}

func TestGeneratePresignedURLWithInvalidExpiry(t *testing.T) {
	tc := setupTest(t)
	defer tc.cleanup()

	// 测试负数过期时间
	url, err := tc.client.GeneratePresignedURL("bucket", "object", -1)
	if err == nil {
		t.Error("Expected error for negative expiry time")
	}
	if url != "" {
		t.Error("URL should be empty when error occurs")
	}
}

// 性能测试
func BenchmarkUploadFile(b *testing.B) {
	endpoint := getEnvOrDefault("MINIO_ENDPOINT", "localhost:9000")
	client := NewMinioClient(endpoint, "minioadmin", "minioadmin")
	testBucket := "benchmark-bucket"
	testFile := "/tmp/benchmark-test.txt"

	// 创建测试文件
	err := ioutil.WriteFile(testFile, []byte("benchmark test content"), 0644)
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(testFile)

	// 创建测试桶
	err = client.CreateBucket(testBucket)
	if err != nil {
		b.Fatal(err)
	}
	defer client.DeleteBucket(testBucket)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		objectName := fmt.Sprintf("benchmark-object-%d", i)
		err := client.UploadFile(testBucket, objectName, testFile)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// 工具函数
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
