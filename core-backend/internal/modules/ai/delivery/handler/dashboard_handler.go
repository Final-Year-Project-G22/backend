package handler

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"gorm.io/gorm"
)

type AIDashboardHandler struct {
	db     *core.Database
	logger core.Logger
}

func NewAIDashboardHandler(db *core.Database, logger core.Logger) *AIDashboardHandler {
	return &AIDashboardHandler{db: db, logger: logger}
}

func (h *AIDashboardHandler) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return h.db.WithContext(ctx)
}

type DocumentStatsResponse struct {
	Total     int64 `json:"total" doc:"Total documents"`
	Completed int64 `json:"completed" doc:"Completed documents"`
	Failed    int64 `json:"failed" doc:"Failed documents"`
	Trend     int64 `json:"trend" doc:"Documents created this month"`
}

func (h *AIDashboardHandler) HandleDocumentStats(ctx context.Context, input *struct{}) (*struct{ Body DocumentStatsResponse }, error) {
	var total int64
	h.getDB(ctx).Table("ingestion_documents").Where("deleted_at IS NULL").Count(&total)

	var completed int64
	h.getDB(ctx).Table("ingestion_documents").Where("status = 'completed' AND deleted_at IS NULL").Count(&completed)

	var failed int64
	h.getDB(ctx).Table("ingestion_documents").Where("status = 'failed' AND deleted_at IS NULL").Count(&failed)

	var thisMonth int64
	h.getDB(ctx).Table("ingestion_documents").Where("created_at >= date_trunc('month', NOW()) AND deleted_at IS NULL").Count(&thisMonth)

	return &struct{ Body DocumentStatsResponse }{
		Body: DocumentStatsResponse{
			Total:     total,
			Completed: completed,
			Failed:    failed,
			Trend:     thisMonth,
		},
	}, nil
}

type DocumentStageCount struct {
	Stage      string `json:"stage" doc:"Pipeline stage name"`
	Count      int64  `json:"count" doc:"Documents in this stage"`
	Percentage int64  `json:"percentage" doc:"Percentage of total"`
}

type DocumentStagesResponse struct {
	Data []DocumentStageCount `json:"data"`
}

func (h *AIDashboardHandler) HandleDocumentStages(ctx context.Context, input *struct{}) (*struct{ Body DocumentStagesResponse }, error) {
	type stageRow struct {
		Stage string
		Count int64
	}
	var rows []stageRow
	h.getDB(ctx).Table("ingestion_documents").
		Select("COALESCE(status, 'unknown') AS stage, COUNT(*) AS count").
		Where("deleted_at IS NULL").
		Group("status").
		Order("count DESC").
		Scan(&rows)

	var total int64
	for _, r := range rows {
		total += r.Count
	}

	data := make([]DocumentStageCount, 0, len(rows))
	for _, r := range rows {
		pct := int64(0)
		if total > 0 {
			pct = r.Count * 100 / total
		}
		data = append(data, DocumentStageCount{
			Stage:      r.Stage,
			Count:      r.Count,
			Percentage: pct,
		})
	}

	return &struct{ Body DocumentStagesResponse }{Body: DocumentStagesResponse{Data: data}}, nil
}
