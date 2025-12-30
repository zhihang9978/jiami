package monitor

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/feiji/feiji-backend/internal/admin/models"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	psnet "github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// Monitor handles system monitoring
type Monitor struct {
	mu sync.RWMutex
	// Cache for service status
	serviceCache map[string]*models.ServiceStatus
	cacheTime    time.Time
	cacheTTL     time.Duration
}

// NewMonitor creates a new monitor
func NewMonitor() *Monitor {
	return &Monitor{
		serviceCache: make(map[string]*models.ServiceStatus),
		cacheTTL:     30 * time.Second,
	}
}

// GetServerStatus returns current server status
func (m *Monitor) GetServerStatus(ctx context.Context) (*models.ServerStatus, error) {
	status := &models.ServerStatus{}

	// CPU info
	cpuPercent, err := cpu.PercentWithContext(ctx, time.Second, false)
	if err == nil && len(cpuPercent) > 0 {
		status.CPUUsage = cpuPercent[0]
	}

	cpuInfo, err := cpu.InfoWithContext(ctx)
	if err == nil && len(cpuInfo) > 0 {
		status.CPUModel = cpuInfo[0].ModelName
		status.CPUCores = int(cpuInfo[0].Cores)
	}

	// Memory info
	memInfo, err := mem.VirtualMemoryWithContext(ctx)
	if err == nil {
		status.MemoryTotal = memInfo.Total
		status.MemoryUsed = memInfo.Used
		status.MemoryUsage = memInfo.UsedPercent
	}

	// Disk info
	diskInfo, err := disk.UsageWithContext(ctx, "/")
	if err == nil {
		status.DiskTotal = diskInfo.Total
		status.DiskUsed = diskInfo.Used
		status.DiskUsage = diskInfo.UsedPercent
	}

	// Network info
	netIO, err := psnet.IOCountersWithContext(ctx, false)
	if err == nil && len(netIO) > 0 {
		status.NetworkIn = netIO[0].BytesRecv
		status.NetworkOut = netIO[0].BytesSent
	}

	// Host info
	hostInfo, err := host.InfoWithContext(ctx)
	if err == nil {
		status.Hostname = hostInfo.Hostname
		status.OS = hostInfo.OS
		status.Platform = hostInfo.Platform
		status.PlatformVersion = hostInfo.PlatformVersion
		status.KernelVersion = hostInfo.KernelVersion
		status.Uptime = hostInfo.Uptime
	}

	// Go runtime info
	status.GoVersion = runtime.Version()
	status.NumGoroutine = runtime.NumGoroutine()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	status.GoMemAlloc = memStats.Alloc
	status.GoMemSys = memStats.Sys

	return status, nil
}

// GetApplicationStatus returns application status
func (m *Monitor) GetApplicationStatus(ctx context.Context) (*models.ApplicationStatus, error) {
	status := &models.ApplicationStatus{
		StartTime: time.Now().Add(-time.Duration(getUptime()) * time.Second),
	}

	// Get current process info
	pid := int32(getPID())
	proc, err := process.NewProcess(pid)
	if err == nil {
		status.PID = int(pid)

		cpuPercent, _ := proc.CPUPercentWithContext(ctx)
		status.CPUUsage = cpuPercent

		memInfo, _ := proc.MemoryInfoWithContext(ctx)
		if memInfo != nil {
			status.MemoryUsage = float64(memInfo.RSS)
		}

		numThreads, _ := proc.NumThreadsWithContext(ctx)
		status.NumThreads = int(numThreads)

		numFDs, _ := proc.NumFDsWithContext(ctx)
		status.NumFDs = int(numFDs)

		connections, _ := proc.ConnectionsWithContext(ctx)
		status.NumConnections = len(connections)
	}

	// Go runtime stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	status.NumGoroutines = runtime.NumGoroutine()
	status.HeapAlloc = memStats.HeapAlloc
	status.HeapSys = memStats.HeapSys
	status.HeapObjects = memStats.HeapObjects
	status.GCPauseTotal = memStats.PauseTotalNs
	status.NumGC = memStats.NumGC

	return status, nil
}

// GetServiceStatus returns status of a specific service
func (m *Monitor) GetServiceStatus(ctx context.Context, serviceName string) (*models.ServiceStatus, error) {
	m.mu.RLock()
	if cached, ok := m.serviceCache[serviceName]; ok && time.Since(m.cacheTime) < m.cacheTTL {
		m.mu.RUnlock()
		return cached, nil
	}
	m.mu.RUnlock()

	status := &models.ServiceStatus{
		Name:      serviceName,
		Status:    "unknown",
		CheckedAt: time.Now(),
	}

	switch serviceName {
	case "api":
		status = m.checkHTTPService("API Server", "http://localhost:8080/health")
	case "websocket":
		status = m.checkWebSocketService("WebSocket Server", "ws://localhost:8080/ws")
	case "turn":
		status = m.checkTURNService("TURN Server", "43.229.114.106:3478")
	case "stun":
		status = m.checkSTUNService("STUN Server", "43.229.114.106:3478")
	case "mysql":
		status = m.checkMySQLService("MySQL", "localhost:3306")
	case "redis":
		status = m.checkRedisService("Redis", "localhost:6379")
	case "nginx":
		status = m.checkHTTPService("Nginx", "http://localhost:80")
	default:
		status.Status = "unknown"
		status.Message = "Unknown service"
	}

	// Cache the result
	m.mu.Lock()
	m.serviceCache[serviceName] = status
	m.cacheTime = time.Now()
	m.mu.Unlock()

	return status, nil
}

