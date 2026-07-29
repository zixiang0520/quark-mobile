package openlist

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"quark-mobile/internal/port"
)

type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

type FileInfo struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	Size     int64     `json:"size"`
	IsDir    bool      `json:"is_dir"`
	Sign     string    `json:"sign,omitempty"`
	RawURL   string    `json:"raw_url,omitempty"`
	Modified time.Time `json:"modified,omitempty"`
}

type ListResponse struct {
	Content []FileInfo `json:"content"`
	Total   int        `json:"total"`
}

type APIResponse struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

type APIListResponse struct {
	Code    int          `json:"code"`
	Data    ListResponse `json:"data"`
	Message string       `json:"message"`
}

type APIGetResponse struct {
	Code    int      `json:"code"`
	Data    FileInfo `json:"data"`
	Message string   `json:"message"`
}

type EmptyResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewClient(baseURL string, token string) *Client {
	return &Client{
		BaseURL:    baseURL,
		Token:      token,
		HTTPClient: &http.Client{
			// 不设置全局超时，由 context 控制生命周期
			// API 调用通过 doRequest 使用 context deadline
			// 文件下载通过 ReadFile 传入更长的 context
		},
	}
}

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	return io.ReadAll(resp.Body)
}

func (c *Client) Login(ctx context.Context, username, password string) error {
	body := map[string]string{
		"username": username,
		"password": password,
	}

	respData, err := c.doRequest(ctx, "POST", "/api/auth/login", body)
	if err != nil {
		return err
	}

	var apiResp struct {
		Code    int    `json:"code"`
		Data    struct {
			Token string `json:"token"`
		} `json:"data"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(respData, &apiResp); err != nil {
		return fmt.Errorf("parse login response: %w", err)
	}

	if apiResp.Code != 200 {
		return fmt.Errorf("login failed: %s", apiResp.Message)
	}

	c.Token = apiResp.Data.Token
	return nil
}

// ListFiles 列举文件
// POST /api/fs/list
func (c *Client) ListFiles(ctx context.Context, path string) ([]port.FileInfo, error) {
	body := map[string]interface{}{
		"path":     path,
		"password": "",
	}

	respData, err := c.doRequest(ctx, "POST", "/api/fs/list", body)
	if err != nil {
		return nil, err
	}

	var listResp APIListResponse
	if err := json.Unmarshal(respData, &listResp); err != nil {
		return nil, fmt.Errorf("parse list response: %w", err)
	}

	if listResp.Code != 200 {
		return nil, fmt.Errorf("list failed: %s", listResp.Message)
	}

	result := make([]port.FileInfo, 0, len(listResp.Data.Content))
	for _, f := range listResp.Data.Content {
		result = append(result, port.FileInfo{
			Name:   f.Name,
			Path:   f.Path,
			Size:   f.Size,
			IsDir:  f.IsDir,
			SHA256: f.Sign,
		})
	}

	return result, nil
}

// GetFileInfo 获取文件详情
// POST /api/fs/get
func (c *Client) GetFileInfo(ctx context.Context, path string) (*port.FileInfo, error) {
	body := map[string]interface{}{
		"path":     path,
		"password": "",
	}

	respData, err := c.doRequest(ctx, "POST", "/api/fs/get", body)
	if err != nil {
		return nil, err
	}

	var getResp APIGetResponse
	if err := json.Unmarshal(respData, &getResp); err != nil {
		return nil, fmt.Errorf("parse get response: %w", err)
	}

	if getResp.Code != 200 {
		return nil, fmt.Errorf("get failed: %s", getResp.Message)
	}

	// 转换为 port.FileInfo
	return &port.FileInfo{
		Name:   getResp.Data.Name,
		Path:   path,
		Size:   getResp.Data.Size,
		IsDir:  getResp.Data.IsDir,
		SHA256: getResp.Data.Sign,
	}, nil
}

// GetRawFileInfo 获取原始 OpenList 文件信息（包含 RawURL）
func (c *Client) GetRawFileInfo(ctx context.Context, path string) (*FileInfo, error) {
	body := map[string]interface{}{
		"path":     path,
		"password": "",
	}

	respData, err := c.doRequest(ctx, "POST", "/api/fs/get", body)
	if err != nil {
		return nil, err
	}

	var getResp APIGetResponse
	if err := json.Unmarshal(respData, &getResp); err != nil {
		return nil, fmt.Errorf("parse get response: %w", err)
	}

	if getResp.Code != 200 {
		return nil, fmt.Errorf("get failed: %s", getResp.Message)
	}

	return &getResp.Data, nil
}

// CopyFile 跨挂载点复制（OpenList 内部自动秒传）
// POST /api/fs/copy
func (c *Client) CopyFile(ctx context.Context, srcDir, srcName, dstDir string) error {
	body := map[string]interface{}{
		"src_dir":   srcDir,
		"src_names": []string{srcName},
		"dst_dir":   dstDir,
		"dst_names": []string{srcName},
		"passwords": []string{""},
	}

	respData, err := c.doRequest(ctx, "POST", "/api/fs/copy", body)
	if err != nil {
		return err
	}

	var emptyResp EmptyResponse
	if err := json.Unmarshal(respData, &emptyResp); err != nil {
		return fmt.Errorf("parse copy response: %w", err)
	}

	if emptyResp.Code != 200 {
		return fmt.Errorf("copy failed: %s", emptyResp.Message)
	}

	return nil
}

// MoveFile 跨挂载点移动
// POST /api/fs/move
func (c *Client) MoveFile(ctx context.Context, srcDir, srcName, dstDir string) error {
	body := map[string]interface{}{
		"src_dir":   srcDir,
		"src_names": []string{srcName},
		"dst_dir":   dstDir,
		"dst_names": []string{srcName},
		"passwords": []string{""},
	}

	respData, err := c.doRequest(ctx, "POST", "/api/fs/move", body)
	if err != nil {
		return err
	}

	var emptyResp EmptyResponse
	if err := json.Unmarshal(respData, &emptyResp); err != nil {
		return fmt.Errorf("parse move response: %w", err)
	}

	if emptyResp.Code != 200 {
		return fmt.Errorf("move failed: %s", emptyResp.Message)
	}

	return nil
}

// Mkdir 创建目录
// POST /api/fs/mkdir
func (c *Client) Mkdir(ctx context.Context, path string) error {
	body := map[string]string{
		"path": path,
	}

	respData, err := c.doRequest(ctx, "POST", "/api/fs/mkdir", body)
	if err != nil {
		return err
	}

	var emptyResp EmptyResponse
	if err := json.Unmarshal(respData, &emptyResp); err != nil {
		return fmt.Errorf("parse mkdir response: %w", err)
	}

	if emptyResp.Code != 200 {
		return fmt.Errorf("mkdir failed: %s", emptyResp.Message)
	}

	return nil
}

// RemoveFile 删除文件
// POST /api/fs/remove
func (c *Client) RemoveFile(ctx context.Context, dir string, names []string) error {
	body := map[string]interface{}{
		"dir":   dir,
		"names": names,
	}

	respData, err := c.doRequest(ctx, "POST", "/api/fs/remove", body)
	if err != nil {
		return err
	}

	var emptyResp EmptyResponse
	if err := json.Unmarshal(respData, &emptyResp); err != nil {
		return fmt.Errorf("parse remove response: %w", err)
	}

	if emptyResp.Code != 200 {
		return fmt.Errorf("remove failed: %s", emptyResp.Message)
	}

	return nil
}

// GetDownloadURL 获取下载直链
func (c *Client) GetDownloadURL(ctx context.Context, path string) (string, error) {
	fileInfo, err := c.GetRawFileInfo(ctx, path)
	if err != nil {
		return "", err
	}
	if fileInfo.RawURL == "" {
		return "", fmt.Errorf("no raw_url available for file: %s", path)
	}
	return fileInfo.RawURL, nil
}
