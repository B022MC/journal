package model

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type Paper struct {
	Id                int64          `db:"id"`
	Title             string         `db:"title"`
	TitleEn           string         `db:"title_en"`
	Abstract          string         `db:"abstract"`
	AbstractEn        sql.NullString `db:"abstract_en"`
	Content           string         `db:"content"`
	AuthorId          int64          `db:"author_id"`
	AuthorName        string         `db:"author_name"`
	Discipline        string         `db:"discipline"`
	Zone              string         `db:"zone"`
	ShitScore         float64        `db:"shit_score"`
	AvgRating         float64        `db:"avg_rating"`
	RatingCount       int32          `db:"rating_count"`
	ViewCount         int32          `db:"view_count"`
	ControversyIndex  float64        `db:"controversy_index"`
	WeightedAvgRating float64        `db:"weighted_avg_rating"`
	ReviewerAuthority float64        `db:"reviewer_authority"`
	FlagCount         int32          `db:"flag_count"`
	DegradationLevel  int32          `db:"degradation_level"`
	FilePath          string         `db:"file_path"`
	Doi               string         `db:"doi"`
	Keywords          string         `db:"keywords"`
	Simhash           uint64         `db:"simhash"`
	Status            int32          `db:"status"`
	PromotedAt        sql.NullTime   `db:"promoted_at"`
	LastAccessedAt    sql.NullTime   `db:"last_accessed_at"`
	CreatedAt         time.Time      `db:"created_at"`
	UpdatedAt         time.Time      `db:"updated_at"`
}

func (p *Paper) GetAbstractEn() string {
	if p.AbstractEn.Valid {
		return p.AbstractEn.String
	}
	return ""
}

func (p *Paper) GetPromotedAt() *time.Time {
	if p.PromotedAt.Valid {
		return &p.PromotedAt.Time
	}
	return nil
}

type PaperModel struct {
	conn sqlx.SqlConn
}

func NewPaperModel(conn sqlx.SqlConn) *PaperModel {
	return &PaperModel{conn: conn}
}

var paperSelectCols = "`id`,`title`,`title_en`,`abstract`,`abstract_en`,`content`,`author_id`,`author_name`,`discipline`,`zone`," +
	"`shit_score`,`avg_rating`,`rating_count`,`view_count`,`controversy_index`," +
	"`weighted_avg_rating`,`reviewer_authority`,`flag_count`,`degradation_level`," +
	"`file_path`,`doi`,`keywords`,`simhash`,`status`,`promoted_at`,`last_accessed_at`,`created_at`,`updated_at`"

// === 写操作 → 主库 ===

