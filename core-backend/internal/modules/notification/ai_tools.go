package notification

import (
	"context"
	"encoding/json"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	"github.com/google/uuid"
)

// CheckComplianceStatusTool returns the user's compliance entry deadlines and calendar.
type CheckComplianceStatusTool struct {
	complianceUC usecase.ComplianceEntryUsecase
}

func NewCheckComplianceStatusTool(complianceUC usecase.ComplianceEntryUsecase) *CheckComplianceStatusTool {
	return &CheckComplianceStatusTool{complianceUC: complianceUC}
}

func (t *CheckComplianceStatusTool) Name() string { return "check_compliance_status" }

func (t *CheckComplianceStatusTool) Description() string {
	return "Check the user's compliance deadlines including trade license, TIN, and business registration expiry dates with reminder windows."
}

func (t *CheckComplianceStatusTool) ParameterSchema() string {
	return `{
		"type": "object",
		"properties": {}
	}`
}

func (t *CheckComplianceStatusTool) Execute(ctx context.Context, _ string, accountID, _ uuid.UUID) (string, error) {
	calendar, err := t.complianceUC.GetCalendar(ctx, accountID)
	if err != nil {
		return "", err
	}

	type complianceItem struct {
		Type     string `json:"type"`
		Title    string `json:"title"`
		DueDate  string `json:"dueDate"`
		DaysLeft int    `json:"daysLeft"`
		Urgency  string `json:"urgency"`
	}

	items := make([]complianceItem, 0, len(calendar.Entries))
	for _, entry := range calendar.Entries {
		urgency := "normal"
		if entry.DaysRemaining <= 7 {
			urgency = "critical"
		} else if entry.DaysRemaining <= 30 {
			urgency = "soon"
		}
		items = append(items, complianceItem{
			Type:     entry.Type,
			Title:    entry.Title,
			DueDate:  entry.Date.Format("2006-01-02"),
			DaysLeft: entry.DaysRemaining,
			Urgency:  urgency,
		})
	}

	result := map[string]interface{}{
		"entries": items,
		"total":   len(items),
	}

	payload, _ := json.Marshal(result)
	return string(payload), nil
}
