package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"GoLinko/pkg/zlog"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

// MinIOStorage MinIO 对象存储服务
// 用于分布式环境下的文件共享，替代本地文件存储
type MinIOStorage struct {
	client *minio.Client
	bucket string
	secure bool // 是否使用 HTTPS
}

// MinIOConfig MinIO 配置
type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

var defaultStorage *MinIOStorage

// InitMinIO 初始化 MinIO 存储服务
func InitMinIO(cfg MinIOConfig) error {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return fmt.Errorf("创建 MinIO 客户端失败: %w", err)
	}

	// 创建 bucket（如果不存在）
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return fmt.Errorf("检查 bucket 失败: %w", err)
	}

	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("创建 bucket 失败: %w", err)
		}
		zlog.GetLogger().Info("MinIO bucket 创建成功", zap.String("bucket", cfg.Bucket))
	}

	defaultStorage = &MinIOStorage{
		client: client,
		bucket: cfg.Bucket,
		secure: cfg.UseSSL,
	}

	zlog.GetLogger().Info("MinIO 初始化成功",
		zap.String("endpoint", cfg.Endpoint),
		zap.String("bucket", cfg.Bucket))

	return nil
}

// GetStorage 获取默认存储实例
func GetStorage() *MinIOStorage {
	return defaultStorage
}

// Upload 上传文件
// objectName: 对象名称（文件路径）
// reader: 文件内容
// size: 文件大小（-1 表示未知，使用分片上传）
// contentType: MIME 类型
func (s *MinIOStorage) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("MinIO 客户端未初始化")
	}

	_, err := s.client.PutObject(ctx, s.bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("上传文件失败: %w", err)
	}

	zlog.GetLogger().Debug("文件上传成功",
		zap.String("bucket", s.bucket),
		zap.String("object", objectName))

	return nil
}

// UploadBytes 上传字节数据
func (s *MinIOStorage) UploadBytes(ctx context.Context, objectName string, data []byte, contentType string) error {
	return s.Upload(ctx, objectName, bytes.NewReader(data), int64(len(data)), contentType)
}

// Download 下载文件
func (s *MinIOStorage) Download(ctx context.Context, objectName string) (*minio.Object, error) {
	if s == nil || s.client == nil {
		return nil, fmt.Errorf("MinIO 客户端未初始化")
	}

	obj, err := s.client.GetObject(ctx, s.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("下载文件失败: %w", err)
	}

	return obj, nil
}

// Delete 删除文件
func (s *MinIOStorage) Delete(ctx context.Context, objectName string) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("MinIO 客户端未初始化")
	}

	if err := s.client.RemoveObject(ctx, s.bucket, objectName, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}

	return nil
}

// GetURL 获取文件访问 URL
// 返回格式: http(s)://{endpoint}/{bucket}/{objectName}
func (s *MinIOStorage) GetURL(objectName string) string {
	if s == nil {
		return ""
	}

	protocol := "http"
	if s.secure {
		protocol = "https"
	}

	return fmt.Sprintf("%s://%s/%s/%s", protocol, s.client.EndpointURL().Host, s.bucket, objectName)
}

// GetPresignedURL 获取预签名 URL（临时访问链接）
// expiry: 链接过期时间
func (s *MinIOStorage) GetPresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	if s == nil || s.client == nil {
		return "", fmt.Errorf("MinIO 客户端未初始化")
	}

	url, err := s.client.PresignedGetObject(ctx, s.bucket, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("生成预签名 URL 失败: %w", err)
	}

	return url.String(), nil
}

// Exists 检查文件是否存在
func (s *MinIOStorage) Exists(ctx context.Context, objectName string) (bool, error) {
	if s == nil || s.client == nil {
		return false, fmt.Errorf("MinIO 客户端未初始化")
	}

	_, err := s.client.StatObject(ctx, s.bucket, objectName, minio.StatObjectOptions{})
	if err != nil {
		// 文件不存在
		return false, nil
	}

	return true, nil
}

// UploadAvatar 上传头像
// userID: 用户ID
// data: 图片数据
// ext: 文件扩展名（如 .png, .jpg）
func (s *MinIOStorage) UploadAvatar(ctx context.Context, userID string, data []byte, ext string) (string, error) {
	objectName := fmt.Sprintf("avatars/%s%s", userID, ext)
	contentType := "image/jpeg"
	if ext == ".png" {
		contentType = "image/png"
	} else if ext == ".gif" {
		contentType = "image/gif"
	}

	if err := s.UploadBytes(ctx, objectName, data, contentType); err != nil {
		return "", err
	}

	return s.GetURL(objectName), nil
}

// UploadFile 上传文件
// fileID: 文件唯一标识
// data: 文件数据
// contentType: MIME 类型
func (s *MinIOStorage) UploadFile(ctx context.Context, fileID string, data []byte, contentType string) (string, error) {
	objectName := fmt.Sprintf("files/%s", fileID)

	if err := s.UploadBytes(ctx, objectName, data, contentType); err != nil {
		return "", err
	}

	return s.GetURL(objectName), nil
}

// IsInitialized 检查 MinIO 是否已初始化
func IsInitialized() bool {
	return defaultStorage != nil && defaultStorage.client != nil
}
