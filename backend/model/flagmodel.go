package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type Flag struct {
	Id                   int64     `db:"id"`
	TargetType           string    `db:"target_type"` // "paper" / "rating" / "user"
	TargetId             int64     `db:"target_id"`
	ReporterId           int64     `db:"reporter_id"`
	Reason               string    `db:"reason"` // "abuse" / "spam" / "plagiarism" / "sensitive" / "manipulation"
	Detail               string    `db:"detail"`
	ReporterContribution float64   `db:"reporter_contribution"`
	Status               int32     `db:"status"` // 0=pending, 1=degraded, 2=dismissed
	CreatedAt            time.Time `db:"created_at"`
}

type FlagModel struct {
	conn sqlx.SqlConn
}

func NewFlagModel(conn sqlx.SqlConn) *FlagModel {
	return &FlagModel{conn: conn}
}

func NewFlagModelFromDB(db *sql.DB) *FlagModel {
	return &FlagModel{conn: sqlx.NewSqlConnFromDB(db)}
}

// Insert creates a new flag record
func (m *FlagModel) Insert(ctx context.Context, f *Flag) (int64, error) {
	query := `INSERT INTO flag (target_type, target_id, reporter_id, reason, detail, reporter_contribution)
		VALUES (?, ?, ?, ?, ?, ?)`
	result, err := m.conn.ExecCtx(ctx, query, f.TargetType, f.TargetId, f.ReporterId,
		f.Reason, f.Detail, f.ReporterContribution)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// HasFlagged checks if a user has already flagged a target
func (m *FlagModel) HasFlagged(ctx context.Context, targetType string, targetId, reporterId int64) (bool, error) {
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count,
		"SELECT COUNT(*) FROM `flag` WHERE `target_type` = ? AND `target_id` = ? AND `reporter_id` = ?",
		targetType, targetId, reporterId)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// FlagStats holds aggregated flag information for a target
type FlagStats struct {
	TotalCount  int     `db:"total_count"`
	WeightedSum float64 `db:"weighted_sum"`
}

// CountByTarget returns the total flag count and weighted sum for a target
func (m *FlagModel) CountByTarget(ctx context.Context, targetType string, targetId int64) (*FlagStats, error) {
	var stats FlagStats
	query := `SELECT COUNT(*) as total_count, IFNULL(SUM(reporter_contribution), 0) as weighted_sum
		FROM flag WHERE target_type = ? AND target_id = ? AND status = 0`
	err := m.conn.QueryRowCtx(ctx, &stats, query, targetType, targetId)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// ListPending returns pending flags with pagination
func (m *FlagModel) ListPending(ctx context.Context, page, pageSize int) ([]*Flag, int64, error) {
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, "SELECT COUNT(*) FROM `flag` WHERE `status` = 0")
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := `SELECT id, target_type, target_id, reporter_id, reason,
		IFNULL(detail, '') as detail, reporter_contribution, status, created_at
		FROM flag WHERE status = 0 ORDER BY created_at DESC LIMIT ? OFFSET ?`
	var flags []*Flag
	err = m.conn.QueryRowsCtx(ctx, &flags, query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	return flags, total, nil
}

// ListByTarget returns all flags for a specific target
func (m *FlagModel) ListByTarget(ctx context.Context, targetType string, targetId int64) ([]*Flag, error) {
	query := `SELECT id, target_type, target_id, reporter_id, reason,
		IFNULL(detail, '') as detail, reporter_contribution, status, created_at
		FROM flag WHERE target_type = ? AND target_id = ? ORDER BY created_at DESC`
	var flags []*Flag
	err := m.conn.QueryRowsCtx(ctx, &flags, query, targetType, targetId)
	if err != nil {
		return nil, err
	}
	return flags, nil
}

// UpdateStatus resolves a flag
func (m *FlagModel) UpdateStatus(ctx context.Context, id int64, status int32) error {
	query := "UPDATE `flag` SET `status` = ? WHERE `id` = ?"
	_, err := m.conn.ExecCtx(ctx, query, status, id)
	return err
}

// ResolveByTarget resolves all pending flags for a target
func (m *FlagModel) ResolveByTarget(ctx context.Context, targetType string, targetId int64, status int32) error {
	query := "UPDATE `flag` SET `status` = ? WHERE `target_type` = ? AND `target_id` = ? AND `status` = 0"
	_, err := m.conn.ExecCtx(ctx, query, status, targetType, targetId)
	return err
}
