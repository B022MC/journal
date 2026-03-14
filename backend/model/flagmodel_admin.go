package model

import (
	"context"
	"fmt"
)

var flagSelectCols = "`id`,`target_type`,`target_id`,`reporter_id`,`reason`,`detail`,`reporter_contribution`,`status`,`created_at`"

// ListFlagsPaginated returns a paginated list of flags
func (m *FlagModel) ListFlagsPaginated(ctx context.Context, page, pageSize int, status int32) ([]*Flag, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	whereClause := ""
	var args []interface{}

	if status >= 0 {
		whereClause = "WHERE status = ?"
		args = append(args, status)
	}

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM `biz_flag` %s", whereClause)
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf("SELECT %s FROM `biz_flag` %s ORDER BY id DESC LIMIT ? OFFSET ?", flagSelectCols, whereClause)
	args = append(args, pageSize, offset)

	var flags []*Flag
	err = m.conn.QueryRowsCtx(ctx, &flags, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return flags, total, nil
}

// UpdateFlagStatus updates flag status
func (m *FlagModel) UpdateFlagStatus(ctx context.Context, flagId int64, status int32) error {
	query := "UPDATE `biz_flag` SET `status` = ? WHERE `id` = ?"
	_, err := m.conn.ExecCtx(ctx, query, status, flagId)
	return err
}
