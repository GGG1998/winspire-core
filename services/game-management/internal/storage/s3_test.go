package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type mockS3API struct {
	getObjectFn func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

func (m *mockS3API) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return nil, errors.New("not implemented")
}

func (m *mockS3API) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	if m.getObjectFn != nil {
		return m.getObjectFn(ctx, params, optFns...)
	}
	return nil, errors.New("getObjectFn not set")
}

func (m *mockS3API) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	return nil, errors.New("not implemented")
}

func (m *mockS3API) ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	return nil, errors.New("not implemented")
}

func (m *mockS3API) DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	return nil, errors.New("not implemented")
}

func (m *mockS3API) PutBucketVersioning(ctx context.Context, params *s3.PutBucketVersioningInput, optFns ...func(*s3.Options)) (*s3.PutBucketVersioningOutput, error) {
	return nil, errors.New("not implemented")
}

func (m *mockS3API) GetBucketVersioning(ctx context.Context, params *s3.GetBucketVersioningInput, optFns ...func(*s3.Options)) (*s3.GetBucketVersioningOutput, error) {
	return nil, errors.New("not implemented")
}

func (m *mockS3API) PutBucketWebsite(ctx context.Context, params *s3.PutBucketWebsiteInput, optFns ...func(*s3.Options)) (*s3.PutBucketWebsiteOutput, error) {
	return nil, errors.New("not implemented")
}

func (m *mockS3API) PutBucketCors(ctx context.Context, params *s3.PutBucketCorsInput, optFns ...func(*s3.Options)) (*s3.PutBucketCorsOutput, error) {
	return nil, errors.New("not implemented")
}

func TestGetBundleFileReturnsFile(t *testing.T) {
	mockClient := &mockS3API{}
	var capturedKey string
	mockClient.getObjectFn = func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
		capturedKey = aws.ToString(params.Key)
		return &s3.GetObjectOutput{
			ContentType: aws.String("application/javascript"),
			Body:        io.NopCloser(bytes.NewReader([]byte("console.log('hi')"))),
		}, nil
	}

	s3Client := &S3Client{
		client: mockClient,
		bucket: "test-bucket",
	}

	file, err := s3Client.GetBundleFile(context.Background(), "games/test", "assets/app.js")
	if err != nil {
		t.Fatalf("GetBundleFile returned error: %v", err)
	}

	if capturedKey != "games/test/assets/app.js" {
		t.Fatalf("expected key games/test/assets/app.js, got %s", capturedKey)
	}

	if file.ContentType != "application/javascript" {
		t.Fatalf("unexpected content type %s", file.ContentType)
	}

	if string(file.Content) != "console.log('hi')" {
		t.Fatalf("unexpected content %s", string(file.Content))
	}
}

func TestGetBundleFileFallsBackToDetectedContentType(t *testing.T) {
	mockClient := &mockS3API{
		getObjectFn: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
			return &s3.GetObjectOutput{
				Body: io.NopCloser(bytes.NewReader([]byte("<html></html>"))),
			}, nil
		},
	}

	s3Client := &S3Client{
		client: mockClient,
		bucket: "test-bucket",
	}

	file, err := s3Client.GetBundleFile(context.Background(), "games/test", "index.html")
	if err != nil {
		t.Fatalf("GetBundleFile returned error: %v", err)
	}

	if file.ContentType != "text/html; charset=utf-8" {
		t.Fatalf("expected fallback content type, got %s", file.ContentType)
	}
}

func TestGetBundleFileRequiresBasePath(t *testing.T) {
	s3Client := &S3Client{
		client: &mockS3API{},
	}

	if _, err := s3Client.GetBundleFile(context.Background(), "", "index.html"); err == nil {
		t.Fatal("expected error when base path is empty")
	}
}
