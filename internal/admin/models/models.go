package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// Admin represents an administrator
type Admin struct {
	ID          int64          `json:"id"`
	Username    string         `json:"username"`
	PasswordHash string        `json:"-"`
	Role        string         `json:"role"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	LastLoginAt sql.NullTime   `json:"last_login_at"`
	LastLoginIP sql.NullString `json:"last_login_ip"`
}

// AdminNotification represents a notification for admin
type AdminNotification struct {
	ID        int64           `json:"id"`
	AdminID   sql.NullInt64   `json:"admin_id"`
	Category  string          `json:"category"`
	Title     string          `json:"title"`
	Message   string          `json:"message"`
	Data      json.RawMessage `json:"data"`
	Priority  string          `json:"priority"`
	IsRead    bool            `json:"is_read"`
	Link      sql.NullString  `json:"link"`
	CreatedAt time.Time       `json:"created_at"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID         int64           `json:"id"`
	AdminID    int64           `json:"admin_id"`
	Module     string          `json:"module"`
	Action     string          `json:"action"`
	TargetType sql.NullString  `json:"target_type"`
	TargetID   sql.NullInt64   `json:"target_id"`
	IPAddress  sql.NullString  `json:"ip_address"`
	UserAgent  sql.NullString  `json:"user_agent"`
	Details    json.RawMessage `json:"details"`
	CreatedAt  time.Time       `json:"created_at"`
}

// LoginLog represents a login log entry
type LoginLog struct {
	ID        int64          `json:"id"`
	AdminID   int64          `json:"admin_id"`
	IPAddress string         `json:"ip_address"`
	UserAgent sql.NullString `json:"user_agent"`
	Device    sql.NullString `json:"device"`
	Browser   sql.NullString `json:"browser"`
	Status    string         `json:"status"`
	CreatedAt time.Time      `json:"created_at"`
}

