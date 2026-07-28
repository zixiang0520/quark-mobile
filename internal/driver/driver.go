package driver

import (
	"context"
	"io"

	"quark-mobile/internal/model"
)

type Driver interface {
	Name() model.DriverType
	List(ctx context.Context, path string) ([]model.FileInfo, error)
	GetFile(ctx context.Context, path string) (*model.FileInfo, error)
	ReadFile(ctx context.Context, path string) (io.ReadCloser, error)
	WriteFile(ctx context.Context, path string, fileName string, reader io.Reader, size int64) error
	InstantUpload(ctx context.Context, path string, fileName string, sha256 string, size int64) (bool, error)
	DeleteFile(ctx context.Context, path string) error
	Mkdir(ctx context.Context, path string) error
}
