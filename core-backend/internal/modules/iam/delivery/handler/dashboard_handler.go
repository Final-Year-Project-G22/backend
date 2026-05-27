package handler

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DashboardHandler struct {
	db     *core.Database
	logger core.Logger
}

func NewDashboardHandler(db *core.Database, logger core.Logger) *DashboardHandler {
	return &DashboardHandler{db: db, logger: logger}
}

func (h *DashboardHandler) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return h.db.WithContext(ctx)
}

type UserStatsResponse struct {
	Total        int64   `json:"total" doc:"Total number of accounts"`
	TrendPercent float64 `json:"trendPercent" doc:"Percentage change vs previous period"`
}

func (h *DashboardHandler) HandleUserStats(ctx context.Context, input *struct{}) (*struct{ Body UserStatsResponse }, error) {
	var total int64
	if err := h.getDB(ctx).Model(&struct{}{}).Table("accounts").Where("deleted_at IS NULL").Count(&total).Error; err != nil {
		return nil, err
	}

	lastMonth := time.Now().AddDate(0, -1, 0)
	var previous int64
	h.getDB(ctx).Model(&struct{}{}).Table("accounts").Where("deleted_at IS NULL AND created_at < ?", lastMonth).Count(&previous)

	trend := 0.0
	if previous > 0 {
		trend = (float64(total-previous) / float64(previous)) * 100
	}

	return &struct{ Body UserStatsResponse }{Body: UserStatsResponse{Total: total, TrendPercent: trend}}, nil
}

type SessionStatsResponse struct {
	Total    int64 `json:"total" doc:"Current active sessions"`
	DailyAvg int64 `json:"dailyAvg" doc:"Daily average active sessions"`
}

func (h *DashboardHandler) HandleSessionStats(ctx context.Context, input *struct{}) (*struct{ Body SessionStatsResponse }, error) {
	var total int64
	h.getDB(ctx).Model(&struct{}{}).Table("sessions").
		Where("revoked_at IS NULL AND expires_at > NOW()").
		Count(&total)

	// Count sessions active in the last 24 hours for daily avg
	var daily int64
	h.getDB(ctx).Model(&struct{}{}).Table("sessions").
		Where("revoked_at IS NULL AND last_active_at > NOW() - INTERVAL '24 hours'").
		Count(&daily)

	return &struct{ Body SessionStatsResponse }{Body: SessionStatsResponse{Total: total, DailyAvg: daily}}, nil
}

type UserGrowthPoint struct {
	Period  string `json:"period" doc:"Period label (e.g. Jan, Feb)"`
	Free    int64  `json:"free" doc:"Users without active subscription"`
	Premium int64  `json:"premium" doc:"Users with active subscription"`
}

type UserGrowthInput struct {
	Period string `query:"period" doc:"Grouping period: monthly, quarterly, yearly" enum:"monthly,quarterly,yearly"`
}

type UserGrowthResponse struct {
	Data []UserGrowthPoint `json:"data"`
}

func (h *DashboardHandler) HandleUserGrowth(ctx context.Context, input *UserGrowthInput) (*struct{ Body UserGrowthResponse }, error) {
	if input.Period == "" {
		input.Period = "monthly"
	}

	// Query accounts created in the last 8 periods
	// For simplicity, use monthly grouping
	rows, err := h.getDB(ctx).Raw(`
		SELECT
			to_char(a.created_at, 'Mon') AS period,
			COUNT(a.id) FILTER (WHERE s.id IS NULL OR s.status != 'active') AS free,
			COUNT(a.id) FILTER (WHERE s.id IS NOT NULL AND s.status = 'active') AS premium
		FROM accounts a
		LEFT JOIN LATERAL (
			SELECT id, status FROM subscriptions
			WHERE account_id = a.id AND status = 'active' AND current_period_end > NOW()
			ORDER BY created_at DESC LIMIT 1
		) s ON true
		WHERE a.deleted_at IS NULL
		GROUP BY to_char(a.created_at, 'Mon'), EXTRACT(MONTH FROM a.created_at)
		ORDER BY EXTRACT(MONTH FROM a.created_at) DESC
		LIMIT 8
	`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []UserGrowthPoint
	for rows.Next() {
		var p UserGrowthPoint
		if err := rows.Scan(&p.Period, &p.Free, &p.Premium); err != nil {
			return nil, err
		}
		data = append(data, p)
	}
	// Reverse to chronological order
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}

	return &struct{ Body UserGrowthResponse }{Body: UserGrowthResponse{Data: data}}, nil
}

type SystemOverviewItem struct {
	Label string `json:"label" doc:"Item label"`
	Value string `json:"value" doc:"Item value"`
	Type  string `json:"type" doc:"Indicator type: success, warning, danger"`
}

type SystemOverviewResponse struct {
	Items []SystemOverviewItem `json:"items"`
}