// GetAllServicesStatus returns status of all services
func (m *Monitor) GetAllServicesStatus(ctx context.Context) ([]*models.ServiceStatus, error) {
	services := []string{"api", "websocket", "turn", "stun", "mysql", "redis", "nginx"}
	var statuses []*models.ServiceStatus

	for _, service := range services {
		status, _ := m.GetServiceStatus(ctx, service)
		statuses = append(statuses, status)
	}

	return statuses, nil
}

// checkHTTPService checks HTTP service status
func (m *Monitor) checkHTTPService(name, url string) *models.ServiceStatus {
	status := &models.ServiceStatus{
		Name:      name,
		Status:    "offline",
		CheckedAt: time.Now(),
	}

	client := &http.Client{Timeout: 5 * time.Second}
	start := time.Now()
	resp, err := client.Get(url)
	status.ResponseTime = time.Since(start).Milliseconds()

	if err != nil {
		status.Message = err.Error()
		return status
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		status.Status = "online"
		status.Message = fmt.Sprintf("HTTP %d", resp.StatusCode)
	} else {
		status.Message = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	return status
}

// checkWebSocketService checks WebSocket service status
func (m *Monitor) checkWebSocketService(name, url string) *models.ServiceStatus {
	status := &models.ServiceStatus{
		Name:      name,
		Status:    "offline",
		CheckedAt: time.Now(),
	}

	// For WebSocket, we just check if the port is open
	host := strings.TrimPrefix(url, "ws://")
	host = strings.TrimPrefix(host, "wss://")
	host = strings.Split(host, "/")[0]

	start := time.Now()
	conn, err := net.DialTimeout("tcp", host, 5*time.Second)
	status.ResponseTime = time.Since(start).Milliseconds()

	if err != nil {
		status.Message = err.Error()
		return status
	}
	conn.Close()

	status.Status = "online"
	status.Message = "Port is open"
	return status
}

// checkTURNService checks TURN service status
func (m *Monitor) checkTURNService(name, addr string) *models.ServiceStatus {
	status := &models.ServiceStatus{
		Name:      name,
		Status:    "offline",
		CheckedAt: time.Now(),
	}

	start := time.Now()
	conn, err := net.DialTimeout("udp", addr, 5*time.Second)
	status.ResponseTime = time.Since(start).Milliseconds()

	if err != nil {
		status.Message = err.Error()
		return status
	}
	conn.Close()

	status.Status = "online"
	status.Message = "TURN server is reachable"
	return status
}

// checkSTUNService checks STUN service status
func (m *Monitor) checkSTUNService(name, addr string) *models.ServiceStatus {
	status := &models.ServiceStatus{
		Name:      name,
		Status:    "offline",
		CheckedAt: time.Now(),
	}

	start := time.Now()
	conn, err := net.DialTimeout("udp", addr, 5*time.Second)
	status.ResponseTime = time.Since(start).Milliseconds()

	if err != nil {
		status.Message = err.Error()
		return status
	}
	conn.Close()

	status.Status = "online"
	status.Message = "STUN server is reachable"
	return status
}

// checkMySQLService checks MySQL service status
func (m *Monitor) checkMySQLService(name, addr string) *models.ServiceStatus {
	status := &models.ServiceStatus{
		Name:      name,
		Status:    "offline",
		CheckedAt: time.Now(),
	}

	start := time.Now()
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	status.ResponseTime = time.Since(start).Milliseconds()

	if err != nil {
		status.Message = err.Error()
		return status
	}
	conn.Close()

	status.Status = "online"
	status.Message = "MySQL is reachable"
	return status
}

// checkRedisService checks Redis service status
func (m *Monitor) checkRedisService(name, addr string) *models.ServiceStatus {
	status := &models.ServiceStatus{
		Name:      name,
		Status:    "offline",
		CheckedAt: time.Now(),
	}

	start := time.Now()
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	status.ResponseTime = time.Since(start).Milliseconds()

	if err != nil {
		status.Message = err.Error()
		return status
	}
	conn.Close()

	status.Status = "online"
	status.Message = "Redis is reachable"
	return status
}

// RestartService restarts a service
func (m *Monitor) RestartService(ctx context.Context, serviceName string) error {
	var cmd *exec.Cmd

	switch serviceName {
	case "api":
		cmd = exec.CommandContext(ctx, "systemctl", "restart", "feiji-api")
	case "nginx":
		cmd = exec.CommandContext(ctx, "systemctl", "restart", "nginx")
	case "mysql":
		cmd = exec.CommandContext(ctx, "systemctl", "restart", "mysql")
	case "redis":
		cmd = exec.CommandContext(ctx, "systemctl", "restart", "redis")
	case "turn":
		cmd = exec.CommandContext(ctx, "systemctl", "restart", "coturn")
	default:
		return fmt.Errorf("unknown service: %s", serviceName)
	}

	return cmd.Run()
}

// StopService stops a service
func (m *Monitor) StopService(ctx context.Context, serviceName string) error {
	var cmd *exec.Cmd

	switch serviceName {
	case "api":
		cmd = exec.CommandContext(ctx, "systemctl", "stop", "feiji-api")
	case "nginx":
		cmd = exec.CommandContext(ctx, "systemctl", "stop", "nginx")
	case "mysql":
		cmd = exec.CommandContext(ctx, "systemctl", "stop", "mysql")
	case "redis":
		cmd = exec.CommandContext(ctx, "systemctl", "stop", "redis")
	case "turn":
		cmd = exec.CommandContext(ctx, "systemctl", "stop", "coturn")
	default:
		return fmt.Errorf("unknown service: %s", serviceName)
	}

	return cmd.Run()
}

// StartService starts a service
func (m *Monitor) StartService(ctx context.Context, serviceName string) error {
	var cmd *exec.Cmd

	switch serviceName {
	case "api":
		cmd = exec.CommandContext(ctx, "systemctl", "start", "feiji-api")
	case "nginx":
		cmd = exec.CommandContext(ctx, "systemctl", "start", "nginx")
	case "mysql":
		cmd = exec.CommandContext(ctx, "systemctl", "start", "mysql")
	case "redis":
		cmd = exec.CommandContext(ctx, "systemctl", "start", "redis")
	case "turn":
		cmd = exec.CommandContext(ctx, "systemctl", "start", "coturn")
	default:
		return fmt.Errorf("unknown service: %s", serviceName)
	}

	return cmd.Run()
}

// GetServiceLogs returns service logs
func (m *Monitor) GetServiceLogs(ctx context.Context, serviceName string, lines int) ([]string, error) {
	var cmd *exec.Cmd

	switch serviceName {
	case "api":
		cmd = exec.CommandContext(ctx, "journalctl", "-u", "feiji-api", "-n", fmt.Sprintf("%d", lines), "--no-pager")
	case "nginx":
		cmd = exec.CommandContext(ctx, "tail", "-n", fmt.Sprintf("%d", lines), "/var/log/nginx/error.log")
	case "mysql":
		cmd = exec.CommandContext(ctx, "tail", "-n", fmt.Sprintf("%d", lines), "/var/log/mysql/error.log")
	case "redis":
		cmd = exec.CommandContext(ctx, "tail", "-n", fmt.Sprintf("%d", lines), "/var/log/redis/redis-server.log")
	default:
		return nil, fmt.Errorf("unknown service: %s", serviceName)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return strings.Split(string(output), "\n"), nil
}

// GetMonitorData returns real-time monitoring data
func (m *Monitor) GetMonitorData(ctx context.Context) (*models.MonitorData, error) {
	data := &models.MonitorData{
		Timestamp: time.Now(),
	}

	// CPU usage
	cpuPercent, err := cpu.PercentWithContext(ctx, time.Second, false)
	if err == nil && len(cpuPercent) > 0 {
		data.CPUUsage = cpuPercent[0]
	}

	// Memory usage
	memInfo, err := mem.VirtualMemoryWithContext(ctx)
	if err == nil {
		data.MemoryUsage = memInfo.UsedPercent
		data.MemoryUsed = memInfo.Used
		data.MemoryTotal = memInfo.Total
	}

	// Disk usage
	diskInfo, err := disk.UsageWithContext(ctx, "/")
	if err == nil {
		data.DiskUsage = diskInfo.UsedPercent
		data.DiskUsed = diskInfo.Used
		data.DiskTotal = diskInfo.Total
	}

	// Network I/O
	netIO, err := psnet.IOCountersWithContext(ctx, false)
	if err == nil && len(netIO) > 0 {
		data.NetworkIn = netIO[0].BytesRecv
		data.NetworkOut = netIO[0].BytesSent
	}

	// Active connections
	connections, err := psnet.ConnectionsWithContext(ctx, "tcp")
	if err == nil {
		data.ActiveConnections = len(connections)
	}

	return data, nil
}

// Helper functions
func getUptime() uint64 {
	info, err := host.Info()
	if err != nil {
		return 0
	}
	return info.Uptime
}

func getPID() int {
	return os.Getpid()
}
