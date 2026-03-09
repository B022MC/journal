package model

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type AdminRBACModel struct {
	conn sqlx.SqlConn
}

func NewAdminRBACModel(conn sqlx.SqlConn) *AdminRBACModel {
	return &AdminRBACModel{conn: conn}
}

func (m *AdminRBACModel) ListPermissionCodesByUserId(ctx context.Context, userId int64) ([]string, error) {
	query := `SELECT DISTINCT p.code
		FROM adm_user_role ur
		JOIN adm_role r ON ur.role_id = r.id AND r.status = 1
		JOIN adm_role_permission rp ON r.id = rp.role_id
		JOIN adm_permission p ON rp.permission_id = p.id AND p.status = 1
		WHERE ur.user_id = ?
		ORDER BY p.code`

	var codes []string
	err := m.conn.QueryRowsCtx(sqlx.WithReadPrimary(ctx), &codes, query, userId)
	if err != nil {
		return nil, err
	}
	if codes == nil {
		codes = []string{}
	}

	return codes, nil
}

func (m *AdminRBACModel) HasPermission(ctx context.Context, userId int64, permissionCode string) (bool, error) {
	var count int64
	query := `SELECT COUNT(1)
		FROM adm_user_role ur
		JOIN adm_role r ON ur.role_id = r.id AND r.status = 1
		JOIN adm_role_permission rp ON r.id = rp.role_id
		JOIN adm_permission p ON rp.permission_id = p.id AND p.status = 1
		WHERE ur.user_id = ? AND p.code = ?`

	err := m.conn.QueryRowCtx(sqlx.WithReadPrimary(ctx), &count, query, userId, permissionCode)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
