package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

// MinIOClient wraps MinIO operations for document storage
type MinIOClient struct {
	client *minio.Client
	logger *zap.Logger
}

// NewMinIOClient creates a new MinIO client
func NewMinIOClient(endpoint, accessKey, secretKey string, useSSL bool, logger *zap.Logger) (*MinIOClient, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}

	return &MinIOClient{
		client: minioClient,
		logger: logger,
	}, nil
}

// CreateBucket creates a new bucket for a tenant
func (c *MinIOClient) CreateBucket(ctx context.Context, bucketName string) error {
	exists, err := c.client.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("check bucket exists: %w", err)
	}

	if exists {
		c.logger.Debug("bucket already exists", zap.String("bucket", bucketName))
		return nil
	}

	err = c.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: "us-east-1"})
	if err != nil {
		return fmt.Errorf("create bucket: %w", err)
	}

	c.logger.Info("bucket created", zap.String("bucket", bucketName))
	return nil
}

// PutObject uploads an object to MinIO
func (c *MinIOClient) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, size int64, contentType string) error {
	info, err := c.client.PutObject(ctx, bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("put object: %w", err)
	}

	c.logger.Debug("object uploaded",
		zap.String("bucket", bucketName),
		zap.String("object", objectName),
		zap.Int64("size", info.Size),
	)
	return nil
}

// GetObject downloads an object from MinIO
func (c *MinIOClient) GetObject(ctx context.Context, bucketName, objectName string) (io.ReadCloser, error) {
	object, err := c.client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}

	c.logger.Debug("object retrieved",
		zap.String("bucket", bucketName),
		zap.String("object", objectName),
	)
	return object, nil
}

// DeleteObject removes an object from MinIO
func (c *MinIOClient) DeleteObject(ctx context.Context, bucketName, objectName string) error {
	err := c.client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("delete object: %w", err)
	}

	c.logger.Debug("object deleted",
		zap.String("bucket", bucketName),
		zap.String("object", objectName),
	)
	return nil
}

// ObjectExists checks if an object exists
func (c *MinIOClient) ObjectExists(ctx context.Context, bucketName, objectName string) (bool, error) {
	info, err := c.client.StatObject(ctx, bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("stat object: %w", err)
	}

	return info.Size > 0, nil
}

// GetObjectSize retrieves the size of an object
func (c *MinIOClient) GetObjectSize(ctx context.Context, bucketName, objectName string) (int64, error) {
	info, err := c.client.StatObject(ctx, bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return 0, fmt.Errorf("stat object: %w", err)
	}

	return info.Size, nil
}

// PresignedGetObject generates a presigned URL for download
func (c *MinIOClient) PresignedGetObject(ctx context.Context, bucketName, objectName string, expiration time.Duration) (string, error) {
	presignedURL, err := c.client.PresignedGetObject(ctx, bucketName, objectName, expiration, nil)
	if err != nil {
		return "", fmt.Errorf("presigned get object: %w", err)
	}

	c.logger.Debug("presigned URL generated",
		zap.String("bucket", bucketName),
		zap.String("object", objectName),
		zap.Duration("expiration", expiration),
	)

	return presignedURL.String(), nil
}

// PresignedPutObject generates a presigned URL for upload
func (c *MinIOClient) PresignedPutObject(ctx context.Context, bucketName, objectName string, expiration time.Duration) (string, error) {
	presignedURL, err := c.client.PresignedPutObject(ctx, bucketName, objectName, expiration)
	if err != nil {
		return "", fmt.Errorf("presigned put object: %w", err)
	}

	c.logger.Debug("presigned upload URL generated",
		zap.String("bucket", bucketName),
		zap.String("object", objectName),
		zap.Duration("expiration", expiration),
	)

	return presignedURL.String(), nil
}

// Close closes the MinIO client connection
func (c *MinIOClient) Close() error {
	// MinIO client doesn't have explicit close in v7
	return nil
}
