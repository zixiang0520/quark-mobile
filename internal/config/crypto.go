package config

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// hashPassword 使用 SHA-256 哈希密码
func hashPassword(password string) string {
	hash := sha256.New()
	hash.Write([]byte(password))
	return hex.EncodeToString(hash.Sum(nil))
}

// verifyPassword 验证密码
func verifyPassword(password, hash string) bool {
	computed := hashPassword(password)
	return computed == hash
}

// GenerateToken 生成简单的 token
func GenerateToken(data string) string {
	hash := sha256.New()
	hash.Write([]byte(data))
	hash.Write([]byte(fmt.Sprintf("%d", 0)))
	return hex.EncodeToString(hash.Sum(nil))
}
