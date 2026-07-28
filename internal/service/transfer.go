package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
	"quark-mobile/internal/driver"
	"quark-mobile/internal/model"
	"quark-mobile/internal/port"
)

type TransferService struct{}

func NewTransferService() *TransferService {
	return &TransferService{}
}

// CalcSHA256File 计算文件 SHA256（从本地路径）
func (s *TransferService) CalcSHA256File(path string) (string, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return "", 0, err
	}

	h := sha256.New()
	buf := make([]byte, 8*1024*1024)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			if _, werr := h.Write(buf[:n]); werr != nil {
				return "", 0, werr
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", 0, err
		}
	}

	return hex.EncodeToString(h.Sum(nil)), info.Size(), nil
}

// CalcSHA256Reader 从 io.Reader 计算 SHA256
func (s *TransferService) CalcSHA256Reader(reader io.Reader) (string, error) {
	h := sha256.New()
	buf := make([]byte, 8*1024*1024)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			if _, werr := h.Write(buf[:n]); werr != nil {
				return "", werr
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// ExecuteTransfer 执行传输任务（优先秒传，失败回退到下载上传）
func (s *TransferService) ExecuteTransfer(ctx context.Context, task *model.TransferTask) error {
	srcDrv, err := driver.GetDriver(task.SourceDriver)
	if err != nil {
		return fmt.Errorf("source driver: %w", err)
	}

	tgtDrv, err := driver.GetDriver(task.TargetDriver)
	if err != nil {
		return fmt.Errorf("target driver: %w", err)
	}

	// 1. 获取源文件信息（OpenList 的 Get 接口返回的 sign 就是 SHA256）
	fileInfo, err := srcDrv.GetFile(ctx, task.SourcePath)
	if err != nil {
		return fmt.Errorf("get source file: %w", err)
	}

	task.FileSize = fileInfo.Size
	task.FileName = fileInfo.Name
	task.SHA256 = fileInfo.SHA256

	// 2. 检查 OpenList 客户端是否可用
	olClient := driver.GetOpenListClient()
	if olClient == nil {
		return fmt.Errorf("OpenList client not initialized")
	}

	// 3. 检查源和目标是否在同一个 OpenList 实例上
	//    如果是，直接用 OpenList Copy 接口（内部自动秒传）
	srcMountPath := getMountPath(task.SourceDriver)
	tgtMountPath := getMountPath(task.TargetDriver)

	if srcMountPath != "" && tgtMountPath != "" {
		// 源目录 = 挂载路径 + 源路径
		srcDir := srcMountPath + getDirFromPath(task.SourcePath)
		tgtDir := tgtMountPath + task.TargetPath

		// 确保目标目录存在
		_ = tgtDrv.Mkdir(ctx, task.TargetPath)

		// 使用 OpenList Copy 接口（内部自动判断秒传）
		if err := olClient.CopyFile(ctx, srcDir, task.FileName, tgtDir); err == nil {
			task.InstantDone = true
			task.Status = "completed"
			task.Progress = 100
			return nil
		}
		// Copy 失败，回退到下载上传方式
	}

	// 4. 回退方案：下载 → 上传（用于不同 OpenList 实例或不支持 copy 的场景）
	return s.downloadAndUpload(ctx, task, srcDrv, tgtDrv)
}

// downloadAndUpload 下载源文件并上传到目标
func (s *TransferService) downloadAndUpload(
	ctx context.Context,
	task *model.TransferTask,
	srcDrv driver.Driver,
	tgtDrv driver.Driver,
) error {
	reader, err := srcDrv.ReadFile(ctx, task.SourcePath)
	if err != nil {
		return fmt.Errorf("read source: %w", err)
	}
	defer reader.Close()

	// 创建临时缓存文件
	cacheDir := viper.GetString("transfer.cache_dir")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}

	cachePath := filepath.Join(cacheDir, task.ID+".tmp")
	cacheFile, err := os.Create(cachePath)
	if err != nil {
		return fmt.Errorf("create cache file: %w", err)
	}
	defer cacheFile.Close()
	defer os.Remove(cachePath)

	// 边下载边计算 SHA256
	h := sha256.New()
	multiWriter := io.MultiWriter(h, cacheFile)

	written, err := io.Copy(multiWriter, reader)
	if err != nil {
		return fmt.Errorf("download & hash: %w", err)
	}

	task.SHA256 = hex.EncodeToString(h.Sum(nil))
	task.FileSize = written

	// 重新打开缓存文件用于上传
	cacheFile2, err := os.Open(cachePath)
	if err != nil {
		return fmt.Errorf("reopen cache: %w", err)
	}
	defer cacheFile2.Close()

	// 上传到目标
	if err := tgtDrv.WriteFile(ctx, task.TargetPath, task.FileName, cacheFile2, task.FileSize); err != nil {
		return fmt.Errorf("upload: %w", err)
	}

	task.Status = "completed"
	task.Progress = 100
	return nil
}

// GetFileList 获取网盘文件列表
func (s *TransferService) GetFileList(ctx context.Context, driverType model.DriverType, path string) ([]port.FileInfo, error) {
	drv, err := driver.GetDriver(driverType)
	if err != nil {
		return nil, err
	}
	return drv.List(ctx, path)
}

// EnsureDir 确保目标目录存在
func (s *TransferService) EnsureDir(ctx context.Context, driverType model.DriverType, path string) error {
	drv, err := driver.GetDriver(driverType)
	if err != nil {
		return err
	}
	return drv.Mkdir(ctx, path)
}

// NewTask 创建新任务
func (s *TransferService) NewTask(req model.TransferRequest) *model.TransferTask {
	now := time.Now()
	return &model.TransferTask{
		ID:           fmt.Sprintf("task_%d", now.UnixNano()),
		SourceDriver: req.SourceDriver,
		SourcePath:   req.SourcePath,
		TargetDriver: req.TargetDriver,
		TargetPath:   req.TargetPath,
		FileName:     req.FileName,
		Status:       "pending",
		Progress:     0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// getMountPath 获取驱动的挂载路径
func getMountPath(driverType model.DriverType) string {
	switch driverType {
	case model.DriverQuark:
		return viper.GetString("openlist.mounts.quark")
	case model.DriverMobile:
		return viper.GetString("openlist.mounts.mobile")
	default:
		return ""
	}
}

// getDirFromPath 从完整路径提取目录部分
func getDirFromPath(path string) string {
	// 如果路径以 / 结尾或没有文件名，直接返回
	if len(path) == 0 || path == "/" {
		return ""
	}
	// 找到最后一个 / 的位置
	lastSlash := len(path) - 1
	for lastSlash >= 0 && path[lastSlash] != '/' {
		lastSlash--
	}
	if lastSlash == 0 {
		return ""
	}
	return path[:lastSlash]
}
