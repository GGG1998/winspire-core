// Package games contains the GameLibrary bounded context (merged from game-management service)
package games

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// s3API abstracts the AWS S3 client to enable testing without real AWS calls.
type s3API interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
	DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error)
	PutBucketVersioning(ctx context.Context, params *s3.PutBucketVersioningInput, optFns ...func(*s3.Options)) (*s3.PutBucketVersioningOutput, error)
	GetBucketVersioning(ctx context.Context, params *s3.GetBucketVersioningInput, optFns ...func(*s3.Options)) (*s3.GetBucketVersioningOutput, error)
	PutBucketWebsite(ctx context.Context, params *s3.PutBucketWebsiteInput, optFns ...func(*s3.Options)) (*s3.PutBucketWebsiteOutput, error)
	PutBucketCors(ctx context.Context, params *s3.PutBucketCorsInput, optFns ...func(*s3.Options)) (*s3.PutBucketCorsOutput, error)
}

// S3Client provides access to AWS S3 for game file storage.
type S3Client struct {
	client        s3API
	presignClient *s3.PresignClient
	bucket        string
	region        string
	endpoint      string
}

// BundleFile represents a static asset stored in S3.
type BundleFile struct {
	ContentType string
	Content     []byte
}

// NewS3Client creates a new S3 client with the provided credentials and optional endpoint.
func NewS3Client(region, bucket, accessKeyID, secretAccessKey, endpoint string) (*S3Client, error) {
	var cfg aws.Config
	var err error

	// If credentials are provided, use them; otherwise use default credential chain
	if accessKeyID != "" && secretAccessKey != "" {
		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				accessKeyID,
				secretAccessKey,
				"",
			)),
		)
	} else {
		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(region),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		// Allow buckets provided as access point ARNs
		o.UseARNRegion = true
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true
		}
	})

	return &S3Client{
		client:        s3Client,
		presignClient: s3.NewPresignClient(s3Client),
		bucket:        bucket,
		region:        region,
		endpoint:      endpoint,
	}, nil
}

// UploadFile uploads a single file to S3.
func (c *S3Client) UploadFile(ctx context.Context, s3Path string, data []byte, contentType string) error {
	if contentType == "" {
		contentType = detectContentType(s3Path)
	}

	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(s3Path),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("upload file to S3: %w", err)
	}

	return nil
}

// UploadMultipleFiles uploads multiple files from a multipart form to S3.
func (c *S3Client) UploadMultipleFiles(ctx context.Context, baseS3Path string, files []*multipart.FileHeader) ([]string, error) {
	uploadedPaths := make([]string, 0, len(files))

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			return uploadedPaths, fmt.Errorf("open file %s: %w", fileHeader.Filename, err)
		}

		data, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			return uploadedPaths, fmt.Errorf("read file %s: %w", fileHeader.Filename, err)
		}

		// Create S3 path: baseS3Path/filename
		s3Path := path.Join(baseS3Path, fileHeader.Filename)
		contentType := fileHeader.Header.Get("Content-Type")
		if contentType == "" {
			contentType = detectContentType(fileHeader.Filename)
		}

		if err := c.UploadFile(ctx, s3Path, data, contentType); err != nil {
			return uploadedPaths, fmt.Errorf("upload %s: %w", fileHeader.Filename, err)
		}

		uploadedPaths = append(uploadedPaths, s3Path)
	}

	return uploadedPaths, nil
}

// UploadMultipleFilesWithPath uploads files using a provided relative path instead of the filename.
// This is needed because browsers strip directory paths from filenames when uploading.
func (c *S3Client) UploadMultipleFilesWithPath(ctx context.Context, baseS3Path string, files []*multipart.FileHeader, relativePath string) ([]string, error) {
	uploadedPaths := make([]string, 0, len(files))

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			return uploadedPaths, fmt.Errorf("open file %s: %w", fileHeader.Filename, err)
		}

		data, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			return uploadedPaths, fmt.Errorf("read file %s: %w", fileHeader.Filename, err)
		}

		// Use the provided relativePath instead of fileHeader.Filename
		// This preserves directory structure (e.g., "Build/game.js")
		s3Path := path.Join(baseS3Path, relativePath)
		contentType := fileHeader.Header.Get("Content-Type")
		if contentType == "" {
			contentType = detectContentType(relativePath)
		}

		if err := c.UploadFile(ctx, s3Path, data, contentType); err != nil {
			return uploadedPaths, fmt.Errorf("upload %s: %w", relativePath, err)
		}

		uploadedPaths = append(uploadedPaths, s3Path)
	}

	return uploadedPaths, nil
}

