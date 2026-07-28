package port

import (
	"context"
	"io"
)

// Driver 驱动接口
type Driver interface {
	Name() string
	List(ctx context.Context, path string) ([]FileInfo, error)
	GetFile(ctx context.Context, path string) (*FileInfo, error)
	ReadFile(ctx context.Context, path string) (io.ReadCloser, error)
	WriteFile(ctx context.Context, path string, fileName string, reader io.Reader, size int64) error
	InstantUpload(ctx context.Context, path string, fileName string, sha256 string, size int64) (bool, error)
	DeleteFile(ctx context.Context, path string) error
	Mkdir(ctx context.Context, path string) error
}

// FileInfo 文件信息结构
type FileInfo struct {
	Name   string
	Path   string
	Size   int64
	IsDir  bool
	SHA256 string
}
