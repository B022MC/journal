package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type News struct {
	Id        int64          `db:"id"`
	Title     string         `db:"title"`
	TitleEn   string         `db:"title_en"`
	Content   string         `db:"content"`
	ContentEn sql.NullString `db:"content_en"`
	AuthorId  int64          `db:"author_id"`
	Category  string         `db:"category"`
	IsPinned  int32          `db:"is_pinned"`
	Status    int32          `db:"status"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
}

func (n *News) GetContentEn() string {
	if n.ContentEn.Valid {
		return n.ContentEn.String
	}
	return ""
}

func (n *News) GetIsPinned() bool {
	return n.IsPinned == 1
}

type NewsModel struct {
	conn sqlx.SqlConn
}

func NewNewsModel(conn sqlx.SqlConn) *NewsModel {
	return &NewsModel{conn: conn}
}

func (m *NewsModel) Insert(ctx context.Context, n *News) (int64, error) {
	query := "INSERT INTO `biz_news` (`title`,`title_en`,`content`,`content_en`,`author_id`,`category`,`is_pinned`,`status`) VALUES (?,?,?,?,?,?,?,?)"
	result, err := m.conn.ExecCtx(ctx, query,
		n.Title, n.TitleEn, n.Content, n.ContentEn, n.AuthorId, n.Category, n.IsPinned, n.Status,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (m *NewsModel) FindById(ctx context.Context, id int64) (*News, error) {
	query := "SELECT `id`,`title`,`title_en`,`content`,`content_en`,`author_id`,`category`,`is_pinned`,`status`,`created_at`,`updated_at` FROM `biz_news` WHERE `id` = ? AND `status` > 0 LIMIT 1"
	var n News
	err := m.conn.QueryRowCtx(ctx, &n, query, id)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (m *NewsModel) List(ctx context.Context, category string, page, pageSize int) ([]*News, int64, error) {
	where := "`status` = 1"
	args := []interface{}{}

	if category != "" {
		where += " AND `category` = ?"
		args = append(args, category)
	}

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM `biz_news` WHERE %s", where)
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf("SELECT `id`,`title`,`title_en`,`content`,`content_en`,`author_id`,`category`,`is_pinned`,`status`,`created_at`,`updated_at` FROM `biz_news` WHERE %s ORDER BY `is_pinned` DESC, `created_at` DESC LIMIT ? OFFSET ?", where)
	args = append(args, pageSize, offset)

	var items []*News
	err = m.conn.QueryRowsCtx(ctx, &items, query, args...)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
