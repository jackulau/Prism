package handlers

import (
	"context"
	"io"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/api/middleware"
	"github.com/jacklau/prism/internal/audit"
	"github.com/jacklau/prism/internal/export"
)

// ExportHandler handles data export endpoints
type ExportHandler struct {
	exportService      *export.Service
	auditLogger        *audit.Logger
	gdprExporter       *export.GDPRExporter
	complianceReporter *export.ComplianceReporter
}

// NewExportHandler creates a new export handler
func NewExportHandler(
	exportService *export.Service,
	auditLogger *audit.Logger,
	gdprExporter *export.GDPRExporter,
	complianceReporter *export.ComplianceReporter,
) *ExportHandler {
	return &ExportHandler{
		exportService:      exportService,
		auditLogger:        auditLogger,
		gdprExporter:       gdprExporter,
		complianceReporter: complianceReporter,
	}
}

// ExportRequest represents a request to create an export
type ExportRequest struct {
	Type       string                 `json:"type"`
	Format     string                 `json:"format"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// ListExportTypes returns available export types
func (h *ExportHandler) ListExportTypes(c *fiber.Ctx) error {
	types := []map[string]interface{}{
		{
			"type":        "gdpr",
			"name":        "GDPR Data Export",
			"description": "Complete export of your personal data as required by GDPR Article 15",
			"formats":     []string{"json", "zip"},
		},
		{
			"type":        "audit_logs",
			"name":        "Audit Logs",
			"description": "Export of audit trail events",
			"formats":     []string{"json", "csv"},
		},
		{
			"type":        "compliance",
			"name":        "Compliance Report",
			"description": "Regulatory compliance reports (SOC2, HIPAA)",
			"formats":     []string{"json"},
			"report_types": []string{
				"access_log",
				"change_log",
				"security_log",
				"user_activity",
				"soc2",
				"hipaa",
			},
		},
	}

	return c.JSON(fiber.Map{
		"export_types": types,
	})
}

// CreateExport initiates a new export job
func (h *ExportHandler) CreateExport(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	var req ExportRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Validate export type
	exportType := export.ExportType(req.Type)
	switch exportType {
	case export.ExportTypeGDPR, export.ExportTypeAuditLogs, export.ExportTypeCompliance:
		// Valid types
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid export type",
		})
	}

	// Validate format
	exportFormat := export.FormatJSON
	if req.Format != "" {
		exportFormat = export.ExportFormat(req.Format)
		switch exportFormat {
		case export.FormatJSON, export.FormatCSV, export.FormatZIP:
			// Valid formats
		default:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid export format",
			})
		}
	}

	// Get organization ID from context (if available)
	orgID := ""
	if orgIDVal := c.Locals("organizationID"); orgIDVal != nil {
		orgID = orgIDVal.(string)
	}

	// Create the export job
	job, err := h.exportService.CreateExportJob(userID, orgID, exportType, exportFormat, req.Parameters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create export job",
		})
	}

	// Log the export request
	if h.auditLogger != nil {
		h.auditLogger.LogUserAction(userID, middleware.GetEmail(c), audit.ActionExport, audit.ResourceExport, job.ID,
			audit.WithIPAddress(c.IP()),
			audit.WithUserAgent(c.Get("User-Agent")),
			audit.WithOrgID(orgID),
			audit.WithMetadata(map[string]interface{}{
				"export_type": req.Type,
				"format":      req.Format,
			}),
		)
	}

	// Start the export asynchronously
	var exportFunc func(ctx context.Context, job *export.ExportJob, progressCh chan<- int) error
	switch exportType {
	case export.ExportTypeGDPR:
		if h.gdprExporter == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "GDPR export not available",
			})
		}
		exportFunc = h.gdprExporter.Export
	case export.ExportTypeCompliance:
		if h.complianceReporter == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "compliance reports not available",
			})
		}
		exportFunc = h.complianceReporter.GenerateReport
	default:
		// For audit logs and other types, use a generic handler
		exportFunc = h.exportAuditLogs
	}

	if err := h.exportService.StartExport(c.Context(), job.ID, exportFunc); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to start export",
		})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"job": job,
	})
}

// GetExport returns the status of an export job
func (h *ExportHandler) GetExport(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	jobID := c.Params("id")
	if jobID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "export id required",
		})
	}

	job, err := h.exportService.GetJob(jobID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get export",
		})
	}
	if job == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "export not found",
		})
	}

	// Verify ownership
	if job.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	return c.JSON(fiber.Map{
		"job": job,
	})
}

// ListExports returns all exports for the current user
func (h *ExportHandler) ListExports(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	limit := 20
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
			if limit > 100 {
				limit = 100
			}
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	jobs, err := h.exportService.GetUserJobs(userID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list exports",
		})
	}

	return c.JSON(fiber.Map{
		"exports": jobs,
	})
}

// DownloadExport handles export file downloads
func (h *ExportHandler) DownloadExport(c *fiber.Ctx) error {
	jobID := c.Params("id")
	downloadKey := c.Query("key")

	if jobID == "" || downloadKey == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "export id and key required",
		})
	}

	reader, filename, size, err := h.exportService.GetDownloadReader(jobID, downloadKey)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	defer reader.Close()

	// Log the download
	if h.auditLogger != nil {
		userID := middleware.GetUserID(c)
		h.auditLogger.LogUserAction(userID, middleware.GetEmail(c), audit.ActionDownload, audit.ResourceExport, jobID,
			audit.WithIPAddress(c.IP()),
			audit.WithUserAgent(c.Get("User-Agent")),
		)
	}

	c.Set("Content-Disposition", "attachment; filename="+filename)
	c.Set("Content-Type", "application/octet-stream")
	c.Set("Content-Length", strconv.FormatInt(size, 10))

	_, err = io.Copy(c.Response().BodyWriter(), reader)
	return err
}

// CancelExport cancels an in-progress export
func (h *ExportHandler) CancelExport(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	jobID := c.Params("id")
	if jobID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "export id required",
		})
	}

	// Verify ownership
	job, err := h.exportService.GetJob(jobID)
	if err != nil || job == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "export not found",
		})
	}
	if job.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	if err := h.exportService.CancelExport(jobID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "export cancelled",
	})
}

// exportAuditLogs is a generic audit log export function
func (h *ExportHandler) exportAuditLogs(ctx context.Context, job *export.ExportJob, progressCh chan<- int) error {
	if h.auditLogger == nil {
		return nil
	}

	progressCh <- 10

	// Parse parameters
	var startTime, endTime *time.Time
	if job.Parameters != nil {
		if st, ok := job.Parameters["start_date"].(string); ok {
			if t, err := time.Parse(time.RFC3339, st); err == nil {
				startTime = &t
			}
		}
		if et, ok := job.Parameters["end_date"].(string); ok {
			if t, err := time.Parse(time.RFC3339, et); err == nil {
				endTime = &t
			}
		}
	}

	// Default to last 30 days
	if startTime == nil {
		t := time.Now().AddDate(0, 0, -30)
		startTime = &t
	}
	if endTime == nil {
		t := time.Now()
		endTime = &t
	}

	progressCh <- 20

	// Query audit logs
	filter := audit.AuditFilter{
		ActorID:   job.UserID,
		StartTime: startTime,
		EndTime:   endTime,
		Limit:     10000,
	}

	events, total, err := h.auditLogger.Query(filter)
	if err != nil {
		return err
	}

	progressCh <- 60

	// Create export data
	exportData := map[string]interface{}{
		"export_info": map[string]interface{}{
			"export_id":    job.ID,
			"user_id":      job.UserID,
			"exported_at":  time.Now().UTC(),
			"period_start": startTime,
			"period_end":   endTime,
			"total_events": total,
		},
		"events": events,
	}

	progressCh <- 80

	// Write to file
	filePath := h.exportService.CreateFilePath(job.ID, job.Format)
	if err := export.WriteJSON(filePath, exportData); err != nil {
		return err
	}

	job.FilePath = filePath
	progressCh <- 100

	return nil
}

// AuditLogHandler handles audit log query endpoints
type AuditLogHandler struct {
	logger *audit.Logger
}

// NewAuditLogHandler creates a new audit log handler
func NewAuditLogHandler(logger *audit.Logger) *AuditLogHandler {
	return &AuditLogHandler{logger: logger}
}

// ListAuditLogs returns audit logs with filtering
func (h *AuditLogHandler) ListAuditLogs(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	// Build filter from query parameters
	filter := audit.AuditFilter{}

	// Users can only see their own audit logs unless they're an admin
	// For now, filter by actor_id
	if actorID := c.Query("actor_id"); actorID != "" {
		filter.ActorID = actorID
	}

	if action := c.Query("action"); action != "" {
		filter.Action = audit.ActionType(action)
	}

	if resourceType := c.Query("resource_type"); resourceType != "" {
		filter.ResourceType = audit.ResourceType(resourceType)
	}

	if resourceID := c.Query("resource_id"); resourceID != "" {
		filter.ResourceID = resourceID
	}

	if startTime := c.Query("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			filter.StartTime = &t
		}
	}

	if endTime := c.Query("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			filter.EndTime = &t
		}
	}

	if success := c.Query("success"); success != "" {
		s := success == "true"
		filter.Success = &s
	}

	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filter.Limit = l
		}
	}

	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			filter.Offset = o
		}
	}

	events, total, err := h.logger.Query(filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to query audit logs",
		})
	}

	return c.JSON(fiber.Map{
		"events": events,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// GetAuditLog returns a single audit log entry
func (h *AuditLogHandler) GetAuditLog(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "audit log id required",
		})
	}

	event, err := h.logger.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get audit log",
		})
	}
	if event == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "audit log not found",
		})
	}

	return c.JSON(fiber.Map{
		"event": event,
	})
}
