package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"quark-mobile/internal/driver"
	"quark-mobile/internal/model"
	"quark-mobile/internal/service"
	"quark-mobile/internal/task"
)

type Router struct {
	router      *gin.Engine
	taskMgr     *task.Manager
	transferSvc *service.TransferService
}

func NewRouter(taskMgr *task.Manager, transferSvc *service.TransferService) *Router {
	r := &Router{
		router:      gin.Default(),
		taskMgr:     taskMgr,
		transferSvc: transferSvc,
	}
	r.setupRoutes()
	return r
}

func (r *Router) setupRoutes() {
	api := r.router.Group("/api")
	{
		// 驱动列表
		api.GET("/drivers", r.listDrivers)

		// 文件浏览
		api.GET("/files/:driver", r.listFiles)

		// 传输任务
		api.POST("/transfer", r.createTransfer)
		api.GET("/tasks", r.listTasks)
		api.GET("/tasks/:id", r.getTask)
		api.DELETE("/tasks/:id", r.cancelTask)
	}

	// 静态文件（前端）
	r.router.Static("/", "./web/dist")
	r.router.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.File("./web/dist/index.html")
	})
}

func (r *Router) Run(addr string) error {
	return r.router.Run(addr)
}

// GET /api/drivers
func (r *Router) listDrivers(c *gin.Context) {
	drivers := driver.RegisteredDrivers()
	result := make([]gin.H, 0, len(drivers))
	for _, d := range drivers {
		result = append(result, gin.H{
			"type": string(d),
		})
	}
	c.JSON(http.StatusOK, gin.H{"drivers": result})
}

// GET /api/files/:driver?path=/
func (r *Router) listFiles(c *gin.Context) {
	driverType := model.DriverType(c.Param("driver"))
	path := c.DefaultQuery("path", "/")

	files, err := r.transferSvc.GetFileList(c.Request.Context(), driverType, path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"driver": driverType,
		"path":   path,
		"files":  files,
	})
}

// POST /api/transfer
func (r *Router) createTransfer(c *gin.Context) {
	var req model.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证驱动
	if _, err := driver.GetDriver(req.SourceDriver); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if _, err := driver.GetDriver(req.TargetDriver); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tsk := r.taskMgr.CreateTask(req)
	c.JSON(http.StatusOK, gin.H{
		"task": tsk,
	})
}

// GET /api/tasks
func (r *Router) listTasks(c *gin.Context) {
	tasks := r.taskMgr.ListTasks()
	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

// GET /api/tasks/:id
func (r *Router) getTask(c *gin.Context) {
	id := c.Param("id")
	tsk, err := r.taskMgr.GetTask(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"task": tsk})
}

// DELETE /api/tasks/:id
func (r *Router) cancelTask(c *gin.Context) {
	id := c.Param("id")
	if err := r.taskMgr.CancelTask(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "task cancelled"})
}
