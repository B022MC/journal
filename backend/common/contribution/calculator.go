package contribution

import (
	"context"
	"math"

	"journal/model"
)

// ZoneWeight maps zone names to their contribution weight multipliers
var ZoneWeight = map[string]float64{
	"latrine":     1.0,
	"septic_tank": 1.5,
	"stone":       2.5,
	"sediment":    4.0,
}

// Weight constants for aggregation
const (
	wAuthor   = 0.40
	wReviewer = 0.35
	wActivity = 0.15
	wDecay    = 0.10

	// Comment depth thresholds (byte length for UTF-8)
	depthShort = 100
	depthLong  = 300

	// Spam detection threshold: consecutive same-score ratings
	spamConsecutiveThreshold = 5
)

// Calculator computes user contribution scores based on authoring,
// reviewing, activity, and decay dimensions.
type Calculator struct {
	userModel   *model.UserModel
	paperModel  *model.PaperModel
	ratingModel *model.RatingModel
}

// NewCalculator creates a new contribution calculator
func NewCalculator(um *model.UserModel, pm *model.PaperModel, rm *model.RatingModel) *Calculator {
	return &Calculator{
		userModel:   um,
		paperModel:  pm,
		ratingModel: rm,
	}
}

// CalcForUser computes the overall contribution score for a single user
func (c *Calculator) CalcForUser(ctx context.Context, userId int64) (float64, error) {
	authorScore, err := c.CalcAuthorScore(ctx, userId)
	if err != nil {
		return 0, err
	}

	reviewerScore, err := c.CalcReviewerScore(ctx, userId)
	if err != nil {
		return 0, err
	}

	activityScore, err := c.CalcActivityScore(ctx, userId)
	if err != nil {
		return 0, err
	}

	decayPenalty, err := c.CalcDecayPenalty(ctx, userId)
	if err != nil {
		return 0, err
	}

	total := wAuthor*authorScore + wReviewer*reviewerScore + wActivity*activityScore - wDecay*decayPenalty
	if total < 0 {
		total = 0
	}
	return math.Round(total*100) / 100, nil
}

// CalcAuthorScore = Σ paper.shit_score × zone_weight(paper.zone)
func (c *Calculator) CalcAuthorScore(ctx context.Context, userId int64) (float64, error) {
	papers, _, err := c.paperModel.ListByAuthor(ctx, userId, 1, 1000) // get all papers
	if err != nil {
		return 0, err
	}
	var score float64
	for _, p := range papers {
		w, ok := ZoneWeight[p.Zone]
		if !ok {
			w = 1.0
		}
		score += p.ShitScore * w
	}
	return score, nil
}

// CalcReviewerScore = Σ review_accuracy(rating) × depth_bonus(comment_len)
// Spam ratings are penalized (×0.1)
func (c *Calculator) CalcReviewerScore(ctx context.Context, userId int64) (float64, error) {
	ratings, err := c.ratingModel.ListAllByUser(ctx, userId)
	if err != nil {
		return 0, err
	}

	isSpammer, err := c.DetectSpam(ctx, userId)
	if err != nil {
		return 0, err
	}

	var score float64
	for _, r := range ratings {
		// Review accuracy: 1 - |user_score - avg| / 10, clamped to [0.1, 1.0]
		paper, err := c.paperModel.FindById(ctx, r.PaperId)
		if err != nil {
			continue // skip if paper deleted
		}
		accuracy := 1.0 - math.Abs(float64(r.Score)-paper.AvgRating)/10.0
		if accuracy < 0.1 {
			accuracy = 0.1
		}

		// Depth bonus based on comment length
		depth := depthBonus(r.Comment)

		contribution := accuracy * depth
		if isSpammer {
			contribution *= 0.1
		}
		if r.Comment == "" {
			contribution *= 0.3
		}
		score += contribution
	}
	return score, nil
}

// CalcActivityScore = log(reviews_30d + 1) × log(logins_30d + 1)
// For simplicity, we approximate logins with last_active_at presence
func (c *Calculator) CalcActivityScore(ctx context.Context, userId int64) (float64, error) {
	user, err := c.userModel.FindById(ctx, userId)
	if err != nil {
		return 0, err
	}
	reviews30d := float64(user.ReviewCount30d)
	// Approximate login activity: if active in last 30 days, count as active
	loginFactor := 1.0
	if user.LastActiveAt.Valid {
		loginFactor = 2.0 // simplified: if they have recent activity
	}
	return math.Log(reviews30d+1) * math.Log(loginFactor+1), nil
}

// CalcDecayPenalty = max(0, days_inactive - 30) × 0.01
func (c *Calculator) CalcDecayPenalty(ctx context.Context, userId int64) (float64, error) {
	inactiveDays, err := c.userModel.GetInactiveDays(ctx, userId)
	if err != nil {
		return 0, err
	}
	penalty := float64(inactiveDays-30) * 0.01
	if penalty < 0 {
		penalty = 0
	}
	return penalty, nil
}

// DetectSpam checks if a user has ≥ spamConsecutiveThreshold consecutive same-score ratings
func (c *Calculator) DetectSpam(ctx context.Context, userId int64) (bool, error) {
	count, err := c.ratingModel.CountConsecutiveSameScore(ctx, userId)
	if err != nil {
		return false, err
	}
	return count >= spamConsecutiveThreshold, nil
}

// depthBonus returns a multiplier based on comment length
func depthBonus(comment string) float64 {
	length := len(comment) // byte length, reasonable proxy for UTF-8
	if length >= depthLong {
		return 1.5
	}
	if length >= depthShort {
		return 1.2
	}
	return 1.0
}

// RoleForScore returns the appropriate role based on contribution score
// 0~10 = member(0), 10~50 = scooper(1), 50~200 = editor(2), 200+ = admin(3)
func RoleForScore(score float64) int32 {
	switch {
	case score >= 200:
		return 3 // admin
	case score >= 50:
		return 2 // editor
	case score >= 10:
		return 1 // scooper
	default:
		return 0 // member
	}
}