func (h *DashboardHandler) HandleSystemOverview(ctx context.Context, input *struct{}) (*struct{ Body SystemOverviewResponse }, error) {
	var pendingReports int64
	h.getDB(ctx).Model(&struct{}{}).Table("content_reports").Where("status = 'pending'").Count(&pendingReports)

	var failedIngestion int64
	h.getDB(ctx).Model(&struct{}{}).Table("ingestion_documents").Where("status = 'failed'").Count(&failedIngestion)

	var activeSubs int64
	h.getDB(ctx).Model(&struct{}{}).Table("subscriptions").Where("status = 'active' AND current_period_end > NOW()").Count(&activeSubs)

	var totalAccounts int64
	h.getDB(ctx).Model(&struct{}{}).Table("accounts").Where("deleted_at IS NULL").Count(&totalAccounts)

	items := []SystemOverviewItem{
		{Label: "Pending Reports", Value: formatInt(pendingReports), Type: "warning"},
		{Label: "Failed Ingestions", Value: formatInt(failedIngestion), Type: "danger"},
		{Label: "Active Subscriptions", Value: formatInt(activeSubs), Type: "success"},
		{Label: "Total Accounts", Value: formatInt(totalAccounts), Type: "success"},
	}
	return &struct{ Body SystemOverviewResponse }{Body: SystemOverviewResponse{Items: items}}, nil
}

type ActivityLogEntry struct {
	AdminName string    `json:"adminName" doc:"Admin display name"`
	Action    string    `json:"action" doc:"Action description"`
	Target    string    `json:"target" doc:"Target entity"`
	Timestamp time.Time `json:"timestamp" doc:"When the action occurred"`
}

type ActivityLogsInput struct {
	Limit int `query:"limit" doc:"Max entries to return" default:"10"`
}

type ActivityLogsResponse struct {
	Data []ActivityLogEntry `json:"data"`
}

func (h *DashboardHandler) HandleActivityLogs(ctx context.Context, input *ActivityLogsInput) (*struct{ Body ActivityLogsResponse }, error) {
	if input.Limit <= 0 {
		input.Limit = 10
	}

	// Union of recent admin actions from content_reports, ingestion_documents
	type rawLog struct {
		AdminName string
		Action    string
		Target    string
		Timestamp time.Time
	}

	var logs []rawLog
	h.getDB(ctx).Raw(`
		(SELECT
			COALESCE(u.first_name || ' ' || u.last_name, 'System') AS admin_name,
			'Resolved report' AS action,
			cr.status AS target,
			cr.resolved_at AS timestamp
		FROM content_reports cr
		JOIN accounts a ON a.id = cr.resolved_by_account_id
		JOIN users u ON u.id = a.user_id
		WHERE cr.resolved_by_account_id IS NOT NULL)

		UNION ALL

		(SELECT
			COALESCE(u.first_name || ' ' || u.last_name, 'System') AS admin_name,
			'Uploaded document' AS action,
			COALESCE(id.source_filename::text, 'Unknown') AS target,
			id.created_at AS timestamp
		FROM ingestion_documents id
		JOIN accounts a ON a.id = id.account_id
		JOIN users u ON u.id = a.user_id)

		ORDER BY timestamp DESC
		LIMIT ?
	`, input.Limit).Scan(&logs)

	data := make([]ActivityLogEntry, 0, len(logs))
	for _, l := range logs {
		data = append(data, ActivityLogEntry(l))
	}

	return &struct{ Body ActivityLogsResponse }{Body: ActivityLogsResponse{Data: data}}, nil
}

func formatInt(n int64) string {
	if n >= 1000000 {
		return formatFloat(float64(n)/1000000, 1) + "M"
	}
	if n >= 1000 {
		return formatFloat(float64(n)/1000, 1) + "K"
	}
	return formatInt64(n)
}

func formatFloat(f float64, decimals int) string {
	scale := 1.0
	for i := 0; i < decimals; i++ {
		scale *= 10
	}
	val := int64(f * scale)
	intPart := val / int64(scale)
	fracPart := val % int64(scale)
	if fracPart == 0 {
		return formatInt64(intPart)
	}
	return formatInt64(intPart) + "." + formatInt64(fracPart)
}

func formatInt64(n int64) string {
	if n == 0 {
		return "0"
	}
	abs := n
	if abs < 0 {
		abs = -abs
	}
	var buf [20]byte
	i := len(buf)
	for abs > 0 {
		i--
		buf[i] = byte('0' + abs%10)
		abs /= 10
	}
	if n < 0 {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

type ReportStatsResponse struct {
	Pending  int64   `json:"pending" doc:"Pending reports"`
	Resolved int64   `json:"resolved" doc:"Resolved reports"`
	Total    int64   `json:"total" doc:"Total reports"`
	Trend    float64 `json:"trendPercent" doc:"Percentage change vs previous period"`
}

func (h *DashboardHandler) HandleReportStats(ctx context.Context, input *struct{}) (*struct{ Body ReportStatsResponse }, error) {
	var total int64
	h.getDB(ctx).Table("content_reports").Count(&total)

	var pending int64
	h.getDB(ctx).Table("content_reports").Where("status = 'pending'").Count(&pending)

	var resolved int64
	h.getDB(ctx).Table("content_reports").Where("status = 'resolved'").Count(&resolved)

	// Trend: comparison with last month
	lastMonth := time.Now().AddDate(0, -1, 0)
	var previous int64
	h.getDB(ctx).Table("content_reports").Where("created_at < ?", lastMonth).Count(&previous)

	trend := 0.0
	if previous > 0 {
		trend = (float64(total-previous) / float64(previous)) * 100
	}

	return &struct{ Body ReportStatsResponse }{Body: ReportStatsResponse{
		Pending: pending, Resolved: resolved, Total: total, Trend: trend,
	}}, nil
}

var _ = uuid.UUID{}
