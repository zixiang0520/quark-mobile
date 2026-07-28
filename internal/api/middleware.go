package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"quark-mobile/internal/config"
)

// AuthMiddleware 验证会话中间件
func AuthMiddleware(cfgMgr *config.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 session cookie 或 header
		sessionID := c.GetHeader("X-Session-ID")
		if sessionID == "" {
			// 从 cookie 获取
			if cookie, err := c.Cookie("session_id"); err == nil {
				sessionID = cookie
			}
		}

		// 验证会话
		if sessionID == "" || !cfgMgr.ValidateSession(sessionID) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		// 将 sessionID 存入 context
		c.Set("session_id", sessionID)
		c.Next()
	}
}

// CORSMiddleware CORS 中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			// 允许所有 origin（本地开发用）
			origin = "*"
		}

		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Session-ID")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// SecurityHeadersMiddleware 安全头中间件
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置安全相关的 HTTP 头
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// HTTPS 强制（如果请求是通过 HTTPS）
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}

// RateLimitMiddleware 简单的速率限制中间件（防止暴力破解）
func RateLimitMiddleware(maxRequests int) gin.HandlerFunc {
	requests := make(map[string]int)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := ip + ":" + c.Request.URL.Path

		if count, exists := requests[key]; exists {
			if count >= maxRequests {
				c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many requests, please try again later"})
				c.Abort()
				return
			}
			requests[key] = count + 1
		} else {
			requests[key] = 1
		}

		c.Next()

		// 清理过期的记录（简单实现，生产环境建议使用时间窗口）
		if len(requests) > 1000 {
			for k := range requests {
				if strings.Count(k, ":") > 0 {
					delete(requests, k)
				}
			}
		}
	}
}
