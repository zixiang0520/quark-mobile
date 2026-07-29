package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config 应用配置结构
type Config struct {
	Server   ServerConfig   `json:"server" yaml:"server"`
	OpenList OpenListConfig `json:"openlist" yaml:"openlist"`
	Transfer TransferConfig `json:"transfer" yaml:"transfer"`
}

type ServerConfig struct {
	Port int    `json:"port" yaml:"port"`
	Mode string `json:"mode" yaml:"mode"`
}

type OpenListConfig struct {
	BaseURL  string `json:"base_url" yaml:"base_url"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	Mounts   struct {
		Quark  string `json:"quark" yaml:"quark"`
		Mobile string `json:"mobile" yaml:"mobile"`
	} `json:"mounts" yaml:"mounts"`
}

type TransferConfig struct {
	MaxConcurrent int    `json:"max_concurrent" yaml:"max_concurrent"`
	CacheDir      string `json:"cache_dir" yaml:"cache_dir"`
	Timeout       int    `json:"timeout" yaml:"timeout"`
}

// Session 会话管理
type Session struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Manager 配置管理器
type Manager struct {
	mu          sync.RWMutex
	config      *Config
	filePath    string
	writePath   string // 实际写入路径（可能与读取路径不同）
	adminPass   string // bcrypt hash
	sessions    map[string]*Session
	encryptionKey []byte
}

var defaultConfig = &Config{
	Server: ServerConfig{
		Port: 18900,
		Mode: "release",
	},
	OpenList: OpenListConfig{
		BaseURL:  "http://localhost:5244",
		Username: "admin",
		Password: "",
	},
	Transfer: TransferConfig{
		MaxConcurrent: 4,
		CacheDir:      "data/cache",
		Timeout:       60,
	},
}

func init() {
	defaultConfig.OpenList.Mounts.Quark = "/quark"
	defaultConfig.OpenList.Mounts.Mobile = "/mobile"
}

// NewManager 创建配置管理器
func NewManager(configPath string) *Manager {
	m := &Manager{
		filePath:  configPath,
		writePath: configPath,
		sessions:  make(map[string]*Session),
	}

	// 检查配置文件路径是否可写，否则使用 data 目录
	if !m.canWrite(configPath) {
		m.writePath = "data/config.yaml"
		// 如果原始配置文件存在，复制它
		if data, err := os.ReadFile(configPath); err == nil {
			os.MkdirAll("data", 0755)
			os.WriteFile(m.writePath, data, 0600)
		}
	}

	// 生成或加载加密密钥
	m.encryptionKey = m.loadOrCreateEncryptionKey()

	// 加载配置
	m.config = m.loadConfig()

	// 加载或创建管理员密码
	m.adminPass = m.loadOrCreateAdminPassword()

	return m
}

// canWrite 检查路径是否可写
func (m *Manager) canWrite(path string) bool {
	dir := filepath.Dir(path)
	info, err := os.Stat(dir)
	if err != nil {
		return false
	}
	// 检查目录权限
	if info.Mode().Perm()&0200 == 0 {
		return false
	}
	// 尝试创建临时文件测试
	tmpFile := filepath.Join(dir, ".write_test")
	if err := os.WriteFile(tmpFile, []byte("test"), 0600); err != nil {
		return false
	}
	os.Remove(tmpFile)
	return true
}

// loadOrCreateEncryptionKey 加载或创建加密密钥
func (m *Manager) loadOrCreateEncryptionKey() []byte {
	// 使用 data 目录存储密钥（有写入权限）
	keyDir := "data"
	
	// 确保目录存在
	os.MkdirAll(keyDir, 0755)
	
	keyFile := filepath.Join(keyDir, ".encryption_key")

	data, err := os.ReadFile(keyFile)
	if err == nil && len(data) == 32 {
		return data
	}

	// 生成新密钥
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		panic(fmt.Sprintf("failed to generate encryption key: %v", err))
	}

	if err := os.WriteFile(keyFile, key, 0600); err != nil {
		panic(fmt.Sprintf("failed to save encryption key: %v", err))
	}

	return key
}

// loadOrCreateAdminPassword 加载或创建管理员密码
func (m *Manager) loadOrCreateAdminPassword() string {
	// 使用 data 目录存储密码（有写入权限）
	passDir := "data"
	
	// 确保目录存在
	os.MkdirAll(passDir, 0755)
	
	passFile := filepath.Join(passDir, ".admin_pass")
	
	data, err := os.ReadFile(passFile)
	if err == nil {
		return string(data)
	}

	// 首次运行，使用默认密码 admin123
	// 实际使用时应通过页面修改
	defaultPass := "admin123"
	hashedPass := hashPassword(defaultPass)

	if err := os.WriteFile(passFile, []byte(hashedPass), 0600); err != nil {
		panic(fmt.Sprintf("failed to save admin password: %v", err))
	}

	return hashedPass
}

// loadConfig 加载配置（使用 viper 支持 YAML/JSON 等多种格式）
func (m *Manager) loadConfig() *Config {
	v := viper.New()
	v.SetConfigFile(m.filePath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return defaultConfig
	}

	cfg := &Config{
		Server: ServerConfig{
			Port: v.GetInt("server.port"),
			Mode: v.GetString("server.mode"),
		},
		OpenList: OpenListConfig{
			BaseURL:  v.GetString("openlist.base_url"),
			Username: v.GetString("openlist.username"),
			Password: v.GetString("openlist.password"),
		},
		Transfer: TransferConfig{
			MaxConcurrent: v.GetInt("transfer.max_concurrent"),
			CacheDir:      v.GetString("transfer.cache_dir"),
			Timeout:       v.GetInt("transfer.timeout"),
		},
	}

	cfg.OpenList.Mounts.Quark = v.GetString("openlist.mounts.quark")
	cfg.OpenList.Mounts.Mobile = v.GetString("openlist.mounts.mobile")

	// 如果配置为空，使用默认值
	if cfg.Server.Port == 0 {
		cfg.Server.Port = defaultConfig.Server.Port
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = defaultConfig.Server.Mode
	}
	if cfg.OpenList.BaseURL == "" {
		cfg.OpenList.BaseURL = defaultConfig.OpenList.BaseURL
	}
	if cfg.Transfer.MaxConcurrent == 0 {
		cfg.Transfer.MaxConcurrent = defaultConfig.Transfer.MaxConcurrent
	}
	if cfg.Transfer.CacheDir == "" {
		cfg.Transfer.CacheDir = defaultConfig.Transfer.CacheDir
	}
	if cfg.Transfer.Timeout == 0 {
		cfg.Transfer.Timeout = defaultConfig.Transfer.Timeout
	}

	return cfg
}

// GetConfig 获取当前配置（密码脱敏）
func (m *Manager) GetConfig() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cfg := m.config
	return map[string]interface{}{
		"server": map[string]interface{}{
			"port": cfg.Server.Port,
			"mode": cfg.Server.Mode,
		},
		"openlist": map[string]interface{}{
			"base_url":  cfg.OpenList.BaseURL,
			"username":  cfg.OpenList.Username,
			"password":  maskPassword(cfg.OpenList.Password),
			"mounts": map[string]interface{}{
				"quark":  cfg.OpenList.Mounts.Quark,
				"mobile": cfg.OpenList.Mounts.Mobile,
			},
		},
		"transfer": map[string]interface{}{
			"max_concurrent": cfg.Transfer.MaxConcurrent,
			"cache_dir":      cfg.Transfer.CacheDir,
			"timeout":        cfg.Transfer.Timeout,
		},
	}
}

// GetOpenListConfig 获取 OpenList 配置（含解密后的密码）
func (m *Manager) GetOpenListConfig() (OpenListConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	password := m.config.OpenList.Password
	if password != "" {
		decrypted, err := m.decrypt(password)
		if err != nil {
			// 解密失败，说明密码是明文存储的，直接使用
			// 这是向后兼容的处理方式
		} else {
			password = decrypted
		}
	}

	cfg := m.config.OpenList
	cfg.Password = password
	return cfg, nil
}

// SaveConfig 保存配置
func (m *Manager) SaveConfig(data map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 更新 OpenList 配置
	if ol, ok := data["openlist"].(map[string]interface{}); ok {
		if baseURL, ok := ol["base_url"].(string); ok {
			m.config.OpenList.BaseURL = baseURL
		}
		if username, ok := ol["username"].(string); ok {
			m.config.OpenList.Username = username
		}
		if password, ok := ol["password"].(string); ok && password != "" && password != "******" {
			// 加密存储密码
			encrypted, err := m.encrypt(password)
			if err != nil {
				return fmt.Errorf("encrypt password: %w", err)
			}
			m.config.OpenList.Password = encrypted
		}
		if mounts, ok := ol["mounts"].(map[string]interface{}); ok {
			if quark, ok := mounts["quark"].(string); ok {
				m.config.OpenList.Mounts.Quark = quark
			}
			if mobile, ok := mounts["mobile"].(string); ok {
				m.config.OpenList.Mounts.Mobile = mobile
			}
		}
	}

	// 保存到文件
	return m.saveToFile()
}

// saveToFile 保存配置到文件（YAML 格式）
func (m *Manager) saveToFile() error {
	data, err := yaml.Marshal(m.config)
	if err != nil {
		return err
	}
	// 使用 writePath 保存配置
	dir := filepath.Dir(m.writePath)
	os.MkdirAll(dir, 0755)
	return os.WriteFile(m.writePath, data, 0600)
}

// encrypt 加密数据
func (m *Manager) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(m.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt 解密数据
func (m *Manager) decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(m.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, encrypted := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// CreateSession 创建会话
func (m *Manager) CreateSession() string {
	m.mu.Lock()
	defer m.mu.Unlock()

	sessionID := generateSessionID()
	m.sessions[sessionID] = &Session{
		ID:        sessionID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	return sessionID
}

// ValidateSession 验证会话
func (m *Manager) ValidateSession(sessionID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return false
	}

	if time.Now().After(session.ExpiresAt) {
		delete(m.sessions, sessionID)
		return false
	}

	return true
}

// DestroySession 销毁会话
func (m *Manager) DestroySession(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, sessionID)
}

// VerifyPassword 验证管理员密码
func (m *Manager) VerifyPassword(password string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return verifyPassword(password, m.adminPass)
}

// ChangePassword 修改管理员密码
func (m *Manager) ChangePassword(oldPass, newPass string) error {
	if !m.VerifyPassword(oldPass) {
		return fmt.Errorf("old password is incorrect")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	hashedPass := hashPassword(newPass)
	m.adminPass = hashedPass

	// 使用 data 目录存储密码
	passFile := filepath.Join("data", ".admin_pass")
	return os.WriteFile(passFile, []byte(hashedPass), 0600)
}

// maskPassword 脱敏密码
func maskPassword(password string) string {
	if password == "" {
		return ""
	}
	return "******"
}

// generateSessionID 生成会话ID
func generateSessionID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}