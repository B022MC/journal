package consts

// ==================== RBAC Role Codes ====================

const (
	RoleCodeSuperAdmin = "super_admin"
)

// ==================== RBAC Permission Codes ====================

const (
	PermAdminDashboardView = "admin.dashboard.view"
	PermAdminKeywordView   = "admin.keyword.view"
	PermAdminKeywordManage = "admin.keyword.manage"
	PermAdminNewsView      = "admin.news.view"
	PermAdminNewsCreate    = "admin.news.create"
	PermAdminNewsUpdate    = "admin.news.update"
	PermAdminNewsDelete    = "admin.news.delete"
	PermAdminPaperView     = "admin.paper.view"
	PermAdminPaperZoneUp   = "admin.paper.zone.update"
	PermAdminUserView      = "admin.user.view"
	PermAdminUserManage    = "admin.user.manage"
	PermAdminRoleView      = "admin.role.view"
	PermAdminRoleManage    = "admin.role.manage"
	PermAdminAuditView     = "admin.audit.view"
)

// ==================== RBAC Status ====================

const (
	RBACStatusDisabled = 0
	RBACStatusActive   = 1
)
