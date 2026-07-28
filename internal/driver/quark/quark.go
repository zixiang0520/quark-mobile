package quark

import (
	"context"
	"fmt"
	"io"

	"quark-mobile/internal/driver/openlist"
	"quark-mobile/internal/model"
	"quark-mobile/internal/port"

	"github.com/spf13/viper"
)

// 确保实现了 port.Driver 接口
var _ port.Driver = (*QuarkDriver)(nil)

type QuarkDriver struct {
	client    *openlist.Client
	mountPath string
}

func NewQuarkDriver(client *openlist.Client) *QuarkDriver {
	return &QuarkDriver{
		client:    client,
		mountPath: viper.GetString("openlist.mounts.quark"),
	}
}

func (q *QuarkDriver) Name() string {
	return string(model.DriverQuark)
}

func (q *QuarkDriver) List(ctx context.Context, path string) ([]port.FileInfo, error) {
	fullPath := q.mountPath + path
	return q.client.ListFiles(ctx, fullPath)
}

func (q *QuarkDriver) GetFile(ctx context.Context, path string) (*port.FileInfo, error) {
	fullPath := q.mountPath + path
	return q.client.GetFileInfo(ctx, fullPath)
}

func (q *QuarkDriver) ReadFile(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath := q.mountPath + path
	rawURL, err := q.client.GetDownloadURL(ctx, fullPath)
	if err != nil {
		return nil, fmt.Errorf("get download url: %w", err)
	}

	resp, err := q.client.HTTPClient.Get(rawURL)
	if err != nil {
		return nil, fmt.Errorf("download file: %w", err)
	}

	if resp.StatusCode >= 400 {
		resp.Body.Close()
		return nil, fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	return resp.Body, nil
}

func (q *QuarkDriver) WriteFile(ctx context.Context, path string, fileName string, reader io.Reader, size int64) error {
	return fmt.Errorf("WriteFile not supported via OpenList, use Copy instead")
}

func (q *QuarkDriver) InstantUpload(ctx context.Context, path string, fileName string, sha256Hash string, size int64) (bool, error) {
	// OpenList 的 Copy 接口内部会自动判断是否秒传
	// 这里直接返回 false，表示没有单独的秒传接口
	return false, nil
}

func (q *QuarkDriver) DeleteFile(ctx context.Context, path string) error {
	dir := q.mountPath
	return q.client.RemoveFile(ctx, dir, []string{path})
}

func (q *QuarkDriver) Mkdir(ctx context.Context, path string) error {
	fullPath := q.mountPath + path
	return q.client.Mkdir(ctx, fullPath)
}

// GetMountPath 获取挂载路径
func (q *QuarkDriver) GetMountPath() string {
	return q.mountPath
}

// CopyTo 复制到另一个驱动的挂载点
func (q *QuarkDriver) CopyTo(ctx context.Context, srcPath, srcName, dstPath string) error {
	srcDir := q.mountPath + srcPath
	return q.client.CopyFile(ctx, srcDir, srcName, dstPath)
}