func (m *PaperModel) Insert(ctx context.Context, p *Paper) (int64, error) {
	query := "INSERT INTO `paper` (`title`,`title_en`,`abstract`,`abstract_en`,`content`,`author_id`,`author_name`,`discipline`,`zone`,`keywords`,`file_path`,`doi`,`simhash`,`status`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
	result, err := m.conn.ExecCtx(ctx, query,
		p.Title, p.TitleEn, p.Abstract, p.AbstractEn, p.Content,
		p.AuthorId, p.AuthorName, p.Discipline, p.Zone,
		p.Keywords, p.FilePath, p.Doi, p.Simhash, p.Status,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

type PaperSimilarity struct {
	Id       int64  `db:"id"`
	Title    string `db:"title"`
	Distance int    `db:"distance"`
}

func (m *PaperModel) FindSimilarBySimhash(ctx context.Context, simhash uint64, maxDistance, limit int) ([]*PaperSimilarity, error) {
	if simhash == 0 || maxDistance < 0 || limit <= 0 {
		return []*PaperSimilarity{}, nil
	}

	query := `SELECT id, title, BIT_COUNT(simhash ^ ?) AS distance
		FROM paper
		WHERE simhash <> 0 AND status > 0 AND BIT_COUNT(simhash ^ ?) <= ?
		ORDER BY distance ASC, id DESC
		LIMIT ?`

	var items []*PaperSimilarity
	err := m.conn.QueryRowsCtx(sqlx.WithReadPrimary(ctx), &items, query, simhash, simhash, maxDistance, limit)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []*PaperSimilarity{}
	}
	return items, nil
}

func (m *PaperModel) UpdateZone(ctx context.Context, id int64, zone string) error {
	query := "UPDATE `paper` SET `zone` = ?, `promoted_at` = ? WHERE `id` = ?"
	_, err := m.conn.ExecCtx(ctx, query, zone, time.Now(), id)
	return err
}

func (m *PaperModel) IncrViewCount(ctx context.Context, id int64) error {
	query := "UPDATE `paper` SET `view_count` = `view_count` + 1 WHERE `id` = ?"
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

func (m *PaperModel) UpdateDoi(ctx context.Context, id int64, doi string) error {
	query := "UPDATE `paper` SET `doi` = ? WHERE `id` = ?"
	_, err := m.conn.ExecCtx(ctx, query, doi, id)
	return err
}

func (m *PaperModel) UpdateScores(ctx context.Context, id int64, avgRating float64, ratingCount int32, controversyIndex float64) error {
	shitScore := CalcShitScore(avgRating, ratingCount, 0, controversyIndex)
	query := "UPDATE `paper` SET `avg_rating` = ?, `rating_count` = ?, `controversy_index` = ?, `shit_score` = ? WHERE `id` = ?"
	_, err := m.conn.ExecCtx(ctx, query, avgRating, ratingCount, controversyIndex, shitScore, id)
	return err
}

// UpdateScoresV2 updates paper scores with v2 algorithm including reviewer authority and freshness
func (m *PaperModel) UpdateScoresV2(ctx context.Context, id int64, avgRating float64, ratingCount int32,
	viewCount int32, controversyIndex float64, weightedAvg float64, reviewerAuthority float64, createdAt time.Time) error {
	shitScore := CalcShitScoreV2(weightedAvg, ratingCount, viewCount, controversyIndex, reviewerAuthority, createdAt)
	query := `UPDATE paper SET avg_rating = ?, rating_count = ?, controversy_index = ?,
		weighted_avg_rating = ?, reviewer_authority = ?, shit_score = ? WHERE id = ?`
	_, err := m.conn.ExecCtx(ctx, query, avgRating, ratingCount, controversyIndex,
		weightedAvg, reviewerAuthority, shitScore, id)
	return err
}

// UpdateDegradationLevel sets the degradation level for a paper
func (m *PaperModel) UpdateDegradationLevel(ctx context.Context, id int64, level int32) error {
	query := "UPDATE `paper` SET `degradation_level` = ? WHERE `id` = ?"
	_, err := m.conn.ExecCtx(ctx, query, level, id)
	return err
}

// IncrFlagCount atomically increments the flag count
func (m *PaperModel) IncrFlagCount(ctx context.Context, id int64) error {
	query := "UPDATE `paper` SET `flag_count` = `flag_count` + 1 WHERE `id` = ?"
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

// === 读操作 → 从库 (SELECT 自动路由) ===

func (m *PaperModel) FindById(ctx context.Context, id int64) (*Paper, error) {
	query := fmt.Sprintf("SELECT %s FROM `paper` WHERE `id` = ? AND `status` > 0 LIMIT 1", paperSelectCols)
	var p Paper
	err := m.conn.QueryRowCtx(ctx, &p, query, id)
	if err != nil {
		return nil, err
	}
	// 异步更新 last_accessed_at，不影响读取响应时间
	go m.updateLastAccessedAt(context.Background(), id)
	return &p, nil
}

// FindByIdPrimary 提交后立即读取详情走主库
func (m *PaperModel) FindByIdPrimary(ctx context.Context, id int64) (*Paper, error) {
	return m.FindById(sqlx.WithReadPrimary(ctx), id)
}

func (m *PaperModel) List(ctx context.Context, zone, discipline, sort string, page, pageSize int) ([]*Paper, int64, error) {
	where := []string{"`status` > 0"}
	args := []interface{}{}

	if zone != "" {
		where = append(where, "`zone` = ?")
		args = append(args, zone)
	}
	if discipline != "" {
		where = append(where, "`discipline` = ?")
		args = append(args, discipline)
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM `paper` WHERE %s", whereClause)
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	orderBy := "`created_at` DESC"
	switch sort {
	case "highest_rated":
		orderBy = "`shit_score` DESC"
	case "most_rated":
		orderBy = "`rating_count` DESC"
	}

	offset := (page - 1) * pageSize
	selectQuery := fmt.Sprintf("SELECT %s FROM `paper` WHERE %s ORDER BY %s LIMIT ? OFFSET ?",
		paperSelectCols, whereClause, orderBy)
	args = append(args, pageSize, offset)

	var papers []*Paper
	err = m.conn.QueryRowsCtx(ctx, &papers, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	return papers, total, nil
}

func (m *PaperModel) ListByAuthor(ctx context.Context, authorId int64, page, pageSize int) ([]*Paper, int64, error) {
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, "SELECT COUNT(*) FROM `paper` WHERE `author_id` = ? AND `status` > 0", authorId)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf("SELECT %s FROM `paper` WHERE `author_id` = ? AND `status` > 0 ORDER BY `created_at` DESC LIMIT ? OFFSET ?", paperSelectCols)
	var papers []*Paper
	err = m.conn.QueryRowsCtx(ctx, &papers, query, authorId, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	return papers, total, nil
}

func (m *PaperModel) CountByAuthor(ctx context.Context, authorId int64) (int64, error) {
	var total int64
	query := "SELECT COUNT(*) FROM `paper` WHERE `author_id` = ? AND `status` > 0"
	err := m.conn.QueryRowCtx(sqlx.WithReadPrimary(ctx), &total, query, authorId)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (m *PaperModel) CountByAuthorZone(ctx context.Context, authorId int64, zone string) (int64, error) {
	var total int64
	query := "SELECT COUNT(*) FROM `paper` WHERE `author_id` = ? AND `zone` = ? AND `status` > 0"
	err := m.conn.QueryRowCtx(sqlx.WithReadPrimary(ctx), &total, query, authorId, zone)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (m *PaperModel) Search(ctx context.Context, query, discipline string, page, pageSize int) ([]*Paper, int64, error) {
	where := []string{"`status` > 0", "MATCH(`title`,`abstract`,`keywords`) AGAINST(? IN BOOLEAN MODE)"}
	args := []interface{}{query}

	if discipline != "" {
		where = append(where, "`discipline` = ?")
		args = append(args, discipline)
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQ := fmt.Sprintf("SELECT COUNT(*) FROM `paper` WHERE %s", whereClause)
	err := m.conn.QueryRowCtx(ctx, &total, countQ, args...)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	selectQ := fmt.Sprintf("SELECT %s FROM `paper` WHERE %s ORDER BY MATCH(`title`,`abstract`,`keywords`) AGAINST(? IN BOOLEAN MODE) DESC LIMIT ? OFFSET ?",
		paperSelectCols, whereClause)
	args = append(args, query, pageSize, offset)

	var papers []*Paper
	err = m.conn.QueryRowsCtx(ctx, &papers, selectQ, args...)
	if err != nil {
		return nil, 0, err
	}
	return papers, total, nil
}

// === 生命周期查询 → 从库 ===

func (m *PaperModel) GetPapersForPromotion(ctx context.Context, zone string, minRatingCount int, minShitScore float64) ([]*Paper, error) {
	query := fmt.Sprintf("SELECT %s FROM `paper` WHERE `zone` = ? AND `rating_count` >= ? AND `shit_score` >= ? AND `status` = 1", paperSelectCols)
	var papers []*Paper
	err := m.conn.QueryRowsCtx(ctx, &papers, query, zone, minRatingCount, minShitScore)
	if err != nil {
		return nil, err
	}
	return papers, nil
}

func (m *PaperModel) GetStalePapers(ctx context.Context, zone string, maxAge time.Duration, belowScore float64) ([]*Paper, error) {
	deadline := time.Now().Add(-maxAge)
	query := fmt.Sprintf("SELECT %s FROM `paper` WHERE `zone` = ? AND `created_at` < ? AND `shit_score` < ? AND `status` = 1", paperSelectCols)
	var papers []*Paper
	err := m.conn.QueryRowsCtx(ctx, &papers, query, zone, deadline, belowScore)
	if err != nil {
		return nil, err
	}
	return papers, nil
}

// GetPapersForPromotionV2 uses multi-dimensional thresholds for zone promotion
func (m *PaperModel) GetPapersForPromotionV2(ctx context.Context, zone string, minRatingCount int,
	minShitScore float64, minWeightedCount int, minReviewerAuth float64, minAgeDays int) ([]*Paper, error) {
	query := fmt.Sprintf(`SELECT %s FROM paper
		WHERE zone = ? AND rating_count >= ? AND shit_score >= ?
		AND rating_count >= ? AND reviewer_authority >= ?
		AND DATEDIFF(NOW(), created_at) >= ?
		AND status = 1 AND degradation_level = 0`, paperSelectCols)
	var papers []*Paper
	err := m.conn.QueryRowsCtx(ctx, &papers, query, zone, minRatingCount, minShitScore,
		minWeightedCount, minReviewerAuth, minAgeDays)
	if err != nil {
		return nil, err
	}
	return papers, nil
}

// GetStalePapersV2 uses enhanced demotion criteria including flag count
func (m *PaperModel) GetStalePapersV2(ctx context.Context, zone string, maxStaleAge time.Duration,
	belowScore float64, maxFlags int) ([]*Paper, error) {
	deadline := time.Now().Add(-maxStaleAge)
	query := fmt.Sprintf(`SELECT %s FROM paper
		WHERE zone = ? AND status = 1
		AND ((updated_at < ? AND shit_score < ?) OR flag_count >= ?)`, paperSelectCols)
	var papers []*Paper
	err := m.conn.QueryRowsCtx(ctx, &papers, query, zone, deadline, belowScore, maxFlags)
	if err != nil {
		return nil, err
	}
	return papers, nil
}

// CalcShitScore: shit_score = w1*norm_avg + w2*log(count+1) + w3*log(views+1) - w4*controversy
func CalcShitScore(avgRating float64, ratingCount int32, viewCount int32, controversyIndex float64) float64 {
	w1, w2, w3, w4 := 0.40, 0.25, 0.15, 0.20
	normAvg := avgRating / 10.0
	score := w1*normAvg +
		w2*math.Log(float64(ratingCount)+1) +
		w3*math.Log(float64(viewCount)+1) -
		w4*controversyIndex
	return math.Round(score*10000) / 10000
}

// CalcShitScoreV2 enhanced score with reviewer authority and freshness
// v2 = 0.30*weighted_avg + 0.20*log(weighted_count+1) + 0.10*log(views+1)
//   - 0.15*controversy + 0.15*reviewer_authority + 0.10*freshness
func CalcShitScoreV2(weightedAvg float64, ratingCount int32, viewCount int32,
	controversyIndex float64, reviewerAuthority float64, createdAt time.Time) float64 {
	w1, w2, w3, w4, w5, w6 := 0.30, 0.20, 0.10, 0.15, 0.15, 0.10
	normAvg := weightedAvg / 10.0
	daysSinceCreation := time.Since(createdAt).Hours() / 24.0
	freshness := 1.0 / (1.0 + daysSinceCreation/30.0)

	score := w1*normAvg +
		w2*math.Log(float64(ratingCount)+1) +
		w3*math.Log(float64(viewCount)+1) -
		w4*controversyIndex +
		w5*reviewerAuthority +
		w6*freshness
	return math.Round(score*10000) / 10000
}

func NewPaperModelFromDB(db *sql.DB) *PaperModel {
	return &PaperModel{conn: sqlx.NewSqlConnFromDB(db)}
}

// === 冷热数据分离 ===

// updateLastAccessedAt 异步更新最后访问时间（fire-and-forget）
func (m *PaperModel) updateLastAccessedAt(ctx context.Context, id int64) {
	query := "UPDATE `paper` SET `last_accessed_at` = NOW() WHERE `id` = ?"
	_, _ = m.conn.ExecCtx(ctx, query, id)
}

// GetColdPapers 查找符合冷数据标准的论文
// 冷数据定义：(>coldDays天未访问 且 zone=sediment) 或 status=0
func (m *PaperModel) GetColdPapers(ctx context.Context, coldDays int, batchSize int) ([]*Paper, error) {
	deadline := time.Now().AddDate(0, 0, -coldDays)
	query := fmt.Sprintf(`SELECT %s FROM paper
		WHERE (last_accessed_at IS NOT NULL AND last_accessed_at < ? AND zone = 'sediment' AND status = 1)
		   OR (status = 0)
		LIMIT ?`, paperSelectCols)
	var papers []*Paper
	err := m.conn.QueryRowsCtx(ctx, &papers, query, deadline, batchSize)
	if err != nil {
		return nil, err
	}
	return papers, nil
}

// ArchiveColdPaper 将单篇论文迁移到 cold_paper 并软删除源表记录
// === 写操作 → 主库 ===
func (m *PaperModel) ArchiveColdPaper(ctx context.Context, id int64) error {
	// Step 1: INSERT INTO cold_paper SELECT ... FROM paper
	archiveQuery := fmt.Sprintf(`INSERT INTO cold_paper
		(%s, archived_at)
		SELECT %s, NOW() FROM paper WHERE id = ?`,
		paperSelectCols, paperSelectCols)
	_, err := m.conn.ExecCtx(ctx, archiveQuery, id)
	if err != nil {
		return fmt.Errorf("archive paper %d: %w", id, err)
	}

	// Step 2: 源表软删除（status=0）
	deleteQuery := "UPDATE `paper` SET `status` = 0 WHERE `id` = ?"
	_, err = m.conn.ExecCtx(ctx, deleteQuery, id)
	if err != nil {
		return fmt.Errorf("soft-delete paper %d: %w", id, err)
	}

	return nil
}
