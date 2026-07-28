package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"quark-mobile/internal/config"
	"quark-mobile/internal/driver/openlist"
)

// SettingsHandler 配置处理器
type SettingsHandler struct {
	cfgMgr *config.Manager
}

// NewSettingsHandler 创建配置处理器
func NewSettingsHandler(cfgMgr *config.Manager) *SettingsHandler {
	return &SettingsHandler{cfgMgr: cfgMgr}
}

// Login 管理员登录
// POST /api/login
func (h *SettingsHandler) Login(c *gin.Context) {
	var req struct {
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password is required"})
		return
	}

	// 验证密码
	if !h.cfgMgr.VerifyPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "incorrect password"})
		return
	}

	// 创建会话
	sessionID := h.cfgMgr.CreateSession()

	// 设置 cookie
	c.SetCookie("session_id", sessionID, 86400, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"session_id": sessionID,
		"message":    "login successful",
	})
}

// Logout 登出
// POST /api/logout
func (h *SettingsHandler) Logout(c *gin.Context) {
	sessionID := c.GetString("session_id")
	if sessionID != "" {
		h.cfgMgr.DestroySession(sessionID)
	}

	c.SetCookie("session_id", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}

// GetSettings 获取配置
// GET /api/settings
func (h *SettingsHandler) GetSettings(c *gin.Context) {
	c.JSON(http.StatusOK, h.cfgMgr.GetConfig())
}

// SaveSettings 保存配置
// POST /api/settings
func (h *SettingsHandler) SaveSettings(c *gin.Context) {
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.cfgMgr.SaveConfig(data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "settings saved successfully"})
}

// TestConnection 测试 OpenList 连接
// POST /api/settings/test
func (h *SettingsHandler) TestConnection(c *gin.Context) {
	var req struct {
		BaseURL  string `json:"base_url"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.BaseURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "base_url is required"})
		return
	}

	if req.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	// 创建 OpenList 客户端并测试连接
	client := openlist.NewClient(req.BaseURL, "")
	ctx := context.Background()

	if req.Password != "" {
		if err := client.Login(ctx, req.Username, req.Password); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"connected": false,
				"error":     err.Error(),
			})
			return
		}
	}

	// 测试列出根目录
	files, err := client.ListFiles(ctx, "/")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"connected": false,
			"error":     "connection failed: " + err.Error(),
		})
		return
	}

	// 获取可用的挂载点
	mounts := make([]string, 0)
	for _, f := range files {
		if f.IsDir {
			mounts = append(mounts, f.Name)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"connected":   true,
		"message":     "connection successful",
		"mount_points": mounts,
		"files_count": len(files),
	})
}

// ChangePassword 修改管理员密码
// POST /api/settings/password
func (h *SettingsHandler) ChangePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "both old and new password are required"})
		return
	}

	if len(req.NewPassword) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "new password must be at least 6 characters"})
		return
	}

	if err := h.cfgMgr.ChangePassword(req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
}
