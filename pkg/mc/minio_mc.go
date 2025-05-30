package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinioClient MinIO客户端封装
type MinioClient struct {
	Client *minio.Client
	Ctx    context.Context
}

// NewMinioClient 创建新的MinIO客户端实例
func NewMinioClient(endpoint, accessKey, secretKey string) *MinioClient {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalf("Failed to create MinIO client: %v", err)
	}
	return &MinioClient{Client: client, Ctx: context.Background()}
}

// CreateBucket 创建存储桶
// 功能等价于: mc mb <bucket>
func (mc *MinioClient) CreateBucket(bucket string) error {
	exists, err := mc.Client.BucketExists(mc.Ctx, bucket)
	if err != nil {
		return err
	}
	if !exists {
		return mc.Client.MakeBucket(mc.Ctx, bucket, minio.MakeBucketOptions{})
	}
	return nil
}

// DeleteBucket 删除存储桶及其所有内容
// 功能等价于: mc rm --recursive --force <bucket>
func (mc *MinioClient) DeleteBucket(bucket string) error {
	// 先删除桶内所有对象
	objectsCh := make(chan minio.ObjectInfo)
	go func() {
		defer close(objectsCh)
		for object := range mc.Client.ListObjects(mc.Ctx, bucket, minio.ListObjectsOptions{Recursive: true}) {
			if object.Err != nil {
				continue
			}
			objectsCh <- object
		}
	}()

	// 等待删除完成
	for removeErr := range mc.Client.RemoveObjects(mc.Ctx, bucket, objectsCh, minio.RemoveObjectsOptions{}) {
		if removeErr.Err != nil {
			return removeErr.Err
		}
	}

	// 删除空桶
	return mc.Client.RemoveBucket(mc.Ctx, bucket)
}

// UploadFile 上传本地文件到MinIO
// 功能等价于: mc cp <local-file> <bucket>/<object>
func (mc *MinioClient) UploadFile(bucket, objectName, filePath string) error {
	_, err := mc.Client.FPutObject(mc.Ctx, bucket, objectName, filePath, minio.PutObjectOptions{})
	return err
}

// DownloadFile 从MinIO下载文件到本地
// 功能等价于: mc cp <bucket>/<object> <local-file>
func (mc *MinioClient) DownloadFile(bucket, objectName, filePath string) error {
	return mc.Client.FGetObject(mc.Ctx, bucket, objectName, filePath, minio.GetObjectOptions{})
}

// ListBuckets 列出所有存储桶
// 功能等价于: mc ls
func (mc *MinioClient) ListBuckets() error {
	buckets, err := mc.Client.ListBuckets(mc.Ctx)
	if err != nil {
		return err
	}
	for _, b := range buckets {
		fmt.Println(b.Name)
	}
	return nil
}

// ListObjects 列出存储桶中的对象
// 功能等价于: mc ls <bucket> [--recursive]
func (mc *MinioClient) ListObjects(bucket string, recursive bool) error {
	for object := range mc.Client.ListObjects(mc.Ctx, bucket, minio.ListObjectsOptions{Recursive: recursive}) {
		if object.Err != nil {
			return object.Err
		}
		fmt.Println(object.Key)
	}
	return nil
}

// DeleteObject 删除指定对象
// 功能等价于: mc rm <bucket>/<object>
func (mc *MinioClient) DeleteObject(bucket, object string) error {
	return mc.Client.RemoveObject(mc.Ctx, bucket, object, minio.RemoveObjectOptions{})
}

// GeneratePresignedURL 生成预签名下载URL
// 功能等价于: mc share download <bucket>/<object> --expire <duration>
func (mc *MinioClient) GeneratePresignedURL(bucket, object string, expirySec int64) (string, error) {
	reqParams := make(url.Values) // 可添加额外参数，如 response-content-type 等
	expiry := time.Duration(expirySec) * time.Second
	presignedURL, err := mc.Client.PresignedGetObject(mc.Ctx, bucket, object, expiry, reqParams)
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
}
