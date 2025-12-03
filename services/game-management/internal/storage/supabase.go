package storage

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client provides access to Supabase Storage for game bundles.
type Client struct {
	baseURL    string
	serviceKey string
	bucket     string
	httpClient *http.Client
	redis      *redis.Client
	cacheTTL   time.Duration
}

// NewClient creates a new Supabase Storage client.
func NewClient(baseURL, serviceKey, bucket string, redisClient *redis.Client, cacheTTL time.Duration) *Client {
	return &Client{
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		serviceKey: serviceKey,
		bucket:     bucket,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		redis:      redisClient,
		cacheTTL:   cacheTTL,
	}
}

// DownloadGame downloads a game ZIP file from Supabase Storage.
func (c *Client) DownloadGame(ctx context.Context, storagePath string) ([]byte, error) {
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", c.baseURL, c.bucket, storagePath)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.serviceKey)
	req.Header.Set("apikey", c.serviceKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download game: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("download failed with status %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return data, nil
}

// UploadGame uploads a game ZIP file to Supabase Storage.
func (c *Client) UploadGame(ctx context.Context, storagePath string, data []byte) error {
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", c.baseURL, c.bucket, storagePath)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.serviceKey)
	req.Header.Set("apikey", c.serviceKey)
	req.Header.Set("Content-Type", "application/zip")
	req.Header.Set("x-upsert", "true")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("upload game: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// DeleteGame removes a game ZIP file from Supabase Storage.
func (c *Client) DeleteGame(ctx context.Context, storagePath string) error {
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", c.baseURL, c.bucket, storagePath)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.serviceKey)
	req.Header.Set("apikey", c.serviceKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete game: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// BundleFile represents a file extracted from a game bundle.
type BundleFile struct {
	Name        string
	Content     []byte
	ContentType string
}

// GetBundleFile retrieves a specific file from a game bundle, using cache when available.
func (c *Client) GetBundleFile(ctx context.Context, gameID, storagePath, filePath string) (*BundleFile, error) {
	cacheKey := fmt.Sprintf("game:%s:file:%s", gameID, filePath)

	// Try cache first
	if c.redis != nil {
		cached, err := c.redis.Get(ctx, cacheKey).Bytes()
		if err == nil && len(cached) > 0 {
			return &BundleFile{
				Name:        path.Base(filePath),
				Content:     cached,
				ContentType: detectContentType(filePath),
			}, nil
		}
	}

	// Download and extract from ZIP
	zipData, err := c.DownloadGame(ctx, storagePath)
	if err != nil {
		return nil, fmt.Errorf("download bundle: %w", err)
	}

	file, err := extractFileFromZip(zipData, filePath)
	if err != nil {
		return nil, err
	}

	// Cache the file
	if c.redis != nil {
		_ = c.redis.Set(ctx, cacheKey, file.Content, c.cacheTTL).Err()
	}

	return file, nil
}

// ExtractBundle extracts all files from a game bundle and caches them.
func (c *Client) ExtractBundle(ctx context.Context, gameID, storagePath string) (map[string]*BundleFile, error) {
	zipData, err := c.DownloadGame(ctx, storagePath)
	if err != nil {
		return nil, fmt.Errorf("download bundle: %w", err)
	}

	files, err := extractAllFromZip(zipData)
	if err != nil {
		return nil, err
	}

	// Cache all files
	if c.redis != nil {
		pipe := c.redis.Pipeline()
		for filePath, file := range files {
			cacheKey := fmt.Sprintf("game:%s:file:%s", gameID, filePath)
			pipe.Set(ctx, cacheKey, file.Content, c.cacheTTL)
		}
		_, _ = pipe.Exec(ctx)
	}

	return files, nil
}

// InvalidateCache removes all cached files for a game.
func (c *Client) InvalidateCache(ctx context.Context, gameID string) error {
	if c.redis == nil {
		return nil
	}

	pattern := fmt.Sprintf("game:%s:file:*", gameID)
	iter := c.redis.Scan(ctx, 0, pattern, 0).Iterator()

	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("scan keys: %w", err)
	}

	if len(keys) > 0 {
		if err := c.redis.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("delete keys: %w", err)
		}
	}

	return nil
}

func extractFileFromZip(zipData []byte, filePath string) (*BundleFile, error) {
	reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, fmt.Errorf("open zip: %w", err)
	}

	// Normalize the file path (remove leading slash)
	filePath = strings.TrimPrefix(filePath, "/")

	for _, file := range reader.File {
		// Handle both exact match and nested paths
		normalizedName := strings.TrimPrefix(file.Name, "/")
		if normalizedName == filePath || strings.HasSuffix(normalizedName, "/"+filePath) {
			if file.FileInfo().IsDir() {
				continue
			}

			rc, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("open file in zip: %w", err)
			}
			defer rc.Close()

			content, err := io.ReadAll(rc)
			if err != nil {
				return nil, fmt.Errorf("read file from zip: %w", err)
			}

			return &BundleFile{
				Name:        path.Base(filePath),
				Content:     content,
				ContentType: detectContentType(filePath),
			}, nil
		}
	}

	return nil, fmt.Errorf("file not found in bundle: %s", filePath)
}

func extractAllFromZip(zipData []byte) (map[string]*BundleFile, error) {
	reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, fmt.Errorf("open zip: %w", err)
	}

	files := make(map[string]*BundleFile)

	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}

		rc, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("open file %s in zip: %w", file.Name, err)
		}

		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, fmt.Errorf("read file %s from zip: %w", file.Name, err)
		}

		normalizedName := strings.TrimPrefix(file.Name, "/")
		files[normalizedName] = &BundleFile{
			Name:        path.Base(file.Name),
			Content:     content,
			ContentType: detectContentType(file.Name),
		}
	}

	return files, nil
}

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
	default:
		return "application/octet-stream"
	}
}

