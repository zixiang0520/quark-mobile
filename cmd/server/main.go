package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"quark-mobile/internal/api"
	"quark-mobile/internal/driver"
	"quark-mobile/internal/driver/openlist"
	"quark-mobile/internal/service"
	"quark-mobile/internal/task"
)

func main() {
	// 加载配置（环境变量优先于配置文件）
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	viper.SetConfigFile(configPath)
	viper.SetEnvPrefix("")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("⚠️  Failed to read config file: %v", err)
		log.Printf("   Will use environment variables only")
	}

	// 环境变量覆盖（优先级最高）
	if envURL := os.Getenv("OL_BASE_URL"); envURL != "" {
		viper.Set("openlist.base_url", envURL)
	}
	if envUser := os.Getenv("OL_USERNAME"); envUser != "" {
		viper.Set("openlist.username", envUser)
	}
	if envPass := os.Getenv("OL_PASSWORD"); envPass != "" {
		viper.Set("openlist.password", envPass)
	}
	if envMountQuark := os.Getenv("OL_MOUNT_QUARK"); envMountQuark != "" {
		viper.Set("openlist.mounts.quark", envMountQuark)
	}
	if envMountMobile := os.Getenv("OL_MOUNT_MOBILE"); envMountMobile != "" {
		viper.Set("openlist.mounts.mobile", envMountMobile)
	}

	// 设置日志级别
	mode := viper.GetString("server.mode")
	if mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化缓存目录
	cacheDir := viper.GetString("transfer.cache_dir")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		log.Fatalf("failed to create cache dir: %v", err)
	}

	// 检查 config.yaml 文件权限
	if configPath == "config.yaml" {
		if info, err := os.Stat(configPath); err == nil {
			if info.Mode().Perm()&077 != 0 {
				log.Printf("⚠️  Warning: config.yaml has insecure permissions (%o), consider chmod 600", info.Mode().Perm())
			}
		}
	}

	// 初始化 OpenList 客户端
	openlistURL := viper.GetString("openlist.base_url")
	openlistUser := viper.GetString("openlist.username")
	openlistPass := viper.GetString("openlist.password")

	olClient := openlist.NewClient(openlistURL, "")

	// 如果配置了密码，尝试登录获取 token
	if openlistPass != "" {
		log.Printf("🔑 Logging into OpenList at %s as %s...", openlistURL, openlistUser)
		ctx := context.Background()
		if err := olClient.Login(ctx, openlistUser, openlistPass); err != nil {
			log.Printf("⚠️  Failed to login OpenList: %v", err)
			log.Printf("   Will continue without token (public access)")
		} else {
			log.Printf("✅ OpenList login successful")
			log.Printf("   Token acquired: %s...", olClient.Token[:min(len(olClient.Token), 20)])
		}
	}

	// 初始化驱动
	driver.InitDrivers(olClient)

	// 初始化服务
	transferSvc := service.NewTransferService()
	taskMgr := task.NewManager(transferSvc, viper.GetInt("transfer.max_concurrent"))

	// 启动 HTTP 服务
	addr := fmt.Sprintf(":%d", viper.GetInt("server.port"))
	router := api.NewRouter(taskMgr, transferSvc)

	log.Printf("🚀 Quark-Mobile Transfer Server starting on %s", addr)
	log.Printf("   OpenList: %s", openlistURL)
	log.Printf("   Mounts: quark=%s, mobile=%s",
		viper.GetString("openlist.mounts.quark"),
		viper.GetString("openlist.mounts.mobile"))
	log.Printf("   Max concurrent: %d", viper.GetInt("transfer.max_concurrent"))

	if err := router.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
