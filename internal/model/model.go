package model

import "time"

type DriverType string

const (
	DriverQuark  DriverType = "quark"
	DriverMobile DriverType = "mobile"
)

type FileInfo struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	Size     int64     `json:"size"`
	IsDir    bool      `json:"is_dir"`
	SHA256   string    `json:"sha256,omitempty"`
	Modified time.Time `json:"modified"`
}

type TransferRequest struct {
	SourceDriver DriverType `json:"source_driver" binding:"required"`
	SourcePath   string     `json:"source_path" binding:"required"`
	TargetDriver DriverType `json:"target_driver" binding:"required"`
	TargetPath   string     `json:"target_path" binding:"required"`
	FileName     string     `json:"file_name,omitempty"`
}

type TransferTask struct {
	ID         string     `json:"id"`
	SourceDriver DriverType `json:"source_driver"`
	SourcePath   string     `json:"source_path"`
	TargetDriver DriverType `json:"target_driver"`
	TargetPath   string     `json:"target_path"`
	FileName     string     `json:"file_name"`
	FileSize     int64      `json:"file_size"`
	SHA256       string     `json:"sha256"`
	Status       string     `json:"status"`
	Progress     float64    `json:"progress"`
	InstantDone  bool       `json:"instant_done"`
	Error        string     `json:"error,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
