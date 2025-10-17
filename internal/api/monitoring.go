package api

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/hoytech/mcpmanager/internal/core/discovery"
	"github.com/hoytech/mcpmanager/internal/core/monitoring"
	"github.com/hoytech/mcpmanager/internal/models"
)

// MonitoringHandlers contains HTTP handlers for monitoring endpoints
type MonitoringHandlers struct {
	monitoringService *monitoring.MonitoringService
	metricsCollector  *monitoring.MetricsCollector
	discoveryService  *discovery.DiscoveryService
}

// NewMonitoringHandlers creates a new MonitoringHandlers instance
func NewMonitoringHandlers(monitoringService *monitoring.MonitoringService, metricsCollector *monitoring.MetricsCollector, discoveryService *discovery.DiscoveryService) *MonitoringHandlers {
	return &MonitoringHandlers{
		monitoringService: monitoringService,
		metricsCollector:  metricsCollector,
		discoveryService:  discoveryService,
	}
}

// LogsResponse is the response structure for GET /servers/{serverId}/logs
type LogsResponse struct {
	Logs    []models.LogEntry `json:"logs"`
	Total   int               `json:"total"`
	HasMore bool              `json:"hasMore"`
}

// AllLogsResponse is the response structure for GET /logs
type AllLogsResponse struct {
	Logs  []models.LogEntry `json:"logs"`
	Total int               `json:"total"`
}

// MetricsResponse is the response structure for GET /servers/{serverId}/metrics
type MetricsResponse struct {
	UptimeSeconds  *int     `json:"uptimeSeconds"`
	MemoryUsageMB  *float64 `json:"memoryUsageMB"`
	RequestCount   *int64   `json:"requestCount"`
	CPUPercent     *float64 `json:"cpuPercent"`
}

// GetServerLogs handles GET /api/v1/servers/{serverId}/logs
func (h *MonitoringHandlers) GetServerLogs(w http.ResponseWriter, r *http.Request) {
	// Extract server ID from URL
	serverID := chi.URLParam(r, "serverId")

	// Validate UUID format
	if _, err := uuid.Parse(serverID); err != nil {
		respondError(w, http.StatusNotFound, "Invalid server ID format")
		return
	}

	// Check if server exists
	_, exists := h.discoveryService.GetServerByID(serverID)
	if !exists {
		respondError(w, http.StatusNotFound, "Server not found")
		return
	}

	// Parse query parameters
	severityFilter := r.URL.Query().Get("severity")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Default values
	limit := 100
	offset := 0

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 1000 {
			limit = parsedLimit
		}
	}

	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get logs from monitoring service
	logs := h.monitoringService.GetLogs(serverID, offset, limit)

	// Filter by severity if provided
	var filteredLogs []models.LogEntry
	if severityFilter != "" {
		for _, log := range logs {
			if string(log.Severity) == severityFilter {
				filteredLogs = append(filteredLogs, log)
			}
		}
	} else {
		filteredLogs = logs
	}

	// Handle nil slice
	if filteredLogs == nil {
		filteredLogs = []models.LogEntry{}
	}

	// Get total count (approximate - from all logs for this server)
	allLogs := h.monitoringService.GetAllLogs(serverID)
	total := len(allLogs)

	// Calculate if there are more logs
	hasMore := (offset + len(filteredLogs)) < total

	response := LogsResponse{
		Logs:    filteredLogs,
		Total:   total,
		HasMore: hasMore,
	}

	respondJSON(w, http.StatusOK, response)
}

// GetAllLogs handles GET /api/v1/logs
func (h *MonitoringHandlers) GetAllLogs(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	serverIDFilter := r.URL.Query().Get("serverId")
	severityFilter := r.URL.Query().Get("severity")
	searchQuery := r.URL.Query().Get("search")
	limitStr := r.URL.Query().Get("limit")

	// Default limit
	limit := 100
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 1000 {
			limit = parsedLimit
		}
	}

	// Validate serverID if provided
	if serverIDFilter != "" {
		if _, err := uuid.Parse(serverIDFilter); err != nil {
			respondError(w, http.StatusBadRequest, "Invalid server ID format")
			return
		}
	}

	// Get all servers
	servers, _, _ := h.discoveryService.GetServers()

	// Collect logs from all servers (or filtered server)
	var allLogs []models.LogEntry

	if serverIDFilter != "" {
		// Get logs for specific server
		logs := h.monitoringService.GetAllLogs(serverIDFilter)
		allLogs = append(allLogs, logs...)
	} else {
		// Get logs from all servers
		for _, server := range servers {
			logs := h.monitoringService.GetAllLogs(server.ID)
			allLogs = append(allLogs, logs...)
		}
	}

	// Apply filters
	var filteredLogs []models.LogEntry
	for _, log := range allLogs {
		// Filter by severity
		if severityFilter != "" && string(log.Severity) != severityFilter {
			continue
		}

		// Filter by search query
		if searchQuery != "" && !contains(log.Message, searchQuery) {
			continue
		}

		filteredLogs = append(filteredLogs, log)
	}

	// Handle nil slice
	if filteredLogs == nil {
		filteredLogs = []models.LogEntry{}
	}

	// Apply limit
	total := len(filteredLogs)
	if len(filteredLogs) > limit {
		filteredLogs = filteredLogs[:limit]
	}

	response := AllLogsResponse{
		Logs:  filteredLogs,
		Total: total,
	}

	respondJSON(w, http.StatusOK, response)
}

// GetServerMetrics handles GET /api/v1/servers/{serverId}/metrics
func (h *MonitoringHandlers) GetServerMetrics(w http.ResponseWriter, r *http.Request) {
	// Extract server ID from URL
	serverID := chi.URLParam(r, "serverId")

	// Validate UUID format
	if _, err := uuid.Parse(serverID); err != nil {
		respondError(w, http.StatusNotFound, "Invalid server ID format")
		return
	}

	// Get server
	server, exists := h.discoveryService.GetServerByID(serverID)
	if !exists {
		respondError(w, http.StatusNotFound, "Server not found")
		return
	}

	// Get PID
	pid := 0
	if server.PID != nil {
		pid = *server.PID
	}

	// Get metrics from metrics collector
	metrics, err := h.metricsCollector.GetMetrics(serverID, &server.Status, pid)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve metrics: "+err.Error())
		return
	}

	// Build response with nullable fields
	response := MetricsResponse{
		UptimeSeconds:  nil,
		MemoryUsageMB:  nil,
		RequestCount:   nil,
		CPUPercent:     nil,
	}

	// Convert uptime to seconds (nullable)
	if metrics.Uptime > 0 {
		uptimeSeconds := int(metrics.Uptime.Seconds())
		response.UptimeSeconds = &uptimeSeconds
	}

	// Convert memory bytes to MB (nullable)
	if metrics.MemoryBytes != nil && *metrics.MemoryBytes > 0 {
		memoryMB := float64(*metrics.MemoryBytes) / (1024 * 1024)
		response.MemoryUsageMB = &memoryMB
	}

	// Request count (nullable)
	if metrics.RequestCount != nil {
		response.RequestCount = metrics.RequestCount
	}

	// CPU percent is not implemented yet (nullable)
	// response.CPUPercent remains nil

	respondJSON(w, http.StatusOK, response)
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		strContains(s, substr))
}

func strContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
