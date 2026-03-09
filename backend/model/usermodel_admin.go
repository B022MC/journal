package model

import (
	"context"
	"fmt"
)

// ListUsersPaginatedAdmin returns a paginated list of users for admin
func (m *UserModel) ListUsersPaginatedAdmin(ctx context.Context, page, pageSize int) ([]*User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, "SELECT COUNT(*) FROM `user`")
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf("SELECT %s FROM `user` ORDER BY id DESC LIMIT ? OFFSET ?", userSelectCols)
	var users []*User
	err = m.conn.QueryRowsCtx(ctx, &users, query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// UpdateUserStatus updates user status
func (m *UserModel) UpdateUserStatus(ctx context.Context, userId int64, status int32) error {
	query := "UPDATE `user` SET `status` = ? WHERE `id` = ?"
	_, err := m.conn.ExecCtx(ctx, query, status, userId)
	return err
}

// UpdateUserRoleAdmin updates user role
func (m *UserModel) UpdateUserRoleAdmin(ctx context.Context, userId int64, role int32) error {
	query := "UPDATE `user` SET `role` = ? WHERE `id` = ?"
	_, err := m.conn.ExecCtx(ctx, query, role, userId)
	return err
}
