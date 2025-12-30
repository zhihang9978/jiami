package handlers

import (
	"net/http"
	"strconv"

	"github.com/feiji/feiji-backend/internal/admin/monitor"
	"github.com/gin-gonic/gin"
)

// MonitorHandlers contains all monitor API handlers
type MonitorHandlers struct {
	monitor *monitor.Monitor
}

// NewMonitorHandlers creates new monitor handlers
func NewMonitorHandlers() *MonitorHandlers {
	return &MonitorHandlers{
		monitor: monitor.NewMonitor(),
	}
}

// GetServerStatus returns server status
func (h *MonitorHandlers) GetServerStatus(c *gin.Context) {
	status, err := h.monitor.GetServerStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetApplicationStatus returns application status
func (h *MonitorHandlers) GetApplicationStatus(c *gin.Context) {
	status, err := h.monitor.GetApplicationStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetServiceStatus returns status of a specific service
func (h *MonitorHandlers) GetServiceStatus(c *gin.Context) {
	serviceName := c.Param("name")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Service name required"})
		return
	}

	status, err := h.monitor.GetServiceStatus(c.Request.Context(), serviceName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetAllServicesStatus returns status of all services
func (h *MonitorHandlers) GetAllServicesStatus(c *gin.Context) {
	statuses, err := h.monitor.GetAllServicesStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, statuses)
}

// RestartService restarts a service
func (h *MonitorHandlers) RestartService(c *gin.Context) {
	serviceName := c.Param("name")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Service name required"})
		return
	}

	if err := h.monitor.RestartService(c.Request.Context(), serviceName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Service restart initiated"})
}

// StopService stops a service
func (h *MonitorHandlers) StopService(c *gin.Context) {
	serviceName := c.Param("name")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Service name required"})
		return
	}

	if err := h.monitor.StopService(c.Request.Context(), serviceName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Service stop initiated"})
}

// StartService starts a service
func (h *MonitorHandlers) StartService(c *gin.Context) {
	serviceName := c.Param("name")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Service name required"})
		return
	}

	if err := h.monitor.StartService(c.Request.Context(), serviceName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Service start initiated"})
}

// GetServiceLogs returns service logs
func (h *MonitorHandlers) GetServiceLogs(c *gin.Context) {
	serviceName := c.Param("name")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Service name required"})
		return
	}

	lines, _ := strconv.Atoi(c.DefaultQuery("lines", "100"))
	if lines <= 0 {
		lines = 100
	}
	if lines > 1000 {
		lines = 1000
	}

	logs, err := h.monitor.GetServiceLogs(c.Request.Context(), serviceName, lines)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"service": serviceName,
		"lines":   len(logs),
		"logs":    logs,
	})
}

// GetMonitorData returns real-time monitoring data
func (h *MonitorHandlers) GetMonitorData(c *gin.Context) {
	data, err := h.monitor.GetMonitorData(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}
