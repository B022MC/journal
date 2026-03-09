package model

import (
	"context"
	"fmt"
	"strings"
)

// ListPapersPaginatedAdmin returns a paginated list of papers for admin
func (m *PaperModel) ListPapersPaginatedAdmin(ctx context.Context, page, pageSize int, zone string, status int32) ([]*Paper, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var conditions []string
	var args []interface{}

	if zone != "" {
		conditions = append(conditions, "zone = ?")
		args = append(args, zone)
	}
	// status: negative means no filter, otherwise exact match
	if status >= 0 {
		conditions = append(conditions, "status = ?")
		args = append(args, status)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM `paper` %s", whereClause)
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf("SELECT %s FROM `paper` %s ORDER BY id DESC LIMIT ? OFFSET ?", paperSelectCols, whereClause)
	args = append(args, pageSize, offset)

	var papers []*Paper
	err = m.conn.QueryRowsCtx(ctx, &papers, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return papers, total, nil
}

// UpdatePaperStatusAdmin updates paper status
func (m *PaperModel) UpdatePaperStatusAdmin(ctx context.Context, paperId int64, status int32) error {
	query := "UPDATE `paper` SET `status` = ? WHERE `id` = ?"
	_, err := m.conn.ExecCtx(ctx, query, status, paperId)
	return err
}

// UpdatePaperZoneAdmin updates paper zone
func (m *PaperModel) UpdatePaperZoneAdmin(ctx context.Context, paperId int64, zone string) error {
	query := "UPDATE `paper` SET `zone` = ? WHERE `id` = ?"
	_, err := m.conn.ExecCtx(ctx, query, zone, paperId)
	return err
}
