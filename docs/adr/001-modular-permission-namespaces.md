# ADR 001: Modular Permission Namespaces

## Status
Accepted

## Date
2026-05-26

## Context
The RBAC system originally defined all permissions under a single `iam.*` namespace — 14 permissions total, all seeded in the IAM module. Non-IAM modules (library, community, guide, notification) had inconsistent authorization coverage:

- **Library and Community** reused generic `iam.read`/`iam.write`/`iam.update`/`iam.delete` as CRUD proxies for their own admin endpoints. Semantically incorrect — `iam.read` means "read IAM resources," not "read library categories."
- **Guide and Notification** had **no permission middleware at all** — only authentication and account-status checks guarded admin endpoints.
- **Taxonomy admin routes** (IAM sub-routes) also had zero permission checks.

The existing Permission entity already had a `Module` field, but it was never leveraged for per-module isolation. All modules depended on IAM's central permission constants.

## Decision
We will implement a modular permission system where each module defines its own permission namespace and registers seeds at application startup. The following specific decisions were reached through a 16-question grill session:

### Per-Module Permission Namespaces (Q1, Q4)
Each module defines its own permission codes using a `module.action` pattern (e.g., `library.read`, `library.write`, `guide.read`, `guide.write`, `community.read`, `community.write`). Module-level CRUD (`.read`, `.write`, `.update`, `.delete`) provides adequate granularity for the current scope.

The existing `iam.read`/`iam.write`/`iam.update`/`iam.delete` permissions continue to exist solely for IAM operations (role listing, permission listing, account management), but library and community stop using them as proxies.

### Global Roles (Q2)
Roles remain **global** — stored in the IAM module, but can hold permissions from any module. This avoids coupling between modules and keeps role management centralized.

### Seed Discovery via FX Group Injection (Q3, Q8, Q9)
Each module provides its permission seeds via FX's group injection mechanism:
- `group:"permission_seeds"` — each module provides `[]SeedPermission`
- `group:"role_seeds"` — each module optionally provides `[]SeedRole` with explicit `PermissionCodes`

The `SeedPermission` and `SeedRole` structs live in `internal/shared/permissions/` so they can be imported by all modules without creating circular dependencies. The IAM module's `RolePermissionSeeder` collects all groups on startup and:

1. Upserts each permission (idempotent — skips existing codes)
2. Upserts each role
3. Assigns `super_admin` ALL permissions from all modules automatically
4. Assigns each role's declared `PermissionCodes` to that role

### Permission Middleware Coverage (Q5, Q10, Q15)
All modules with admin endpoints — including those currently without checks (guide, notification, taxonomy) — get permission middleware applied uniformly. Each module's middleware passes its own permission constant (e.g., `library.read`) and `["super_admin"]` as the role bypass list.

Taxonomy admin routes (IAM sub-routes) use the existing `iam.write`/`iam.update` permissions rather than creating a new `taxonomy.*` namespace.

### Role Definitions (Q7, Q12, Q16)
- **super_admin** — system role, automatically receives all permissions from all modules
- **iam_admin** — system role, receives explicit list of `iam.*` permissions only; modules may optionally define their own roles (e.g., `library_admin` with `library.*` permissions)
- No magic exclusion logic — each role's permission assignment is an explicit list

### Migration (Q11)
Direct swap — no dual-gating migration window. Library and community routes switch from `iam.read`/`iam.write` to `library.read`/`library.write` immediately. Since `super_admin` automatically receives all new permissions, and `iam_admin` is IAM-only, no disruption occurs.

## Consequences

### Positive
- **Authorization co-located with domain code** — each module's permission constants, seeds, and middleware live alongside the code they protect. No cross-module dependency on IAM for authorization strings.
- **Clear semantic meaning** — `library.read` means "read library resources," not "read IAM resources for library purposes."
- **Uniform coverage** — every admin endpoint across all modules has consistent permission middleware, eliminating blind spots.
- **Role bypass as standard pattern** — all modules uniformly pass `["super_admin"]` as the bypass list, making super_admin traversal fast and consistent.
- **Self-documenting seeds** — the permission and role seeds act as a declarative manifest of what each module authorizes and which roles are available.

### Negative
- **More seed functions to maintain** — each module now manages its own seed data. However, this is proportional to module complexity and the seed definitions are trivial (4 constants per module).
- **Potential for namespace collisions** — two modules could accidentally define the same permission code. Mitigated by consistent `module.action` naming convention and code review.
- **One-time migration of library/community routes** — RouteDependencies structs and middleware wiring need to switch from `permissions.IAMRead` to `LibraryRead`.

### Dependencies
- Relies on `go.uber.org/fx` group injection (`fx.ResultTags`, `fx.ParamTags`) for seed discovery
- Relies on existing `PermissionRepository`, `RoleRepository`, `RolePermissionRepository` in IAM module
- Relies on existing `shared/middleware.PermissionMiddleware` signature (unchanged)
- Existing `iam.read`/`iam.write`/`iam.update`/`iam.delete` permissions remain in the database for IAM CRUD

## References
- 16-question grill session covering namespace strategy, role scope, injection mechanism, granularity, coverage, migration, and role definitions
- Existing permission middleware: `internal/shared/middleware/permission_middleware.go`
- Existing seeder: `internal/modules/iam/application/service/role_permission_seeder.go`
- Existing IAM module: `internal/modules/iam/module.go`
- CONTEXT.md permission terms section (added 2026-05-26)
