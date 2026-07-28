package mobile

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
var _ port.Driver = (*MobileDriver)(nil)

type MobileDriver struct {
	client    *openlist.Client
	mountPath string
}

func NewMobileDriver(client *openlist.Client) *MobileDriver {
	return &MobileDriver{
		client:    client,
		mountPath: viper.GetString("openlist.mounts.mobile"),
	}
}

func (m *MobileDriver) Name() string {
	return string(model.DriverMobile)
}

func (m *MobileDriver) List(ctx context.Context, path string) ([]port.FileInfo, error) {
	fullPath := m.mountPath + path
	return m.client.ListFiles(ctx, fullPath)
}

func (m *MobileDriver) GetFile(ctx context.Context, path string) (*port.FileInfo, error) {
	fullPath := m.mountPath + path
	fileInfo, err := m.client.GetFileInfo(ctx, fullPath)
	if err != nil {
		return nil, err
	}

	return &port.FileInfo{
		Name:   fileInfo.Name,
		Path:   fileInfo.Path,
		Size:   fileInfo.Size,
		IsDir:  fileInfo.IsDir,
		SHA256: fileInfo.Sign,
	}
}

func (m *MobileDriver) ReadFile(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath := m.mountPath + path
	rawURL, err := m.client.GetDownloadURL(ctx, fullPath)
	if err != nil {
		return nil, fmt.Errorf("get download url: %w", err)
	}

	resp, err := m.client.HTTPClient.Get(rawURL)
	if err != nil {
		return nil, fmt.Errorf("download file: %w", err)
	}

	if resp.StatusCode >= 400 {
		resp.Body.Close()
		return nil, fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	return resp.Body, nil
}

func (m *MobileDriver) WriteFile(ctx context.Context, path string, fileName string, reader io.Reader, size int64) error {
	return fmt.Errorf("WriteFile not supported via OpenList, use Copy instead")
}

func (m *MobileDriver) InstantUpload(ctx context.Context, path string, fileName string, sha256Hash string, size int64) (bool, error) {
	return false, nil
}

func (m *MobileDriver) DeleteFile(ctx context.Context, path string) error {
	dir := m.mountPath
	return m.client.RemoveFile(ctx, dir, []string{path})
}

func (m *MobileDriver) Mkdir(ctx context.Context, path string) error {
	fullPath := m.mountPath + path
	return m.client.Mkdir(ctx, fullPath)
}

func (m *MobileDriver) GetMountPath() string {
	return m.mountPath
}

func (m *MobileDriver) CopyTo(ctx context.Context, srcPath, srcName, dstPath string) error {
	srcDir := m.mountPath + srcPath
	return m.client.CopyFile(ctx, srcDir, srcName, dstPath)
}
