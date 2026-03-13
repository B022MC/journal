package model

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type AdmRole struct {
	Id          int64     `db:"id"`
	Code        string    `db:"code"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	IsSuper     int32     `db:"is_super"`
	Status      int32     `db:"status"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type AdmPermission struct {
	Id          int64     `db:"id"`
	Code        string    `db:"code"`
	Name        string    `db:"name"`
	Module      string    `db:"module"`
	Resource    string    `db:"resource"`
	Action      string    `db:"action"`
	Description string    `db:"description"`
	Status      int32     `db:"status"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type AdmUserRole struct {
	Id        int64     `db:"id"`
	UserId    int64     `db:"user_id"`
	RoleId    int64     `db:"role_id"`
	CreatedAt time.Time `db:"created_at"`
}

type AdmAuditLog struct {
	Id             int64     `db:"id"`
	ActorUserId    int64     `db:"actor_user_id"`
	PermissionCode string    `db:"permission_code"`
	Action         string    `db:"action"`
	TargetType     string    `db:"target_type"`
	TargetId       int64     `db:"target_id"`
	Detail         string    `db:"detail"`
	CreatedAt      time.Time `db:"created_at"`
}

// ==================== Model ====================

type AdminRBACModel struct {
	conn sqlx.SqlConn
}

func NewAdminRBACModel(conn sqlx.SqlConn) *AdminRBACModel {
	return &AdminRBACModel{conn: conn}
}

// IsSuperAdmin checks if a user has any role with is_super=1.
func (m *AdminRBACModel) IsSuperAdmin(ctx context.Context, userId int64) (bool, error) {
	var count int64
	query := `SELECT COUNT(1)
		FROM adm_user_role ur
		JOIN adm_role r ON ur.role_id = r.id AND r.status = 1 AND r.is_super = 1
		WHERE ur.user_id = ?`
	err := m.conn.QueryRowCtx(sqlx.WithReadPrimary(ctx), &count, query, userId)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ==================== Permission Check ====================

// HasPermission checks if a user has a specific permission.
// Super admins bypass all permission checks and always return true.
func (m *AdminRBACModel) HasPermission(ctx context.Context, userId int64, permissionCode string) (bool, error) {
	// Super admin bypass
	isSuper, err := m.IsSuperAdmin(ctx, userId)
	if err != nil {
		return false, err
	}
	if isSuper {
		return true, nil
	}

	var count int64
	query := `SELECT COUNT(1)
		FROM adm_user_role ur
		JOIN adm_role r ON ur.role_id = r.id AND r.status = 1
		JOIN adm_role_permission rp ON r.id = rp.role_id
		JOIN adm_permission p ON rp.permission_id = p.id AND p.status = 1
		WHERE ur.user_id = ? AND p.code = ?`
	err = m.conn.QueryRowCtx(sqlx.WithReadPrimary(ctx), &count, query, userId, permissionCode)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ListPermissionCodesByUserId returns all permission codes a user has across all roles.
func (m *AdminRBACModel) ListPermissionCodesByUserId(ctx context.Context, userId int64) ([]string, error) {
	// Check super admin first: if super, return all active permission codes
	isSuper, err := m.IsSuperAdmin(ctx, userId)
	if err != nil {
		return nil, err
	}
	if isSuper {
		return m.listAllActivePermissionCodes(ctx)
	}

	query := `SELECT DISTINCT p.code
		FROM adm_user_role ur
		JOIN adm_role r ON ur.role_id = r.id AND r.status = 1
		JOIN adm_role_permission rp ON r.id = rp.role_id
		JOIN adm_permission p ON rp.permission_id = p.id AND p.status = 1
		WHERE ur.user_id = ?
		ORDER BY p.code`

	var codes []string
	err = m.conn.QueryRowsCtx(sqlx.WithReadPrimary(ctx), &codes, query, userId)
	if err != nil {
		return nil, err
	}
	if codes == nil {
		codes = []string{}
	}
	return codes, nil
}

func (m *AdminRBACModel) listAllActivePermissionCodes(ctx context.Context) ([]string, error) {
	query := `SELECT code FROM adm_permission WHERE status = 1 ORDER BY code`
	var codes []string
	err := m.conn.QueryRowsCtx(sqlx.WithReadPrimary(ctx), &codes, query)
	if err != nil {
		return nil, err
	}
	if codes == nil {
		codes = []string{}
	}
	return codes, nil
}

// HasAnyAdminRole checks if a user has at least one active admin role.
func (m *AdminRBACModel) HasAnyAdminRole(ctx context.Context, userId int64) (bool, error) {
	var count int64
	query := `SELECT COUNT(1)
		FROM adm_user_role ur
		JOIN adm_role r ON ur.role_id = r.id AND r.status = 1
		WHERE ur.user_id = ?`
	err := m.conn.QueryRowCtx(sqlx.WithReadPrimary(ctx), &count, query, userId)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ==================== Role CRUD ====================

func (m *AdminRBACModel) ListRoles(ctx context.Context) ([]*AdmRole, error) {
	query := `SELECT id, code, name, IFNULL(description,'') as description, is_super, status, created_at, updated_at
		FROM adm_role ORDER BY id`
	var roles []*AdmRole
	err := m.conn.QueryRowsCtx(ctx, &roles, query)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (m *AdminRBACModel) GetRole(ctx context.Context, id int64) (*AdmRole, error) {
	query := `SELECT id, code, name, IFNULL(description,'') as description, is_super, status, created_at, updated_at
		FROM adm_role WHERE id = ?`
	var role AdmRole
	err := m.conn.QueryRowCtx(ctx, &role, query, id)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (m *AdminRBACModel) CreateRole(ctx context.Context, code, name, description string) (int64, error) {
	query := `INSERT INTO adm_role (code, name, description) VALUES (?, ?, ?)`
	result, err := m.conn.ExecCtx(ctx, query, code, name, description)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (m *AdminRBACModel) UpdateRole(ctx context.Context, id int64, name, description string, status int32) error {
	query := `UPDATE adm_role SET name = ?, description = ?, status = ? WHERE id = ? AND is_super = 0`
	result, err := m.conn.ExecCtx(ctx, query, name, description, status, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("role not found or is built-in super role")
	}
	return nil
}

func (m *AdminRBACModel) DeleteRole(ctx context.Context, id int64) error {
	// Prevent deletion of super admin roles
	query := `DELETE FROM adm_role WHERE id = ? AND is_super = 0`
	result, err := m.conn.ExecCtx(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("role not found or is built-in super role")
	}
	// Clean up role_permission and user_role bindings
	_, _ = m.conn.ExecCtx(ctx, `DELETE FROM adm_role_permission WHERE role_id = ?`, id)
	_, _ = m.conn.ExecCtx(ctx, `DELETE FROM adm_user_role WHERE role_id = ?`, id)
	return nil
}

// ==================== Permission Query ====================

func (m *AdminRBACModel) ListPermissions(ctx context.Context) ([]*AdmPermission, error) {
	query := `SELECT id, code, name, module, resource, action, IFNULL(description,'') as description, status, created_at, updated_at
		FROM adm_permission ORDER BY module, code`
	var perms []*AdmPermission
	err := m.conn.QueryRowsCtx(ctx, &perms, query)
	if err != nil {
		return nil, err
	}
	return perms, nil
}

func (m *AdminRBACModel) ListRolePermissionIds(ctx context.Context, roleId int64) ([]int64, error) {
	query := `SELECT permission_id FROM adm_role_permission WHERE role_id = ?`
	var ids []int64
	err := m.conn.QueryRowsCtx(ctx, &ids, query, roleId)
	if err != nil {
		return nil, err
	}
	if ids == nil {
		ids = []int64{}
	}
	return ids, nil
}

// SetRolePermissions replaces all permissions for a role (delete-then-insert).
func (m *AdminRBACModel) SetRolePermissions(ctx context.Context, roleId int64, permissionIds []int64) error {
	return m.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		_, err := session.ExecCtx(ctx, `DELETE FROM adm_role_permission WHERE role_id = ?`, roleId)
		if err != nil {
			return err
		}
		if len(permissionIds) == 0 {
			return nil
		}
		// Batch insert
		valueStrings := make([]string, 0, len(permissionIds))
		valueArgs := make([]interface{}, 0, len(permissionIds)*2)
		for _, pid := range permissionIds {
			valueStrings = append(valueStrings, "(?, ?)")
			valueArgs = append(valueArgs, roleId, pid)
		}
		insertQuery := fmt.Sprintf("INSERT INTO adm_role_permission (role_id, permission_id) VALUES %s",
			strings.Join(valueStrings, ","))
		_, err = session.ExecCtx(ctx, insertQuery, valueArgs...)
		return err
	})
}

// ==================== User-Role Association ====================

func (m *AdminRBACModel) ListUserRolesByUserId(ctx context.Context, userId int64) ([]*AdmRole, error) {
	query := `SELECT r.id, r.code, r.name, IFNULL(r.description,'') as description, r.is_super, r.status, r.created_at, r.updated_at
		FROM adm_user_role ur
		JOIN adm_role r ON ur.role_id = r.id
		WHERE ur.user_id = ?
		ORDER BY r.id`
	var roles []*AdmRole
	err := m.conn.QueryRowsCtx(ctx, &roles, query, userId)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (m *AdminRBACModel) AssignUserRole(ctx context.Context, userId, roleId int64) error {
	query := `INSERT IGNORE INTO adm_user_role (user_id, role_id) VALUES (?, ?)`
	_, err := m.conn.ExecCtx(ctx, query, userId, roleId)
	return err
}

func (m *AdminRBACModel) RevokeUserRole(ctx context.Context, userId, roleId int64) error {
	query := `DELETE FROM adm_user_role WHERE user_id = ? AND role_id = ?`
	_, err := m.conn.ExecCtx(ctx, query, userId, roleId)
	return err
}

// ==================== Audit Log ====================

func (m *AdminRBACModel) InsertAuditLog(ctx context.Context, log *AdmAuditLog) error {
	query := `INSERT INTO adm_audit_log (actor_user_id, permission_code, action, target_type, target_id, detail)
		VALUES (?, ?, ?, ?, ?, ?)`
	_, err := m.conn.ExecCtx(ctx, query,
		log.ActorUserId, log.PermissionCode, log.Action, log.TargetType, log.TargetId, log.Detail)
	return err
}

func (m *AdminRBACModel) ListAuditLogs(ctx context.Context, page, pageSize int) ([]*AdmAuditLog, int64, error) {
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, `SELECT COUNT(*) FROM adm_audit_log`)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := `SELECT id, actor_user_id, permission_code, action,
		IFNULL(target_type,'') as target_type, IFNULL(target_id,0) as target_id,
		IFNULL(detail,'') as detail, created_at
		FROM adm_audit_log ORDER BY created_at DESC LIMIT ? OFFSET ?`
	var logs []*AdmAuditLog
	err = m.conn.QueryRowsCtx(ctx, &logs, query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}
