package service

import (
	"fmt"
	"strings"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
)

type TemplateRenderer struct{}

func NewTemplateRenderer() *TemplateRenderer {
	return &TemplateRenderer{}
}

func (r *TemplateRenderer) Render(content map[string]interface{}, variables map[string]string) (map[string]interface{}, error) {
	rendered := make(map[string]interface{}, len(content))
	for channel, channelContent := range content {
		renderedChannel, err := r.renderValue(channelContent, variables)
		if err != nil {
			return nil, fmt.Errorf("failed to render channel %s: %w", channel, err)
		}
		rendered[channel] = renderedChannel
	}
	return rendered, nil
}

func (r *TemplateRenderer) RenderMultiChannel(content map[string]interface{}, variables map[string]string, channels []entity.Channel) (map[string]interface{}, error) {
	rendered := make(map[string]interface{}, len(channels))
	for _, channel := range channels {
		channelContent, ok := content[string(channel)]
		if !ok {
			continue
		}
		renderedChannel, err := r.renderValue(channelContent, variables)
		if err != nil {
			return nil, fmt.Errorf("failed to render channel %s: %w", channel, err)
		}
		rendered[string(channel)] = renderedChannel
	}
	return rendered, nil
}

func (r *TemplateRenderer) ValidateVariables(variablesSchema map[string]interface{}, variables map[string]string) error {
	required, ok := variablesSchema["required"]
	if !ok {
		return nil
	}
	requiredList, ok := required.([]interface{})
	if !ok {
		return nil
	}
	for _, req := range requiredList {
		reqStr, ok := req.(string)
		if !ok {
			continue
		}
		if _, exists := variables[reqStr]; !exists {
			return fmt.Errorf("missing required variable: %s", reqStr)
		}
	}
	return nil
}

func (r *TemplateRenderer) renderValue(value interface{}, variables map[string]string) (interface{}, error) {
	switch v := value.(type) {
	case string:
		return r.replaceVariables(v, variables), nil
	case map[string]interface{}:
		result := make(map[string]interface{}, len(v))
		for key, val := range v {
			renderedVal, err := r.renderValue(val, variables)
			if err != nil {
				return nil, err
			}
			result[key] = renderedVal
		}
		return result, nil
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			renderedItem, err := r.renderValue(item, variables)
			if err != nil {
				return nil, err
			}
			result[i] = renderedItem
		}
		return result, nil
	default:
		return value, nil
	}
}

func (r *TemplateRenderer) replaceVariables(template string, variables map[string]string) string {
	result := template
	for key, value := range variables {
		placeholder := "{{" + key + "}}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}