// BundleFileStream represents a streamable bundle asset from S3.
type BundleFileStream struct {
	ContentType   string
	ContentLength int64
	Body          io.ReadCloser
}

// GetBundleFile retrieves a specific bundle asset from S3.
func (c *S3Client) GetBundleFile(ctx context.Context, basePath, relativePath string) (*BundleFile, error) {
	stream, err := c.GetBundleFileStream(ctx, basePath, relativePath)
	if err != nil {
		return nil, err
	}
	defer stream.Body.Close()

	data, err := io.ReadAll(stream.Body)
	if err != nil {
		return nil, fmt.Errorf("read bundle file from S3: %w", err)
	}

	return &BundleFile{
		ContentType: stream.ContentType,
		Content:     data,
	}, nil
}

// GetBundleFileStream retrieves a bundle asset as a stream for large files.
func (c *S3Client) GetBundleFileStream(ctx context.Context, basePath, relativePath string) (*BundleFileStream, error) {
	if basePath == "" {
		return nil, fmt.Errorf("base path is required")
	}

	trimmedBase := strings.Trim(basePath, "/")
	if trimmedBase == "" {
		return nil, fmt.Errorf("base path is required")
	}
	if strings.HasSuffix(strings.ToLower(trimmedBase), ".zip") {
		dir := path.Dir(trimmedBase)
		if dir == "." {
			return nil, fmt.Errorf("invalid storage path: %s", basePath)
		}
		trimmedBase = dir
	}
	trimmedRel := strings.Trim(relativePath, "/")
	s3Path := trimmedBase
	if trimmedRel != "" {
		s3Path = path.Join(trimmedBase, trimmedRel)
	}

	out, err := c.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(s3Path),
	})
	if err != nil {
		return nil, fmt.Errorf("get bundle file from S3: %w", err)
	}

	contentType := aws.ToString(out.ContentType)
	if contentType == "" {
		contentType = detectContentType(relativePath)
	}

	return &BundleFileStream{
		ContentType:   contentType,
		ContentLength: aws.ToInt64(out.ContentLength),
		Body:          out.Body,
	}, nil
}

// GetBundleFilePresignedURL generates a presigned URL for direct S3 access.
// This is useful for large files to avoid streaming through the backend.
func (c *S3Client) GetBundleFilePresignedURL(ctx context.Context, basePath, relativePath string, expiry time.Duration) (string, error) {
	if basePath == "" {
		return "", fmt.Errorf("base path is required")
	}

	trimmedBase := strings.Trim(basePath, "/")
	if trimmedBase == "" {
		return "", fmt.Errorf("base path is required")
	}
	if strings.HasSuffix(strings.ToLower(trimmedBase), ".zip") {
		dir := path.Dir(trimmedBase)
		if dir == "." {
			return "", fmt.Errorf("invalid storage path: %s", basePath)
		}
		trimmedBase = dir
	}
	trimmedRel := strings.Trim(relativePath, "/")
	s3Path := trimmedBase
	if trimmedRel != "" {
		s3Path = path.Join(trimmedBase, trimmedRel)
	}

	presignResult, err := c.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(s3Path),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("generate presigned URL: %w", err)
	}

	return presignResult.URL, nil
}

// DeleteFile removes a file from S3.
func (c *S3Client) DeleteFile(ctx context.Context, s3Path string) error {
	_, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(s3Path),
	})
	if err != nil {
		return fmt.Errorf("delete file from S3: %w", err)
	}

	return nil
}

