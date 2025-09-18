package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"sync"
	"time"
)

// ServiceStatus 表示服务状态
type ServiceStatus int

const (
	// StatusOnline 在线状态
	StatusOnline ServiceStatus = iota
	// StatusOffline 离线状态
	StatusOffline
)

// String 返回状态的字符串表示
func (s ServiceStatus) String() string {
	switch s {
	case StatusOnline:
		return "online"
	case StatusOffline:
		return "offline"
	default:
		return "unknown"
	}
}

// MarshalJSON 实现JSON序列化
func (s ServiceStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%d", int(s))), nil
}

// StatusChecker 定义状态检查接口
type StatusChecker interface {
	// CheckStatus 检查服务状态，返回状态和错误信息
	CheckStatus() (ServiceStatus, error)
}

// Service 表示一个服务
type Service struct {
	// Name 服务名称
	Name string `json:"name"`
	// Description 服务描述
	Description string `json:"description"`
	// URL 服务URL
	URL string `json:"url"`
	// Status 当前状态
	Status ServiceStatus `json:"status"`
	// LastChecked 最后检查时间
	LastChecked time.Time `json:"last_checked"`
	// Checker 状态检查器
	Checker StatusChecker `json:"-"`
}

// HTTPChecker HTTP状态检查器
type HTTPChecker struct {
	// URL 要检查的URL
	URL string
	// Timeout 超时时间
	Timeout time.Duration
}

// CheckStatus 实现StatusChecker接口，检查HTTP服务状态
func (h *HTTPChecker) CheckStatus() (ServiceStatus, error) {
	client := &http.Client{
		Timeout: h.Timeout,
	}

	resp, err := client.Get(h.URL)
	if err != nil {
		return StatusOffline, fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return StatusOnline, nil
	}

	return StatusOffline, fmt.Errorf("HTTP状态码异常: %d", resp.StatusCode)
}

// PingChecker 简单的ping检查器（模拟）
type PingChecker struct {
	// Host 主机地址
	Host string
}

// CheckStatus 实现StatusChecker接口，模拟ping检查
func (p *PingChecker) CheckStatus() (ServiceStatus, error) {
	// 这里简化实现，实际项目中可以使用真正的ping
	// 为了演示，我们随机返回状态
	if len(p.Host) > 0 {
		return StatusOnline, nil
	}
	return StatusOffline, fmt.Errorf("主机地址为空")
}

// CmdChecker 基于命令行的进程状态检查器
type CmdChecker struct {
	// ProcessName 要检查的进程名称
	ProcessName string
	// Timeout 命令执行超时时间
	Timeout time.Duration
}

// CheckStatus 实现StatusChecker接口，通过ps命令检查进程状态
func (c *CmdChecker) CheckStatus() (ServiceStatus, error) {
	if c.ProcessName == "" {
		return StatusOffline, fmt.Errorf("进程名称不能为空")
	}

	// 设置默认超时时间
	timeout := c.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	// 构建命令：ps ax | grep 进程名 | grep -v grep
	cmdStr := fmt.Sprintf("ps ax | grep '%s' | grep -v grep", c.ProcessName)
	cmd := exec.Command("bash", "-c", cmdStr)

	// 设置超时
	done := make(chan error, 1)
	go func() {
		_, err := cmd.Output()
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			// 如果命令执行失败或没有找到进程，认为服务离线
			return StatusOffline, fmt.Errorf("进程 '%s' 未运行: %v", c.ProcessName, err)
		}
		// 命令执行成功且有输出，说明进程存在
		return StatusOnline, nil
	case <-time.After(timeout):
		// 超时
		cmd.Process.Kill()
		return StatusOffline, fmt.Errorf("检查进程 '%s' 超时", c.ProcessName)
	}
}

// ServiceManager 服务管理器
type ServiceManager struct {
	lock        *sync.RWMutex
	refreshFlag bool
	// services 服务列表
	services []*Service
}

// NewServiceManager 创建新的服务管理器
func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		lock:        new(sync.RWMutex),
		refreshFlag: false,
		services:    make([]*Service, 0),
	}
}

// AddService 添加服务
func (sm *ServiceManager) AddService(service *Service) {
	sm.services = append(sm.services, service)
}

// GetServices 获取所有服务
func (sm *ServiceManager) GetServices() []*Service {
	return sm.services
}

// UpdateStatus 更新服务状态
func (sm *ServiceManager) UpdateStatus(service *Service) {
	if service.Checker != nil {
		status, err := service.Checker.CheckStatus()
		service.Status = status
		service.LastChecked = time.Now()
		if err != nil {
			fmt.Printf("检查服务 %s 状态时出错: %v\n", service.Name, err)
		}
	}
}

// UpdateAllStatus 更新所有服务状态
func (sm *ServiceManager) UpdateAllStatus() {
	sm.lock.Lock()
	defer sm.lock.Unlock()
	if sm.refreshFlag {
		return
	}
	sm.refreshFlag = true
	for _, service := range sm.services {
		sm.UpdateStatus(service)
	}
	sm.refreshFlag = false
}
