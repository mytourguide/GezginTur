package service

// S3 uyumlu object storage (AWS S3 / MinIO / DigitalOcean Spaces) icin
// imzali (presigned) PUT URL uretimi. Istemci dosyayi dogrudan S3'e yukler;
// sunucumuz dosya trafiginden gecmez.

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"seyahat/backend/internal/config"
)

type UploadService struct {
	client    *s3.PresignClient
	bucket    string
	publicURL string
	enabled   bool
}

func NewUploadService(cfg *config.Config) *UploadService {
	if cfg.S3Bucket == "" || cfg.S3AccessKey == "" {
		return &UploadService{enabled: false} // yukleme devre disi
	}
	awsCfg, err := awscfg.LoadDefaultConfig(context.Background(),
		awscfg.WithRegion(cfg.S3Region),
		awscfg.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.S3AccessKey, cfg.S3SecretKey, "")),
		awscfg.WithBaseEndpoint(cfg.S3Endpoint), // bos ise AWS endpoint kullanilir
	)
	if err != nil {
		return &UploadService{enabled: false}
	}
	return &UploadService{
		client:    s3.NewPresignClient(s3.NewFromConfig(awsCfg)),
		bucket:    cfg.S3Bucket,
		publicURL: cfg.S3PublicURL,
		enabled:   true,
	}
}

func (s *UploadService) Enabled() bool { return s.enabled }

// SignPut, verilen anahtar icin 15 dakikalik imzali PUT URL uretir.
// Istemci bu URL'e dosyayi PUT eder; yetki imza ile sinirlandirilmistir.
func (s *UploadService) SignPut(ctx context.Context, key, contentType string) (string, error) {
	if !s.enabled {
		return "", fmt.Errorf("dosya yukleme yapilandirilmamis (S3_BUCKET/S3_ACCESS_KEY eksik)")
	}
	req, err := s.client.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

// PublicURLFor, yuklenen dosyanin erisilebilir URL'ini dondurur.
func (s *UploadService) PublicURLFor(key string) string {
	if s.publicURL != "" {
		return fmt.Sprintf("%s/%s", s.publicURL, key)
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, "eu-central-1", key)
}
