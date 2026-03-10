package model

import (
	"context"
	"database/sql"
	"math"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type Rating struct {
	Id                int64     `db:"id"`
	PaperId           int64     `db:"paper_id"`
	UserId            int64     `db:"user_id"`
	Score             int32     `db:"score"`
	Comment           string    `db:"comment"`
	ReviewerWeight    float64   `db:"reviewer_weight"`
	SourceIP          string    `db:"source_ip"`
	UserAgent         string    `db:"user_agent"`
	DeviceFingerprint string    `db:"device_fingerprint"`
	CreatedAt         time.Time `db:"created_at"`
	UpdatedAt         time.Time `db:"updated_at"`
}

type RatingModel struct {
	conn sqlx.SqlConn
}

func NewRatingModel(conn sqlx.SqlConn) *RatingModel {
	return &RatingModel{conn: conn}
}

func (m *RatingModel) Upsert(ctx context.Context, r *Rating) (int64, error) {
	query := "INSERT INTO `rating` (`paper_id`,`user_id`,`score`,`comment`,`reviewer_weight`) VALUES (?,?,?,?,?) ON DUPLICATE KEY UPDATE `score` = VALUES(`score`), `comment` = VALUES(`comment`), `reviewer_weight` = VALUES(`reviewer_weight`)"
	result, err := m.conn.ExecCtx(ctx, query, r.PaperId, r.UserId, r.Score, r.Comment, r.ReviewerWeight)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (m *RatingModel) FindById(ctx context.Context, id int64) (*Rating, error) {
	query := "SELECT `id`,`paper_id`,`user_id`,`score`,IFNULL(`comment`,'') as `comment`,`reviewer_weight`,IFNULL(`source_ip`,'') as `source_ip`,IFNULL(`user_agent`,'') as `user_agent`,IFNULL(`device_fingerprint`,'') as `device_fingerprint`,`created_at`,`updated_at` FROM `rating` WHERE `id` = ? LIMIT 1"
	var item Rating
	err := m.conn.QueryRowCtx(ctx, &item, query, id)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *RatingModel) FindByIdPrimary(ctx context.Context, id int64) (*Rating, error) {
	return m.FindById(sqlx.WithReadPrimary(ctx), id)
}

type RatingWithUser struct {
	Id        int64     `db:"id"`
	PaperId   int64     `db:"paper_id"`
	UserId    int64     `db:"user_id"`
	Score     int32     `db:"score"`
	Comment   string    `db:"comment"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	Username  string    `db:"username"`
	Nickname  string    `db:"nickname"`
}

func (m *RatingModel) ListByPaper(ctx context.Context, paperId int64, page, pageSize int) ([]*RatingWithUser, int64, error) {
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, "SELECT COUNT(*) FROM `rating` WHERE `paper_id` = ?", paperId)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := "SELECT r.`id`, r.`paper_id`, r.`user_id`, r.`score`, IFNULL(r.`comment`,'') as `comment`, r.`created_at`, r.`updated_at`, u.`username`, u.`nickname` FROM `rating` r JOIN `user` u ON r.`user_id` = u.`id` WHERE r.`paper_id` = ? ORDER BY r.`created_at` DESC LIMIT ? OFFSET ?"

	var items []*RatingWithUser
	err = m.conn.QueryRowsCtx(ctx, &items, query, paperId, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (m *RatingModel) ListByUser(ctx context.Context, userId int64, page, pageSize int) ([]*RatingWithUser, int64, error) {
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, "SELECT COUNT(*) FROM `rating` WHERE `user_id` = ?", userId)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := "SELECT r.`id`, r.`paper_id`, r.`user_id`, r.`score`, IFNULL(r.`comment`,'') as `comment`, r.`created_at`, r.`updated_at`, u.`username`, u.`nickname` FROM `rating` r JOIN `user` u ON r.`user_id` = u.`id` WHERE r.`user_id` = ? ORDER BY r.`created_at` DESC LIMIT ? OFFSET ?"

	var items []*RatingWithUser
	err = m.conn.QueryRowsCtx(ctx, &items, query, userId, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

type RatingStats struct {
	AvgScore float64 `db:"avg_score"`
	Count    int32   `db:"cnt"`
	Stddev   float64 `db:"stddev_val"`
}

func (m *RatingModel) GetPaperRatingStats(ctx context.Context, paperId int64) (avgScore float64, count int32, stddev float64, err error) {
	var stats RatingStats
	query := "SELECT IFNULL(AVG(`score`),0) as `avg_score`, COUNT(*) as `cnt`, IFNULL(STDDEV_POP(`score`),0) as `stddev_val` FROM `rating` WHERE `paper_id` = ?"
	err = m.conn.QueryRowCtx(ctx, &stats, query, paperId)
	if err != nil {
		return
	}
	avgScore = stats.AvgScore
	count = stats.Count
	stddev = math.Min(stats.Stddev/5.0, 1.0)
	return
}

func (m *RatingModel) HasRated(ctx context.Context, paperId, userId int64) (bool, error) {
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, "SELECT COUNT(*) FROM `rating` WHERE `paper_id` = ? AND `user_id` = ?", paperId, userId)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func NewRatingModelFromDB(db *sql.DB) *RatingModel {
	return &RatingModel{conn: sqlx.NewSqlConnFromDB(db)}
}

// === 治理模块方法 ===

// ListAllByUser returns all ratings by a user (no pagination, for contribution calc)
func (m *RatingModel) ListAllByUser(ctx context.Context, userId int64) ([]*Rating, error) {
	query := "SELECT `id`,`paper_id`,`user_id`,`score`,IFNULL(`comment`,'') as `comment`,`reviewer_weight`,IFNULL(`source_ip`,'') as `source_ip`,IFNULL(`user_agent`,'') as `user_agent`,IFNULL(`device_fingerprint`,'') as `device_fingerprint`,`created_at`,`updated_at` FROM `rating` WHERE `user_id` = ? ORDER BY `created_at` DESC"
	var items []*Rating
	err := m.conn.QueryRowsCtx(ctx, &items, query, userId)
	if err != nil {
		return nil, err
	}
	return items, nil
}

// CountConsecutiveSameScore counts the maximum consecutive same-score ratings for a user
func (m *RatingModel) CountConsecutiveSameScore(ctx context.Context, userId int64) (int, error) {
	query := `SELECT score FROM rating WHERE user_id = ? ORDER BY created_at DESC LIMIT 20`
	var scores []int32
	err := m.conn.QueryRowsCtx(ctx, &scores, query, userId)
	if err != nil {
		return 0, err
	}
	if len(scores) == 0 {
		return 0, nil
	}
	maxConsecutive := 1
	current := 1
	for i := 1; i < len(scores); i++ {
		if scores[i] == scores[i-1] {
			current++
			if current > maxConsecutive {
				maxConsecutive = current
			}
		} else {
			current = 1
		}
	}
	return maxConsecutive, nil
}

// CountByUserSince returns the number of ratings submitted by a user since the given time.
func (m *RatingModel) CountByUserSince(ctx context.Context, userId int64, since time.Time) (int32, error) {
	var count int32
	query := "SELECT COUNT(*) FROM `rating` WHERE `user_id` = ? AND `created_at` >= ?"
	err := m.conn.QueryRowCtx(sqlx.WithReadPrimary(ctx), &count, query, userId, since)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m *RatingModel) CountByUser(ctx context.Context, userId int64) (int64, error) {
	var count int64
	query := "SELECT COUNT(*) FROM `rating` WHERE `user_id` = ?"
	err := m.conn.QueryRowCtx(sqlx.WithReadPrimary(ctx), &count, query, userId)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m *RatingModel) UpdateRequestFingerprint(ctx context.Context, paperId, userId int64, sourceIP, userAgent, deviceFingerprint string) error {
	query := "UPDATE `rating` SET `source_ip` = ?, `user_agent` = ?, `device_fingerprint` = ? WHERE `paper_id` = ? AND `user_id` = ?"
	_, err := m.conn.ExecCtx(ctx, query, sourceIP, userAgent, deviceFingerprint, paperId, userId)
	return err
}

func (m *RatingModel) CountDistinctUsersByPaperIPSince(ctx context.Context, paperId int64, sourceIP string, since time.Time) (int32, error) {
	var count int32
	query := "SELECT COUNT(DISTINCT `user_id`) FROM `rating` WHERE `paper_id` = ? AND `source_ip` = ? AND `updated_at` >= ?"
	err := m.conn.QueryRowCtx(sqlx.WithReadPrimary(ctx), &count, query, paperId, sourceIP, since)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m *RatingModel) CountDistinctUsersByPaperFingerprintSince(ctx context.Context, paperId int64, deviceFingerprint string, since time.Time) (int32, error) {
	var count int32
	query := "SELECT COUNT(DISTINCT `user_id`) FROM `rating` WHERE `paper_id` = ? AND `device_fingerprint` = ? AND `updated_at` >= ?"
	err := m.conn.QueryRowCtx(sqlx.WithReadPrimary(ctx), &count, query, paperId, deviceFingerprint, since)
	if err != nil {
		return 0, err
	}
	return count, nil
}

type ScoreHistogram struct {
	Total   int32
	Buckets [10]int32
}

type scoreHistogramRow struct {
	Score int32 `db:"score"`
	Count int32 `db:"cnt"`
}

func (m *RatingModel) GetPaperScoreHistogram(ctx context.Context, paperId int64) (*ScoreHistogram, error) {
	rows := make([]*scoreHistogramRow, 0)
	query := "SELECT `score`, COUNT(*) as `cnt` FROM `rating` WHERE `paper_id` = ? GROUP BY `score`"
	if err := m.conn.QueryRowsCtx(sqlx.WithReadPrimary(ctx), &rows, query, paperId); err != nil {
		return nil, err
	}

	histogram := &ScoreHistogram{}
	for _, row := range rows {
		if row == nil || row.Score < 1 || row.Score > 10 {
			continue
		}
		histogram.Buckets[row.Score-1] = row.Count
		histogram.Total += row.Count
	}
	return histogram, nil
}

// WeightedRatingStats holds the result of a weighted rating aggregation
type WeightedRatingStats struct {
	WeightedAvg     float64 `db:"weighted_avg"`
	Count           int32   `db:"cnt"`
	Stddev          float64 `db:"stddev_val"`
	AvgReviewerAuth float64 `db:"avg_reviewer_auth"`
}

// GetWeightedRatingStats computes weighted average using reviewer_weight
func (m *RatingModel) GetWeightedRatingStats(ctx context.Context, paperId int64) (*WeightedRatingStats, error) {
	var stats WeightedRatingStats
	query := `SELECT
		IFNULL(SUM(score * reviewer_weight) / NULLIF(SUM(reviewer_weight), 0), 0) as weighted_avg,
		COUNT(*) as cnt,
		IFNULL(STDDEV_POP(score), 0) as stddev_val,
		IFNULL(AVG(reviewer_weight), 0) as avg_reviewer_auth
		FROM rating WHERE paper_id = ?`
	err := m.conn.QueryRowCtx(ctx, &stats, query, paperId)
	if err != nil {
		return nil, err
	}
	stats.Stddev = math.Min(stats.Stddev/5.0, 1.0)
	return &stats, nil
}
