package degradation

import (
	"context"
	"log"
	"math"

	"journal/model"
)

// Degradation levels
const (
	LevelNormal    int32 = 0 // 正常
	LevelWatched   int32 = 1 // 观察: 标注⚠️, 搜索权重-50%
	LevelThrottled int32 = 2 // 限流: 列表页隐藏, 不计入作者贡献
	LevelSealed    int32 = 3 // 封存: 仅作者和举报者可见
)

// Engine encapsulates degradation logic: quorum checking, level assignment, etc.
type Engine struct {
	flagModel  *model.FlagModel
	paperModel *model.PaperModel
	userModel  *model.UserModel
}

// NewEngine creates a new degradation engine
func NewEngine(fm *model.FlagModel, pm *model.PaperModel, um *model.UserModel) *Engine {
	return &Engine{
		flagModel:  fm,
		paperModel: pm,
		userModel:  um,
	}
}

// ProcessFlag handles a new flag submission: checks quorum and applies degradation
func (e *Engine) ProcessFlag(ctx context.Context, flag *model.Flag) (newLevel int32, err error) {
	// 1. Insert the flag
	_, err = e.flagModel.Insert(ctx, flag)
	if err != nil {
		return 0, err
	}

	// 2. Increment flag count on paper (if target is paper)
	if flag.TargetType == "paper" {
		if err := e.paperModel.IncrFlagCount(ctx, flag.TargetId); err != nil {
			log.Printf("[degradation] error incrementing flag_count for paper %d: %v", flag.TargetId, err)
		}
	}

	// 3. Check quorum and determine degradation level
	return e.EvaluateDegradation(ctx, flag.TargetType, flag.TargetId)
}

// EvaluateDegradation checks flag stats and determines proper degradation level
func (e *Engine) EvaluateDegradation(ctx context.Context, targetType string, targetId int64) (int32, error) {
	stats, err := e.flagModel.CountByTarget(ctx, targetType, targetId)
	if err != nil {
		return 0, err
	}

	// Calculate quorum threshold
	quorum := e.calcQuorum(ctx, targetType, targetId)

	var level int32
	switch {
	case stats.TotalCount >= quorum*2 || stats.WeightedSum >= 100:
		level = LevelSealed
	case stats.TotalCount >= quorum || stats.WeightedSum >= 50:
		level = LevelThrottled
	case stats.TotalCount >= 2 || stats.WeightedSum >= 10:
		level = LevelWatched
	default:
		level = LevelNormal
	}

	// Apply degradation if target is paper
	if targetType == "paper" && level > LevelNormal {
		if err := e.ApplyDegradation(ctx, targetId, level); err != nil {
			return level, err
		}

		// If fully degraded, resolve all pending flags as "degraded"
		if level >= LevelThrottled {
			_ = e.flagModel.ResolveByTarget(ctx, targetType, targetId, 1)
		}
	}

	return level, nil
}

// ApplyDegradation sets the degradation level on a paper
func (e *Engine) ApplyDegradation(ctx context.Context, paperId int64, level int32) error {
	return e.paperModel.UpdateDegradationLevel(ctx, paperId, level)
}

// calcQuorum returns the minimum number of flags needed for quorum
// quorum = max(3, sqrt(rating_count))
func (e *Engine) calcQuorum(ctx context.Context, targetType string, targetId int64) int {
	if targetType != "paper" {
		return 3 // default quorum for non-paper targets
	}

	paper, err := e.paperModel.FindById(ctx, targetId)
	if err != nil {
		return 3
	}

	q := int(math.Sqrt(float64(paper.RatingCount)))
	if q < 3 {
		return 3
	}
	return q
}

// CheckQuorum returns whether quorum has been reached for a target
func (e *Engine) CheckQuorum(ctx context.Context, targetType string, targetId int64) (bool, error) {
	stats, err := e.flagModel.CountByTarget(ctx, targetType, targetId)
	if err != nil {
		return false, err
	}
	quorum := e.calcQuorum(ctx, targetType, targetId)
	return stats.TotalCount >= quorum, nil
}
