package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type KeywordRule struct {
	Id            int64     `db:"id"`
	Pattern       string    `db:"pattern"`
	MatchType     string    `db:"match_type"`
	Category      string    `db:"category"`
	Enabled       int32     `db:"enabled"`
	CreatorUserId int64     `db:"creator_user_id"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

type KeywordRuleModel struct {
	conn sqlx.SqlConn
}

func NewKeywordRuleModel(conn sqlx.SqlConn) *KeywordRuleModel {
	return &KeywordRuleModel{conn: conn}
}

var keywordRuleSelectCols = "`id`,`pattern`,`match_type`,`category`,`enabled`,`creator_user_id`,`created_at`,`updated_at`"

func (m *KeywordRuleModel) Insert(ctx context.Context, rule *KeywordRule) (int64, error) {
	query := "INSERT INTO `keyword_rule` (`pattern`,`match_type`,`category`,`enabled`,`creator_user_id`) VALUES (?,?,?,?,?)"
	result, err := m.conn.ExecCtx(ctx, query, rule.Pattern, rule.MatchType, rule.Category, rule.Enabled, rule.CreatorUserId)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (m *KeywordRuleModel) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM `keyword_rule` WHERE `id` = ?"
	result, err := m.conn.ExecCtx(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (m *KeywordRuleModel) FindById(ctx context.Context, id int64) (*KeywordRule, error) {
	query := fmt.Sprintf("SELECT %s FROM `keyword_rule` WHERE `id` = ? LIMIT 1", keywordRuleSelectCols)
	var rule KeywordRule
	if err := m.conn.QueryRowCtx(ctx, &rule, query, id); err != nil {
		return nil, err
	}
	return &rule, nil
}

func (m *KeywordRuleModel) FindByIdPrimary(ctx context.Context, id int64) (*KeywordRule, error) {
	return m.FindById(sqlx.WithReadPrimary(ctx), id)
}

func (m *KeywordRuleModel) List(ctx context.Context, enabledOnly bool) ([]*KeywordRule, error) {
	query := fmt.Sprintf("SELECT %s FROM `keyword_rule`", keywordRuleSelectCols)
	args := []interface{}{}
	if enabledOnly {
		query += " WHERE `enabled` = 1"
	}
	query += " ORDER BY `id` DESC"

	var items []*KeywordRule
	if err := m.conn.QueryRowsCtx(ctx, &items, query, args...); err != nil {
		return nil, err
	}
	if items == nil {
		items = []*KeywordRule{}
	}
	return items, nil
}

func (m *KeywordRuleModel) ListPrimary(ctx context.Context, enabledOnly bool) ([]*KeywordRule, error) {
	return m.List(sqlx.WithReadPrimary(ctx), enabledOnly)
}
