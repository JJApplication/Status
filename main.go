package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// PageData 页面数据结构
type PageData struct {
	// Title 页面标题
	Title string
	// Services 服务列表
	Services []*Service
	// LastUpdated 最后更新时间
	LastUpdated string
}

var (
	// serviceManager 全局服务管理器
	serviceManager *ServiceManager
)

// initServices 初始化服务列表
func initServices() {
	serviceManager = NewServiceManager()

	// 添加示例服务
	serviceManager.AddService(&Service{
		Name:        "JJApps Center",
		Description: "微服务管理中心",
		URL:         "https://service.renj.io",
		Status:      StatusOnline,
		Checker: &CmdChecker{
			ProcessName: "apollo",
			Timeout:     5 * time.Second,
		},
	})

	serviceManager.AddService(&Service{
		Name:        "Sandwich Proxy",
		Description: " Sandwich 网关代理服务",
		URL:         "",
		Status:      StatusOnline,
		Checker: &CmdChecker{
			ProcessName: "sandwich",
			Timeout:     5 * time.Second,
		},
	})

	serviceManager.AddService(&Service{
		Name:        "Helios",
		Description: "前端静态代理服务",
		URL:         "",
		Status:      StatusOnline,
		Checker: &CmdChecker{
			ProcessName: "helios",
			Timeout:     5 * time.Second,
		},
	})

	serviceManager.AddService(&Service{
		Name:        "Black Hole",
		Description: "内容分发网络",
		URL:         "https://pkg.renj.io",
		Status:      StatusOnline,
		Checker: &CmdChecker{
			ProcessName: "black-hole",
			Timeout:     5 * time.Second,
		},
	})

	serviceManager.AddService(&Service{
		Name:        "Docker",
		Description: "Docker 容器进程",
		URL:         "",
		Status:      StatusOnline,
		Checker: &CmdChecker{
			ProcessName: "docker",
			Timeout:     5 * time.Second,
		},
	})

	serviceManager.AddService(&Service{
		Name:        "Proxy",
		Description: "Proxy代理",
		URL:         "",
		Status:      StatusOnline,
		Checker: &CmdChecker{
			ProcessName: "xray",
			Timeout:     5 * time.Second,
		},
	})

	// 初始化时更新一次状态
	serviceManager.UpdateAllStatus()
}

// indexHandler 首页处理器
func indexHandler(c *gin.Context) {
	// 不再同步更新状态，快速渲染页面
	// 准备页面数据（使用缓存的服务列表，不更新状态）
	data := PageData{
		Title:       "JJApps Status",
		Services:    serviceManager.GetServices(),
		LastUpdated: "加载中...",
	}

	// 渲染模板
	c.HTML(http.StatusOK, "index.html", data)
}

// apiStatusHandler API状态接口
func apiStatusHandler(c *gin.Context) {
	// 更新服务状态
	go serviceManager.UpdateAllStatus()

	// 返回JSON格式的服务状态
	c.JSON(http.StatusOK, gin.H{
		"services":     serviceManager.GetServices(),
		"last_updated": time.Now().Format("2006-01-02 15:04:05"),
	})
}

func main() {
	// 初始化服务
	initServices()

	// 创建Gin引擎
	r := gin.Default()
	gin.SetMode(gin.ReleaseMode)
	// 加载HTML模板
	r.LoadHTMLGlob("templates/*")

	// 静态文件服务
	r.Static("/static", "./static")

	// 路由设置
	r.GET("/", indexHandler)
	r.GET("/api/status", apiStatusHandler)

	port := os.Getenv("PORTS")
	if port == "" {
		return
	}
	// 启动服务器
	r.Run(fmt.Sprintf("127.0.0.1:%s", port))
}
