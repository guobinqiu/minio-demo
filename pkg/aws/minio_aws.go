package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type MinioClient struct {
	s3Client *s3.S3
	session  *session.Session
}

func NewMinioClient(endpoint, accessKey, secretKey string) *MinioClient {
	sess := session.Must(session.NewSession(&aws.Config{
		Region:           aws.String("us-east-1"),
		Endpoint:         aws.String(endpoint),
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	}))

	return &MinioClient{
		s3Client: s3.New(sess),
		session:  sess,
	}
}

// CreateBucket 创建存储桶
func (m *MinioClient) CreateBucket(bucketName string) error {
	_, err := m.s3Client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == s3.ErrCodeBucketAlreadyOwnedByYou {
				return nil
			}
		}
	}
	return err
}

// DeleteBucket 删除存储桶
func (m *MinioClient) DeleteBucket(bucketName string) error {
	_, err := m.s3Client.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	return err
}

// ListBuckets 列出所有存储桶
func (m *MinioClient) ListBuckets() ([]*s3.Bucket, error) {
	result, err := m.s3Client.ListBuckets(nil)
	if err != nil {
		return nil, err
	}
	return result.Buckets, nil
}

// UploadFile 上传文件到存储桶
func (m *MinioClient) UploadFile(bucketName, objectKey, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	uploader := s3manager.NewUploader(m.session)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   file,
	})
	return err
}

// DownloadFile 从存储桶下载文件
func (m *MinioClient) DownloadFile(bucketName, objectKey, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	downloader := s3manager.NewDownloader(m.session)
	_, err = downloader.Download(file, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	return err
}

// ListObjects 列出存储桶中的对象
func (m *MinioClient) ListObjects(bucketName string, recursive bool) ([]*s3.Object, error) {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	}
	if !recursive {
		input.Delimiter = aws.String("/")
	}

	result, err := m.s3Client.ListObjectsV2(input)
	if err != nil {
		return nil, err
	}
	return result.Contents, nil
}

// DeleteObject 删除存储桶中的对象
func (m *MinioClient) DeleteObject(bucketName, objectKey string) error {
	_, err := m.s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	return err
}

// RenameObject 重命名存储桶中的对象
func (m *MinioClient) RenameObject(bucketName, oldKey, newKey string) error {
	_, err := m.s3Client.CopyObject(&s3.CopyObjectInput{
		Bucket:     aws.String(bucketName),
		CopySource: aws.String(fmt.Sprintf("%s/%s", bucketName, oldKey)),
		Key:        aws.String(newKey),
	})
	if err != nil {
		return err
	}
	return m.DeleteObject(bucketName, oldKey)
}

// GeneratePresignedURL 生成预签名URL
func (m *MinioClient) GeneratePresignedURL(bucketName, objectKey string, expiry int64) (string, error) {
	req, _ := m.s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	return req.Presign(time.Duration(expiry) * time.Second)
}

func main() {
	// 创建 MinIO 客户端
	client := NewMinioClient("http://localhost:9000", "minioadmin", "minioadmin")

	// 创建 bucket
	err := client.CreateBucket("my-bucket")
	if err != nil {
		log.Fatal(err)
	}

	// 上传文件
	err = client.UploadFile("my-bucket", "girl.png", "../../girl.png")
	if err != nil {
		log.Fatal(err)
	}

	// 列出 bucket 中的对象
	objects, err := client.ListObjects("my-bucket", false)
	if err != nil {
		log.Fatal(err)
	}
	for _, obj := range objects {
		fmt.Println(*obj.Key)
	}

	// 生成预签名 URL
	url, err := client.GeneratePresignedURL("my-bucket", "girl.png", 3600)
	if err != nil {
		log.Fatalf("GeneratePresignedURL error: %v", err)
	}
	fmt.Printf("Presigned URL: %s\n", url)
}
