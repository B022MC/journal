package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type User struct {
	Id                int64        `db:"id"`
	Username          string       `db:"username"`
	Email             string       `db:"email"`
	PasswordHash      string       `db:"password_hash"`
	Nickname          string       `db:"nickname"`
	Avatar            string       `db:"avatar"`
	Role              int32        `db:"role"`
	ContributionScore float64      `db:"contribution_score"`
	LastActiveAt      sql.NullTime `db:"last_active_at"`
	ReviewCount30d    int32        `db:"review_count_30d"`
	Status            int32        `db:"status"`
	CreatedAt         time.Time    `db:"created_at"`
	UpdatedAt         time.Time    `db:"updated_at"`
}

type UserModel struct {
	conn sqlx.SqlConn
}

func NewUserModel(conn sqlx.SqlConn) *UserModel {
	return &UserModel{conn: conn}
}

// === 写操作 → 主库 ===

func (m *UserModel) Insert(ctx context.Context, u *User) (int64, error) {
	query := "INSERT INTO `biz_user` (`username`, `email`, `password_hash`, `nickname`, `avatar`, `role`, `status`) VALUES (?, ?, ?, ?, ?, ?, ?)"
	result, err := m.conn.ExecCtx(ctx, query,
		u.Username, u.Email, u.PasswordHash, u.Nickname, u.Avatar, u.Role, u.Status,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (m *UserModel) UpdateProfile(ctx context.Context, id int64, nickname, avatar string) error {
	query := "UPDATE `biz_user` SET `nickname` = ?, `avatar` = ? WHERE `id` = ?"
	result, err := m.conn.ExecCtx(ctx, query, nickname, avatar, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("user not found: %d", id)
	}
	return nil
}

// === 读操作 → 从库 ===

var userSelectCols = "`id`,`username`,`email`,`password_hash`,`nickname`,`avatar`,`role`,`contribution_score`,`last_active_at`,`review_count_30d`,`status`,`created_at`,`updated_at`"

func (m *UserModel) FindByUsername(ctx context.Context, username string) (*User, error) {
	query := fmt.Sprintf("SELECT %s FROM `biz_user` WHERE `username` = ? LIMIT 1", userSelectCols)
	var u User
	err := m.conn.QueryRowCtx(ctx, &u, query, username)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (m *UserModel) FindById(ctx context.Context, id int64) (*User, error) {
	query := fmt.Sprintf("SELECT %s FROM `biz_user` WHERE `id` = ? LIMIT 1", userSelectCols)
	var u User
	err := m.conn.QueryRowCtx(ctx, &u, query, id)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// FindByIdPrimary reads from primary to avoid replica lag when the freshest score is required.
func (m *UserModel) FindByIdPrimary(ctx context.Context, id int64) (*User, error) {
	return m.FindById(sqlx.WithReadPrimary(ctx), id)
}

func (m *UserModel) FindByEmail(ctx context.Context, email string) (*User, error) {
	query := fmt.Sprintf("SELECT %s FROM `biz_user` WHERE `email` = ? LIMIT 1", userSelectCols)
	var u User
	err := m.conn.QueryRowCtx(ctx, &u, query, email)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (m *UserModel) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	query := "SELECT COUNT(*) FROM `biz_user` WHERE `username` = ?"
	err := m.conn.QueryRowCtx(ctx, &count, query, username)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (m *UserModel) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	query := "SELECT COUNT(*) FROM `biz_user` WHERE `email` = ?"
	err := m.conn.QueryRowCtx(ctx, &count, query, email)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// === 强一致性读 → 写后立即读走主库 ===

// FindByUsernamePrimary 注册时检查唯一性需走主库（写后立即读）
func (m *UserModel) FindByUsernamePrimary(ctx context.Context, username string) (*User, error) {
	return m.FindByUsername(sqlx.WithReadPrimary(ctx), username)
}

// ExistsByUsernamePrimary 注册唯一性检查走主库
func (m *UserModel) ExistsByUsernamePrimary(ctx context.Context, username string) (bool, error) {
	return m.ExistsByUsername(sqlx.WithReadPrimary(ctx), username)
}

// ExistsByEmailPrimary 注册唯一性检查走主库
func (m *UserModel) ExistsByEmailPrimary(ctx context.Context, email string) (bool, error) {
	return m.ExistsByEmail(sqlx.WithReadPrimary(ctx), email)
}

func NewUserModelFromDB(db *sql.DB) *UserModel {
	return &UserModel{conn: sqlx.NewSqlConnFromDB(db)}
}

// === 治理模块方法 ===

// UpdateContributionScore sets the contribution score for a user
func (m *UserModel) UpdateContributionScore(ctx context.Context, userId int64, score float64) error {
	query := "UPDATE `biz_user` SET `contribution_score` = ? WHERE `id` = ?"
	_, err := m.conn.ExecCtx(ctx, query, score, userId)
	return err
}

// AutoAssignRole updates the user role based on their contribution score
func (m *UserModel) AutoAssignRole(ctx context.Context, userId int64, newRole int32) error {
	query := "UPDATE `biz_user` SET `role` = ? WHERE `id` = ?"
	_, err := m.conn.ExecCtx(ctx, query, newRole, userId)
	return err
}

// BatchDecayContribution decays contribution_score for users inactive > inactiveDays
func (m *UserModel) BatchDecayContribution(ctx context.Context, inactiveDays int, decayRate float64) (int64, error) {
	query := `UPDATE biz_user SET contribution_score = GREATEST(0, contribution_score - contribution_score * ?)
		WHERE last_active_at < DATE_SUB(NOW(), INTERVAL ? DAY)
		AND contribution_score > 0 AND status = 1`
	result, err := m.conn.ExecCtx(ctx, query, decayRate, inactiveDays)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// GetInactiveDays returns the number of days since a user was last active
func (m *UserModel) GetInactiveDays(ctx context.Context, userId int64) (int, error) {
	var days int
	query := `SELECT IFNULL(DATEDIFF(NOW(), last_active_at), DATEDIFF(NOW(), created_at))
		FROM biz_user WHERE id = ?`
	err := m.conn.QueryRowCtx(ctx, &days, query, userId)
	if err != nil {
		return 0, err
	}
	return days, nil
}

// UpdateLastActive sets last_active_at to now
func (m *UserModel) UpdateLastActive(ctx context.Context, userId int64) error {
	query := "UPDATE `biz_user` SET `last_active_at` = NOW() WHERE `id` = ?"
	_, err := m.conn.ExecCtx(ctx, query, userId)
	return err
}

// IncrReviewCount30d increments the 30-day review counter
func (m *UserModel) IncrReviewCount30d(ctx context.Context, userId int64) error {
	query := "UPDATE `biz_user` SET `review_count_30d` = `review_count_30d` + 1 WHERE `id` = ?"
	_, err := m.conn.ExecCtx(ctx, query, userId)
	return err
}

// GetActiveUsers returns users with contribution_score > 0 and recent activity
func (m *UserModel) GetActiveUsers(ctx context.Context, minScore float64) ([]*User, error) {
	query := fmt.Sprintf("SELECT %s FROM `biz_user` WHERE `contribution_score` >= ? AND `status` = 1", userSelectCols)
	var users []*User
	err := m.conn.QueryRowsCtx(ctx, &users, query, minScore)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// GetAllActiveUsers returns all active users for batch processing
func (m *UserModel) GetAllActiveUsers(ctx context.Context) ([]*User, error) {
	query := fmt.Sprintf("SELECT %s FROM `biz_user` WHERE `status` = 1", userSelectCols)
	var users []*User
	err := m.conn.QueryRowsCtx(ctx, &users, query)
	if err != nil {
		return nil, err
	}
	return users, nil
}