// BroadcastMessage represents a broadcast message
type BroadcastMessage struct {
	ID             int64           `json:"id"`
	AdminID        int64           `json:"admin_id"`
	Title          string          `json:"title"`
	Content        string          `json:"content"`
	MessageType    string          `json:"message_type"`
	MessageContent string          `json:"message_content"`
	MediaURL       sql.NullString  `json:"media_url"`
	TargetType     string          `json:"target_type"`
	TargetUserIDs  sql.NullString  `json:"target_user_ids"`
	TargetFilters  json.RawMessage `json:"target_filters"`
	TotalUsers     int             `json:"total_users"`
	TotalCount     int             `json:"total_count"`
	SuccessCount   int             `json:"success_count"`
	FailedCount    int             `json:"failed_count"`
	Status         string          `json:"status"`
	ScheduledAt    sql.NullTime    `json:"scheduled_at"`
	SentAt         sql.NullTime    `json:"sent_at"`
	CompletedAt    sql.NullTime    `json:"completed_at"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// BroadcastDetail represents a broadcast detail entry
type BroadcastDetail struct {
	ID           int64          `json:"id"`
	BroadcastID  int64          `json:"broadcast_id"`
	UserID       int64          `json:"user_id"`
	Status       string         `json:"status"`
	ErrorMessage sql.NullString `json:"error_message"`
	SentAt       time.Time      `json:"sent_at"`
}

// MessageTemplate represents a message template
type MessageTemplate struct {
	ID             int64          `json:"id"`
	Name           string         `json:"name"`
	Content        string         `json:"content"`
	TemplateType   string         `json:"template_type"`
	Variables      []string       `json:"variables"`
	MessageType    string         `json:"message_type"`
	MessageContent string         `json:"message_content"`
	MediaURL       sql.NullString `json:"media_url"`
	UsageCount     int            `json:"usage_count"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// AutoMessageConfig represents an auto message configuration
type AutoMessageConfig struct {
	ID               int64           `json:"id"`
	ConfigType       string          `json:"config_type"`
	Type             string          `json:"type"`
	MessageContent   string          `json:"message_content"`
	IsEnabled        bool            `json:"is_enabled"`
	TemplateID       sql.NullInt64   `json:"template_id"`
	DelaySeconds     int             `json:"delay_seconds"`
	TriggerCondition json.RawMessage `json:"trigger_condition"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// SystemConfig represents a system configuration
type SystemConfig struct {
	ID          int64          `json:"id"`
	ConfigKey   string         `json:"config_key"`
	ConfigValue string         `json:"config_value"`
	ConfigType  string         `json:"config_type"`
	Description sql.NullString `json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// UserWithAdmin extends user with admin-specific fields
type UserWithAdmin struct {
	ID             int64          `json:"id"`
	Phone          string         `json:"phone"`
	Username       sql.NullString `json:"username"`
	FirstName      sql.NullString `json:"first_name"`
	LastName       sql.NullString `json:"last_name"`
	Bio            sql.NullString `json:"bio"`
	Status         string         `json:"status"`
	CustomCode     sql.NullString `json:"custom_code"`
	CodeExpiresAt  sql.NullTime   `json:"code_expires_at"`
	AllowCall      bool           `json:"allow_call"`
	AllowVideoCall bool           `json:"allow_video_call"`
	Remark         sql.NullString `json:"remark"`
	IsBot          bool           `json:"is_bot"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      sql.NullTime   `json:"deleted_at"`
	LastLoginAt    sql.NullTime   `json:"last_login_at"`
	IsOnline       bool           `json:"is_online"`
}

// ServiceStatus represents a service status
type ServiceStatus struct {
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	Port         int       `json:"port,omitempty"`
	PID          int       `json:"pid,omitempty"`
	Uptime       int64     `json:"uptime,omitempty"`
	RestartCount int       `json:"restart_count,omitempty"`
	LastRestart  string    `json:"last_restart,omitempty"`
	ActiveRelays int       `json:"active_relays,omitempty"`
	TotalRelays  int       `json:"total_relays,omitempty"`
	Connections  int       `json:"connections,omitempty"`
	Traffic      *Traffic  `json:"traffic,omitempty"`
	Message      string    `json:"message,omitempty"`
	ResponseTime int64     `json:"response_time,omitempty"`
	CheckedAt    time.Time `json:"checked_at"`
}

// Traffic represents network traffic
type Traffic struct {
	Upload   int64 `json:"upload"`
	Download int64 `json:"download"`
}

// ServerStatus represents server resource status
type ServerStatus struct {
	CPU             float64 `json:"cpu"`
	Memory          float64 `json:"memory"`
	Disk            float64 `json:"disk"`
	Network         *Traffic `json:"network,omitempty"`
	CPUUsage        float64 `json:"cpu_usage"`
	CPUModel        string  `json:"cpu_model"`
	CPUCores        int     `json:"cpu_cores"`
	MemoryTotal     uint64  `json:"memory_total"`
	MemoryUsed      uint64  `json:"memory_used"`
	MemoryUsage     float64 `json:"memory_usage"`
	DiskTotal       uint64  `json:"disk_total"`
	DiskUsed        uint64  `json:"disk_used"`
	DiskUsage       float64 `json:"disk_usage"`
	NetworkIn       uint64  `json:"network_in"`
	NetworkOut      uint64  `json:"network_out"`
	Hostname        string  `json:"hostname"`
	OS              string  `json:"os"`
	Platform        string  `json:"platform"`
	PlatformVersion string  `json:"platform_version"`
	KernelVersion   string  `json:"kernel_version"`
	Uptime          uint64  `json:"uptime"`
	GoVersion       string  `json:"go_version"`
	NumGoroutine    int     `json:"num_goroutine"`
	GoMemAlloc      uint64  `json:"go_mem_alloc"`
	GoMemSys        uint64  `json:"go_mem_sys"`
}

// ApplicationStatus represents application status
type ApplicationStatus struct {
	OnlineUsers          int       `json:"online_users"`
	ActiveCalls          int       `json:"active_calls"`
	WebSocketConnections int       `json:"websocket_connections"`
	APIRequestsPerMinute int       `json:"api_requests_per_minute"`
	StartTime            time.Time `json:"start_time"`
	PID                  int       `json:"pid"`
	CPUUsage             float64   `json:"cpu_usage"`
	MemoryUsage          float64   `json:"memory_usage"`
	NumThreads           int       `json:"num_threads"`
	NumFDs               int       `json:"num_fds"`
	NumConnections       int       `json:"num_connections"`
	NumGoroutines        int       `json:"num_goroutines"`
	HeapAlloc            uint64    `json:"heap_alloc"`
	HeapSys              uint64    `json:"heap_sys"`
	HeapObjects          uint64    `json:"heap_objects"`
	GCPauseTotal         uint64    `json:"gc_pause_total"`
	NumGC                uint32    `json:"num_gc"`
}

// MonitorData represents real-time monitoring data
type MonitorData struct {
	Server            *ServerStatus      `json:"server,omitempty"`
	Application       *ApplicationStatus `json:"application,omitempty"`
	Services          map[string]string  `json:"services,omitempty"`
	Timestamp         time.Time          `json:"timestamp"`
	CPUUsage          float64            `json:"cpu_usage"`
	MemoryUsage       float64            `json:"memory_usage"`
	MemoryUsed        uint64             `json:"memory_used"`
	MemoryTotal       uint64             `json:"memory_total"`
	DiskUsage         float64            `json:"disk_usage"`
	DiskUsed          uint64             `json:"disk_used"`
	DiskTotal         uint64             `json:"disk_total"`
	NetworkIn         uint64             `json:"network_in"`
	NetworkOut        uint64             `json:"network_out"`
	ActiveConnections int                `json:"active_connections"`
}

// UserStatistics represents user statistics
type UserStatistics struct {
	TotalUsers       int          `json:"total_users"`
	ActiveUsers      int          `json:"active_users"`
	NewUsersToday    int          `json:"new_users_today"`
	NewUsersWeek     int          `json:"new_users_week"`
	NewUsersMonth    int          `json:"new_users_month"`
	TodayNewUsers    int          `json:"today_new_users"`
	WeekNewUsers     int          `json:"week_new_users"`
	MonthNewUsers    int          `json:"month_new_users"`
	TodayActiveUsers int          `json:"today_active_users"`
	WeekActiveUsers  int          `json:"week_active_users"`
	MonthActiveUsers int          `json:"month_active_users"`
	GrowthTrend      []DailyCount `json:"growth_trend"`
}

// CallStatistics represents call statistics
type CallStatistics struct {
	TotalCalls         int            `json:"total_calls"`
	TotalDuration      int64          `json:"total_duration"`
	TodayCalls         int            `json:"today_calls"`
	TodayDuration      int64          `json:"today_duration"`
	WeekCalls          int            `json:"week_calls"`
	MonthCalls         int            `json:"month_calls"`
	CallsToday         int            `json:"calls_today"`
	CallsWeek          int            `json:"calls_week"`
	CallsMonth         int            `json:"calls_month"`
	AvgDuration        float64        `json:"avg_duration"`
	VideoCalls         int            `json:"video_calls"`
	AudioCalls         int            `json:"audio_calls"`
	SuccessRate        float64        `json:"success_rate"`
	TypeDistribution   map[string]int `json:"type_distribution"`
	StatusDistribution map[string]int `json:"status_distribution"`
	HourlyDistribution []int          `json:"hourly_distribution"`
	DailyTrend         []DailyCount   `json:"daily_trend"`
}

// DailyCount represents a daily count
type DailyCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// BotStatistics represents bot statistics
type BotStatistics struct {
	TotalMessages        int     `json:"total_messages"`
	TodayMessages        int     `json:"today_messages"`
	WeekMessages         int     `json:"week_messages"`
	MonthMessages        int     `json:"month_messages"`
	BroadcastCount       int     `json:"broadcast_count"`
	TotalBroadcasts      int     `json:"total_broadcasts"`
	TotalMessagesSent    int     `json:"total_messages_sent"`
	SuccessfulDeliveries int     `json:"successful_deliveries"`
	FailedDeliveries     int     `json:"failed_deliveries"`
	AverageDeliveryRate  float64 `json:"average_delivery_rate"`
}

// DashboardData represents dashboard data
type DashboardData struct {
	TotalUsers        int             `json:"total_users"`
	OnlineUsers       int             `json:"online_users"`
	NewUsersToday     int             `json:"new_users_today"`
	TotalCalls        int             `json:"total_calls"`
	TodayCalls        int             `json:"today_calls"`
	CallsToday        int             `json:"calls_today"`
	TotalBroadcasts   int             `json:"total_broadcasts"`
	TotalCallDuration int64           `json:"total_call_duration"`
	UserGrowthTrend   []DailyCount    `json:"user_growth_trend"`
	CallTrend         []DailyCount    `json:"call_trend"`
	RecentLogins      []UserWithAdmin `json:"recent_logins"`
	RecentCalls       []CallRecord    `json:"recent_calls"`
	RecentLogs        []AuditLog      `json:"recent_logs"`
}

// CallRecord represents a call record for admin
type CallRecord struct {
	ID           int64        `json:"id"`
	CallerID     int64        `json:"caller_id"`
	CallerName   string       `json:"caller_name"`
	CalleeID     int64        `json:"callee_id"`
	CalleeName   string       `json:"callee_name"`
	CallType     string       `json:"call_type"`
	IsVideo      bool         `json:"is_video"`
	Status       string       `json:"status"`
	Duration     int          `json:"duration"`
	StartTime    time.Time    `json:"start_time"`
	StartedAt    sql.NullTime `json:"started_at"`
	EndTime      sql.NullTime `json:"end_time"`
	EndedAt      sql.NullTime `json:"ended_at"`
	CreatedAt    time.Time    `json:"created_at"`
	AccessHash   int64        `json:"access_hash"`
	E2EEEnabled  bool         `json:"e2ee_enabled"`
	E2EEVersion  string       `json:"e2ee_version"`
}

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type      string          `json:"type"`
	Category  string          `json:"category,omitempty"`
	Title     string          `json:"title,omitempty"`
	Message   string          `json:"message,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
	Priority  string          `json:"priority,omitempty"`
	Sound     bool            `json:"sound,omitempty"`
	Link      string          `json:"link,omitempty"`
	Timestamp int64           `json:"timestamp,omitempty"`
	Token     string          `json:"token,omitempty"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token     string `json:"token"`
	Admin     *Admin `json:"admin"`
	ExpiresAt int64  `json:"expires_at"`
}

// CreateUserRequest represents a create user request
type CreateUserRequest struct {
	Phone          string `json:"phone" binding:"required"`
	Username       string `json:"username"`
	FirstName      string `json:"first_name"`
	CustomCode     string `json:"custom_code"`
	CodeExpiresAt  string `json:"code_expires_at"`
	Status         string `json:"status"`
	AllowCall      bool   `json:"allow_call"`
	AllowVideoCall bool   `json:"allow_video_call"`
	Remark         string `json:"remark"`
}

// UpdateUserRequest represents an update user request
type UpdateUserRequest struct {
	Username       string `json:"username"`
	FirstName      string `json:"first_name"`
	Phone          string `json:"phone"`
	CustomCode     string `json:"custom_code"`
	CodeExpiresAt  string `json:"code_expires_at"`
	Status         string `json:"status"`
	AllowCall      *bool  `json:"allow_call"`
	AllowVideoCall *bool  `json:"allow_video_call"`
	Remark         string `json:"remark"`
}

// BatchOperationRequest represents a batch operation request
type BatchOperationRequest struct {
	UserIDs   []int64 `json:"user_ids" binding:"required"`
	Operation string  `json:"operation" binding:"required"`
}

// SendBroadcastRequest represents a send broadcast request
type SendBroadcastRequest struct {
	Title          string          `json:"title"`
	Content        string          `json:"content"`
	MessageType    string          `json:"message_type" binding:"required"`
	MessageContent string          `json:"message_content"`
	MediaURL       string          `json:"media_url"`
	TargetType     string          `json:"target_type" binding:"required"`
	TargetUserIDs  []int64         `json:"target_user_ids"`
	TargetFilters  json.RawMessage `json:"target_filters"`
	ScheduledAt    string          `json:"scheduled_at"`
}

// CreateTemplateRequest represents a create template request
type CreateTemplateRequest struct {
	Name           string   `json:"name" binding:"required"`
	Content        string   `json:"content"`
	TemplateType   string   `json:"template_type"`
	Variables      []string `json:"variables"`
	MessageType    string   `json:"message_type"`
	MessageContent string   `json:"message_content"`
	MediaURL       string   `json:"media_url"`
}

// UpdateAutoMessageRequest represents an update auto message request
type UpdateAutoMessageRequest struct {
	MessageContent string `json:"message_content"`
	IsEnabled      *bool  `json:"is_enabled"`
	TemplateID     *int64 `json:"template_id"`
	DelaySeconds   int    `json:"delay_seconds"`
}

// ChangePasswordRequest represents a change password request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

// UpdateConfigRequest represents an update config request
type UpdateConfigRequest struct {
	ConfigValue string `json:"config_value" binding:"required"`
}

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"page_size" json:"page_size"`
	Search   string `form:"search" json:"search"`
	Status   string `form:"status" json:"status"`
	SortBy   string `form:"sort_by" json:"sort_by"`
	SortOrder string `form:"sort_order" json:"sort_order"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	Data     interface{} `json:"data"`
}