// DeleteFolder removes all files with the given prefix from S3.
func (c *S3Client) DeleteFolder(ctx context.Context, folderPrefix string) error {
	// List all objects with the prefix
	listOutput, err := c.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(c.bucket),
		Prefix: aws.String(folderPrefix),
	})
	if err != nil {
		return fmt.Errorf("list objects in folder: %w", err)
	}

	if len(listOutput.Contents) == 0 {
		return nil // Nothing to delete
	}

	// Build delete request
	objectsToDelete := make([]types.ObjectIdentifier, len(listOutput.Contents))
	for i, obj := range listOutput.Contents {
		objectsToDelete[i] = types.ObjectIdentifier{
			Key: obj.Key,
		}
	}

	_, err = c.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(c.bucket),
		Delete: &types.Delete{
			Objects: objectsToDelete,
			Quiet:   aws.Bool(true),
		},
	})
	if err != nil {
		return fmt.Errorf("delete objects from S3: %w", err)
	}

	return nil
}

// GetPublicURL returns the public URL for a file in S3.
func (c *S3Client) GetPublicURL(s3Path string) string {
	if c.endpoint != "" {
		return fmt.Sprintf("%s/%s/%s", strings.TrimSuffix(c.endpoint, "/"), c.bucket, s3Path)
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", c.bucket, c.region, s3Path)
}

// EnableVersioning enables versioning on the S3 bucket.
func (c *S3Client) EnableVersioning(ctx context.Context) error {
	_, err := c.client.PutBucketVersioning(ctx, &s3.PutBucketVersioningInput{
		Bucket: aws.String(c.bucket),
		VersioningConfiguration: &types.VersioningConfiguration{
			Status: types.BucketVersioningStatusEnabled,
		},
	})
	if err != nil {
		return fmt.Errorf("enable versioning: %w", err)
	}

	return nil
}

// GetVersioningStatus returns the versioning status of the bucket.
func (c *S3Client) GetVersioningStatus(ctx context.Context) (string, error) {
	output, err := c.client.GetBucketVersioning(ctx, &s3.GetBucketVersioningInput{
		Bucket: aws.String(c.bucket),
	})
	if err != nil {
		return "", fmt.Errorf("get versioning status: %w", err)
	}

	if output.Status == "" {
		return "Disabled", nil
	}

	return string(output.Status), nil
}

// ConfigureStaticWebsite configures the bucket for static website hosting.
func (c *S3Client) ConfigureStaticWebsite(ctx context.Context, indexDocument, errorDocument string) error {
	if indexDocument == "" {
		indexDocument = "index.html"
	}
	if errorDocument == "" {
		errorDocument = "error.html"
	}

	_, err := c.client.PutBucketWebsite(ctx, &s3.PutBucketWebsiteInput{
		Bucket: aws.String(c.bucket),
		WebsiteConfiguration: &types.WebsiteConfiguration{
			IndexDocument: &types.IndexDocument{
				Suffix: aws.String(indexDocument),
			},
			ErrorDocument: &types.ErrorDocument{
				Key: aws.String(errorDocument),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("configure static website: %w", err)
	}

	return nil
}

// SetBucketCORS configures CORS for the bucket.
func (c *S3Client) SetBucketCORS(ctx context.Context) error {
	_, err := c.client.PutBucketCors(ctx, &s3.PutBucketCorsInput{
		Bucket: aws.String(c.bucket),
		CORSConfiguration: &types.CORSConfiguration{
			CORSRules: []types.CORSRule{
				{
					AllowedHeaders: []string{"*"},
					AllowedMethods: []string{"GET", "HEAD"},
					AllowedOrigins: []string{"*"},
					ExposeHeaders:  []string{"ETag"},
					MaxAgeSeconds:  aws.Int32(3600),
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("set bucket CORS: %w", err)
	}

	return nil
}

// detectContentType returns the MIME type based on file extension.
func detectContentType(filePath string) string {
	ext := strings.ToLower(path.Ext(filePath))
	switch ext {
	case ".html", ".htm":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js", ".mjs":
		return "application/javascript; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".webp":
		return "image/webp"
	case ".ico":
		return "image/x-icon"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	case ".eot":
		return "application/vnd.ms-fontobject"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".ogg":
		return "audio/ogg"
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".wasm":
		return "application/wasm"
	case ".xml":
		return "application/xml"
	case ".txt":
		return "text/plain; charset=utf-8"
	case ".zip":
		return "application/zip"
	default:
		return "application/octet-stream"
	}
}
