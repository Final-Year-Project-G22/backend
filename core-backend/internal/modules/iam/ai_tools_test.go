package iam

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUserProfileTool_Name(t *testing.T) {
	tool := &GetUserProfileTool{}
	assert.Equal(t, "get_user_profile", tool.Name())
}

func TestGetUserProfileTool_Description(t *testing.T) {
	tool := &GetUserProfileTool{}
	assert.Contains(t, tool.Description(), "business profile")
}

func TestGetUserProfileTool_ParameterSchema(t *testing.T) {
	tool := &GetUserProfileTool{}
	var schema map[string]interface{}
	err := json.Unmarshal([]byte(tool.ParameterSchema()), &schema)
	assert.NoError(t, err)
	assert.Equal(t, "object", schema["type"])
}
