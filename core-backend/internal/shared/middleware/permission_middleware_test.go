package middleware

import (
	"context"
	"crypto/tls"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/usecase"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

func TestHasRole(t *testing.T) {
	superAdmin := &entity.Role{Code: "super_admin"}
	iamAdmin := &entity.Role{Code: "iam_admin"}
	libraryAdmin := &entity.Role{Code: "library_admin"}

	tests := []struct {
		name     string
		roles    []*entity.Role
		allowed  []string
		expected bool
	}{
		{
			name:     "empty roles and empty allowed",
			roles:    []*entity.Role{},
			allowed:  []string{},
			expected: false,
		},
		{
			name:     "exact match single role",
			roles:    []*entity.Role{superAdmin},
			allowed:  []string{"super_admin"},
			expected: true,
		},
		{
			name:     "case insensitive match",
			roles:    []*entity.Role{{Code: "Super_Admin"}},
			allowed:  []string{"super_admin"},
			expected: true,
		},
		{
			name:     "multiple allowed roles, one matches",
			roles:    []*entity.Role{iamAdmin},
			allowed:  []string{"super_admin", "iam_admin"},
			expected: true,
		},
		{
			name:     "no match",
			roles:    []*entity.Role{libraryAdmin},
			allowed:  []string{"super_admin"},
			expected: false,
		},
		{
			name:     "multiple roles, one matches",
			roles:    []*entity.Role{iamAdmin, libraryAdmin},
			allowed:  []string{"super_admin", "library_admin"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasRole(tt.roles, tt.allowed)
			if got != tt.expected {
				t.Errorf("hasRole(%v, %v) = %v, want %v", tt.roles, tt.allowed, got, tt.expected)
			}
		})
	}
}

func TestPermissionMiddleware(t *testing.T) {
	validAccountID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	nilUUID := uuid.UUID{}

	tests := []struct {
		name           string
		accountID      uuid.UUID
		permissionName string
		allowedRoles   []string
		hasPermission  bool
		hasRole        bool
		nextCalled     bool
	}{
		{
			name:           "missing account context returns early",
			accountID:      nilUUID,
			permissionName: "test.read",
			allowedRoles:   nil,
			hasPermission:  false,
			hasRole:        false,
			nextCalled:     false,
		},
		{
			name:           "has permission calls next",
			accountID:      validAccountID,
			permissionName: "test.read",
			allowedRoles:   nil,
			hasPermission:  true,
			hasRole:        false,
			nextCalled:     true,
		},
		{
			name:           "no permission and no role denies access",
			accountID:      validAccountID,
			permissionName: "test.read",
			allowedRoles:   nil,
			hasPermission:  false,
			hasRole:        false,
			nextCalled:     false,
		},
		{
			name:           "allowed role bypass calls next even without permission",
			accountID:      validAccountID,
			permissionName: "test.read",
			allowedRoles:   []string{"super_admin"},
			hasPermission:  false,
			hasRole:        true,
			nextCalled:     true,
		},
		{
			name:           "allowed role but no permission match does not bypass",
			accountID:      validAccountID,
			permissionName: "test.read",
			allowedRoles:   []string{"super_admin"},
			hasPermission:  false,
			hasRole:        false,
			nextCalled:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &mockAPI{}
			usecase := &mockRoleAssignmentUsecase{
				hasPermission: tt.hasPermission,
				hasRole:       tt.hasRole,
			}

			ctx := context.WithValue(context.Background(), contextkeys.AccountID, tt.accountID)

			middlewareFunc := PermissionMiddleware(api, usecase, tt.permissionName, tt.allowedRoles)
			mockCtx := &mockContext{ctx: ctx}

			nextCalled := false
			next := func(ctx huma.Context) {
				nextCalled = true
			}

			middlewareFunc(mockCtx, next)

			if nextCalled != tt.nextCalled {
				t.Errorf("next called = %v, want %v", nextCalled, tt.nextCalled)
			}
		})
	}
}

type mockRoleAssignmentUsecase struct {
	usecase.RoleAssignmentUsecase
	hasPermission bool
	hasRole       bool
}

func (m *mockRoleAssignmentUsecase) HasPermission(ctx context.Context, accountID uuid.UUID, permissionCode string) (bool, error) {
	return m.hasPermission, nil
}

func (m *mockRoleAssignmentUsecase) ListAccountRoles(ctx context.Context, accountID uuid.UUID) ([]*entity.Role, error) {
	if m.hasRole {
		return []*entity.Role{{Code: "super_admin"}}, nil
	}
	return []*entity.Role{}, nil
}

type mockAPI struct{}

func (m *mockAPI) Negotiate(accept string) (string, error) {
	return "application/json", nil
}

func (m *mockAPI) Transform(ctx huma.Context, status string, v any) (any, error) {
	return v, nil
}

func (m *mockAPI) Marshal(w io.Writer, contentType string, v any) error {
	return nil
}

func (m *mockAPI) Unmarshal(contentType string, data []byte, v any) error {
	return nil
}

func (m *mockAPI) Adapter() huma.Adapter {
	return &mockAdapter{}
}

func (m *mockAPI) OpenAPI() *huma.OpenAPI {
	return nil
}

func (m *mockAPI) UseMiddleware(middlewares ...func(ctx huma.Context, next func(huma.Context))) {}
func (m *mockAPI) Middlewares() huma.Middlewares                                                { return huma.Middlewares{} }

type mockAdapter struct{}

func (a *mockAdapter) Handle(op *huma.Operation, handler func(ctx huma.Context)) {}
func (a *mockAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request)          {}

type mockContext struct {
	ctx context.Context
}

func (c *mockContext) Context() context.Context {
	return c.ctx
}

func (c *mockContext) Operation() *huma.Operation {
	return nil
}

func (c *mockContext) TLS() *tls.ConnectionState {
	return nil
}

func (c *mockContext) Method() string {
	return "GET"
}

func (c *mockContext) Host() string {
	return "localhost"
}

func (c *mockContext) RemoteAddr() string {
	return "127.0.0.1"
}

func (c *mockContext) URL() url.URL {
	return url.URL{Path: "/test"}
}

func (c *mockContext) Param(name string) string {
	return ""
}

func (c *mockContext) Query(name string) string {
	return ""
}

func (c *mockContext) Header(name string) string {
	return ""
}

func (c *mockContext) EachHeader(cb func(name, value string)) {}

func (c *mockContext) BodyReader() io.Reader {
	return nil
}

func (c *mockContext) GetMultipartForm() (*multipart.Form, error) {
	return nil, nil
}

func (c *mockContext) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *mockContext) SetStatus(code int) {}

func (c *mockContext) Status() int {
	return 200
}

func (c *mockContext) SetHeader(name, value string) {}

func (c *mockContext) AppendHeader(name, value string) {}

func (c *mockContext) BodyWriter() io.Writer {
	return io.Discard
}

func (c *mockContext) Version() huma.ProtoVersion {
	return huma.ProtoVersion{}
}
