package dto

import "github.com/google/uuid"

type PermissionDTO struct {
	ID          uuid.UUID `json:"id" doc:"Permission identifier"`
	Code        string    `json:"code" doc:"Permission code"`
	Name        string    `json:"name" doc:"Permission name"`
	Description *string   `json:"description,omitempty" doc:"Permission description"`
	Module      string    `json:"module" doc:"Permission module"`
}

type ListPermissionsInput struct {
	Module string   `query:"module" doc:"Filter by module"`
	Codes  []string `query:"code" doc:"Filter by permission code"`
}

type ListPermissionsOutput struct {
	Body ListPermissionsResponseBody
}

type ListPermissionsResponseBody struct {
	Permissions []PermissionDTO `json:"permissions" doc:"List of permissions"`
}
