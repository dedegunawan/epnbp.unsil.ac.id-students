package utils

import (
	"bytes"
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var MinioClient *minio.Client

func InitStorage() {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")

	fmt.Println("Initializing minio client")

	useSSL := false
	useSSLStr := os.Getenv("MINIO_USE_SSL")
	if val, err := strconv.ParseBool(useSSLStr); err == nil {
		useSSL = val
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		Log.Fatal("❌ Failed to connect to MinIO: %v", err)
	}

	MinioClient = client
	Log.Println("✅ MinIO client initialized")
}

func EnsureBucket(bucketName string) error {
	ctx := context.Background()
	exists, err := MinioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}
	if !exists {
		return MinioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	}
	return nil
}

// MinioSignedURL mengembalikan URL sementara (signed) dari bucket private
func MinioUrl(objectName string, expiry ...time.Duration) (string, error) {
	bucketName := os.Getenv("MINIO_BUCKET")
	ctx := context.Background()

	reqParams := make(url.Values) // kosong, bisa diisi ?response-content-type dll

	// Gunakan default jika tidak dikirim
	finalExpiry := 7 * 24 * time.Hour // 7 hari
	if len(expiry) > 0 {
		finalExpiry = expiry[0]
	}

	url, err := MinioClient.PresignedGetObject(ctx, bucketName, objectName, finalExpiry, reqParams)
	if err != nil {
		return "", err
	}

	return url.String(), nil
}

// MinioSignedURL mengembalikan URL sementara (signed) dari bucket private
func MinioUrlExistsOrError(objectName string, expiry ...time.Duration) (string, error) {
	bucketName := os.Getenv("MINIO_BUCKET")
	ctx := context.Background()

	// Gunakan default jika tidak dikirim
	finalExpiry := 7 * 24 * time.Hour // 7 hari
	if len(expiry) > 0 {
		finalExpiry = expiry[0]
	}

	// ✅ Cek apakah file ada di MinIO
	_, err := MinioClient.StatObject(ctx, bucketName, objectName, minio.StatObjectOptions{})
	if err == nil {
		// ✅ Jika ada → generate signed URL
		return MinioUrl(objectName, finalExpiry)
	}

	return "", err
}

// MinioSignedURL mengembalikan URL sementara (signed) dari bucket private
func MinioSignedURL(bucketName, objectName string, expiry time.Duration) (string, error) {
	ctx := context.Background()

	reqParams := make(url.Values) // kosong, bisa diisi ?response-content-type dll

	url, err := MinioClient.PresignedGetObject(ctx, bucketName, objectName, expiry, reqParams)
	if err != nil {
		return "", err
	}

	return url.String(), nil
}

func UploadFileToMinio(objectName, filePath string) (string, error) {
	ctx := context.Background()
	bucketName := os.Getenv("MINIO_BUCKET")
	if bucketName == "" {
		bucketName = "default"
	}

	// Pastikan bucket tersedia
	if err := EnsureBucket(bucketName); err != nil {
		return "", err
	}

	// Deteksi content-type berdasarkan ekstensi file
	ext := filepath.Ext(filePath)
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream" // fallback default
	}

	// Upload
	_, err := MinioClient.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}

	// Kembalikan URL file
	// Setelah upload sukses
	signedURL, err := MinioSignedURL(bucketName, objectName, 1*time.Hour)
	if err != nil {
		return "", err
	}
	return signedURL, nil
}

func UploadObjectToMinio(objectName string, data []byte, contentType string) (string, error) {
	ctx := context.Background()
	bucketName := os.Getenv("MINIO_BUCKET")
	if bucketName == "" {
		bucketName = "default"
	}

	// Pastikan bucket tersedia
	if err := EnsureBucket(bucketName); err != nil {
		return "", err
	}

	reader := bytes.NewReader(data)

	if contentType == "" {
		contentType = http.DetectContentType(data)
	}

	_, err := MinioClient.PutObject(ctx, bucketName, objectName, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}

	// Kembalikan URL signed
	return MinioSignedURL(bucketName, objectName, 1*time.Hour)
}

func DownloadFileFromMinio(objectName, localDir string) (string, error) {
	ctx := context.Background()
	bucketName := os.Getenv("MINIO_BUCKET")
	if bucketName == "" {
		bucketName = "default"
	}

	object, err := MinioClient.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return "", err
	}

	localPath := filepath.Join(localDir, objectName)

	outFile, err := os.Create(localPath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	// ✅ Gunakan io.Copy untuk membaca isi object dari MinIO ke file lokal
	_, err = io.Copy(outFile, object)
	if err != nil {
		return "", err
	}

	return localPath, nil
}

type FileStreamResult struct {
	Stream      io.ReadCloser
	ContentType string
	Size        int64
	Filename    string
}

func GetFileStreamFromMinio(objectName string) (*FileStreamResult, error) {
	ctx := context.Background()
	bucketName := os.Getenv("MINIO_BUCKET")
	if bucketName == "" {
		bucketName = "default"
	}

	object, err := MinioClient.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	info, err := object.Stat()
	if err != nil {
		return nil, err
	}

	return &FileStreamResult{
		Stream:      object,
		ContentType: info.ContentType,
		Size:        info.Size,
		Filename:    filepath.Base(objectName),
	}, nil
}

// DeleteFileFromMinio menghapus file (object) dari MinIO
func DeleteFileFromMinio(objectName string) error {
	ctx := context.Background()

	bucketName := os.Getenv("MINIO_BUCKET")
	if bucketName == "" {
		bucketName = "default"
	}

	err := MinioClient.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}
