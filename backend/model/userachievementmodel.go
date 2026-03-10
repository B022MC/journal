package model

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type UserAchievement struct {
	Id         int64     `db:"id"`
	UserId     int64     `db:"user_id"`
	Code       string    `db:"code"`
	UnlockedAt time.Time `db:"unlocked_at"`
}

type UserAchievementModel struct {
	conn sqlx.SqlConn
}

func NewUserAchievementModel(conn sqlx.SqlConn) *UserAchievementModel {
	return &UserAchievementModel{conn: conn}
}

func (m *UserAchievementModel) InsertIgnore(ctx context.Context, userId int64, code string) error {
	query := "INSERT IGNORE INTO `user_achievement` (`user_id`, `code`) VALUES (?, ?)"
	_, err := m.conn.ExecCtx(ctx, query, userId, code)
	return err
}

func (m *UserAchievementModel) ListByUser(ctx context.Context, userId int64) ([]*UserAchievement, error) {
	query := "SELECT `id`,`user_id`,`code`,`unlocked_at` FROM `user_achievement` WHERE `user_id` = ? ORDER BY `unlocked_at` ASC, `id` ASC"
	var items []*UserAchievement
	err := m.conn.QueryRowsCtx(ctx, &items, query, userId)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (m *UserAchievementModel) ListByUserPrimary(ctx context.Context, userId int64) ([]*UserAchievement, error) {
	return m.ListByUser(sqlx.WithReadPrimary(ctx), userId)
}
