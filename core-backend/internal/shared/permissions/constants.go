package permissions

// IAM module permissions follow the pattern: module.action
// These constants are used for both the permission code and name in the database.

// Generic IAM permissions
const (
	IAMRead   = "iam.read"
	IAMWrite  = "iam.write"
	IAMUpdate = "iam.update"
	IAMDelete = "iam.delete"
)

// Admin management permissions
const (
	AdminList          = "iam.admin.list"
	AdminRead          = "iam.admin.read"
	AdminCreate        = "iam.admin.create"
	AdminRolesUpdate   = "iam.admin.roles.update"
	AdminResetPassword = "iam.admin.reset_password"
	AdminStatusUpdate  = "iam.admin.status.update"
)

// Role management permissions
const (
	RoleRead   = "iam.role.read"
	RoleCreate = "iam.role.create"
	RoleUpdate = "iam.role.update"
	RoleDelete = "iam.role.delete"
)

// Permission catalog
const (
	PermissionRead = "iam.permission.read"
)
