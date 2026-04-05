package dto

import "github.com/google/uuid"

type RoleDTO struct {
	ID          uuid.UUID `json:"id" doc:"Role identifier"`
	Code        string    `json:"code" doc:"Role code"`
	Name        string    `json:"name" doc:"Role name"`
	Description *string   `json:"description,omitempty" doc:"Role description"`
	Type        string    `json:"type" doc:"Role type"`
	IsSystem    bool      `json:"isSystem" doc:"System role"`
	IsMutable   bool      `json:"isMutable" doc:"Mutable role"`
}

type ListRolesInput struct{}

type ListRolesOutput struct {
	Body ListRolesResponseBody
}

type ListRolesResponseBody struct {
	Roles []RoleDTO `json:"roles" doc:"List of roles"`
}

type GetRoleInput struct {
	RoleID uuid.UUID `path:"roleId" doc:"Role identifier"`
}

type GetRoleResponseBody struct {
	Role        RoleDTO         `json:"role" doc:"Role details"`
	Permissions []PermissionDTO `json:"permissions" doc:"Role permissions"`
}

type GetRoleOutput struct {
	Body GetRoleResponseBody
}

type CreateRoleRequest struct {
	Code          string      `json:"code" doc:"Role code" minLength:"1" maxLength:"100"`
	Name          string      `json:"name" doc:"Role name" minLength:"1" maxLength:"100"`
	Description   *string     `json:"description,omitempty" doc:"Role description"`
	PermissionIDs []uuid.UUID `json:"permissionIds" doc:"Permissions for role"`
}

type CreateRoleInput struct {
	Body CreateRoleRequest
}

type CreateRoleResponseBody struct {
	Role RoleDTO `json:"role" doc:"Created role"`
}

type CreateRoleOutput struct {
	Body CreateRoleResponseBody
}

type UpdateRoleRequest struct {
	Name          *string     `json:"name,omitempty" doc:"Role name" minLength:"1" maxLength:"100"`
	Description   *string     `json:"description,omitempty" doc:"Role description"`
	PermissionIDs []uuid.UUID `json:"permissionIds" doc:"Permissions for role"`
}

type UpdateRoleInput struct {
	RoleID uuid.UUID `path:"roleId" doc:"Role identifier"`
	Body   UpdateRoleRequest
}

type UpdateRoleResponseBody struct {
	Role RoleDTO `json:"role" doc:"Updated role"`
}

type UpdateRoleOutput struct {
	Body UpdateRoleResponseBody
}

type DeleteRoleInput struct {
	RoleID uuid.UUID `path:"roleId" doc:"Role identifier"`
}

type DeleteRoleOutput struct {
	Body struct {
		Message string `json:"message" doc:"Status message"`
	}
}
