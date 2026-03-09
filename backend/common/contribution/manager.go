package contribution

import (
	"context"
	"math"

	"journal/model"
)

// Manager coordinates contribution score recomputation and role sync.
type Manager struct {
	calc      *Calculator
	userModel *model.UserModel
}

// NewManager creates a contribution manager.
func NewManager(um *model.UserModel, pm *model.PaperModel, rm *model.RatingModel) *Manager {
	return &Manager{
		calc:      NewCalculator(um, pm, rm),
		userModel: um,
	}
}

// SyncUser recomputes one user's contribution score and synchronizes role.
func (m *Manager) SyncUser(ctx context.Context, userId int64) (float64, int32, error) {
	score, err := m.calc.CalcForUser(ctx, userId)
	if err != nil {
		return 0, 0, err
	}

	score = math.Round(score*100) / 100
	if err := m.userModel.UpdateContributionScore(ctx, userId, score); err != nil {
		return 0, 0, err
	}

	role := RoleForScore(score)
	if err := m.userModel.AutoAssignRole(ctx, userId, role); err != nil {
		return 0, 0, err
	}

	return score, role, nil
}
