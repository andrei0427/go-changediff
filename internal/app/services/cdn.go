package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"mime"
	"mime/multipart"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
)

type CDNService struct{}

func NewCDNService() *CDNService {
	return &CDNService{}
}

func (s *CDNService) UploadImage(file *multipart.FileHeader, maxWidth int) (*string, error) {
	if valid := s.isValidFileType(file); !valid {
		return nil, errors.New("invalid file type, only .jpg, .jpeg or .png supported")
	}

	if valid := s.isValidFileSize(file); !valid {
		return nil, errors.New("file too big, max file size allowed is 10MB")
	}

	src, err := file.Open()
	if err != nil {
		return nil, errors.New("error opening file")
	}

	inputImage, _, err := image.Decode(src)
	if err != nil {
		return nil, errors.New("error opening file")
	}

	resizedImage := imaging.Resize(inputImage, maxWidth, 0, imaging.Lanczos)

	var resizedImageBuf bytes.Buffer
	err = jpeg.Encode(&resizedImageBuf, resizedImage, nil)
	if err != nil {
		return nil, errors.New("error resizing image")
	}

	uniqueFileName := s.generateUniqueFileName(file)

	err = s.uploadFile(bytes.NewReader(resizedImageBuf.Bytes()), uniqueFileName, mime.TypeByExtension(s.GetFileExt(file.Filename)))
	if err != nil {
		return nil, err

	}

	return &uniqueFileName, nil
}

func (s *CDNService) GetFileExt(fileName string) string {
	fileParts := strings.Split(strings.ToLower(fileName), ".")

	if len(fileParts) == 1 {
		return ""
	}

	return fileParts[len(fileParts)-1]
}

func (s *CDNService) isValidFileType(file *multipart.FileHeader) bool {
	allowedExtensions := []string{".jpg", ".jpeg", ".png"}
	ext := "." + s.GetFileExt(file.Filename)

	for _, allowedExt := range allowedExtensions {
		if ext == allowedExt {
			return true
		}
	}

	return false
}

func (s *CDNService) isValidFileSize(file *multipart.FileHeader) bool {
	maxSize := int64(10 * 1024 * 1024) // 10MB
	return maxSize >= file.Size
}

func (s *CDNService) generateUniqueFileName(file *multipart.FileHeader) string {
	uid := uuid.New()
	return uid.String() + "_" + file.Filename
}

func (s *CDNService) uploadFile(r io.Reader, fileName string, mime string) error {
	cfR2Secret := os.Getenv("CLOUDFLARE_R2_SECRET")
	cfR2AccessKey := os.Getenv("CLOUDFLARE_R2_ACCESS_KEY")
	cfR2AccountId := os.Getenv("CLOUDFLARE_R2_ACCOUNT_ID")
	cfR2BucketName := os.Getenv("CLOUDFLARE_R2_BUCKET_NAME")

	if len(cfR2AccessKey) == 0 || len(cfR2Secret) == 0 || len(cfR2AccountId) == 0 || len(cfR2BucketName) == 0 {
		return errors.New("invalid or missing CloudFlare R2 Params in .env")
	}

	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service string, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfR2AccountId),
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithEndpointResolverWithOptions(r2Resolver), config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfR2AccessKey, cfR2Secret, "")))
	if err != nil {
		return errors.New("error loading r2 config")
	}

	ctx, cancelFn := context.WithTimeout(context.Background(), time.Minute)
	defer cancelFn()

	client := s3.NewFromConfig(cfg)

	_, err = client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(cfR2BucketName),
		Key:    aws.String(fileName),
	})

	if err == nil {
		return errors.New("file already exists")
	}

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(cfR2BucketName),
		Key:         aws.String(fileName),
		Body:        r,
		ContentType: aws.String(mime),
	})

	if err != nil {
		return errors.New("error uploading file")
	}

	return nil
}
