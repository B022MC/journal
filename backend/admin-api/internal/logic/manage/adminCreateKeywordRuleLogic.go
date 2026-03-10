package manage

import (
	"context"
	"encoding/json"
	"errors"

	"journal/admin-api/internal/svc"
	"journal/admin-api/internal/types"
	"journal/common/consts"
	"journal/common/degradation"
	"journal/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminCreateKeywordRuleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminCreateKeywordRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminCreateKeywordRuleLogic {
	return &AdminCreateKeywordRuleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminCreateKeywordRuleLogic) AdminCreateKeywordRule(req *types.CreateKeywordRuleReq) (*types.CreateKeywordRuleResp, error) {
	actorUserId, err := requireKeywordRuleManageAccess(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}

	id, err := l.svcCtx.KeywordFilter.CreateRule(l.ctx, &model.KeywordRule{
		Pattern:       req.Pattern,
		MatchType:     req.MatchType,
		Category:      req.Category,
		Enabled:       1,
		CreatorUserId: actorUserId,
	})
	if err != nil {
		resp := &types.CreateKeywordRuleResp{
			Success: false,
			Message: err.Error(),
		}

		switch {
		case errors.Is(err, degradation.ErrDuplicateRule),
			errors.Is(err, degradation.ErrKeywordRuleEmpty),
			errors.Is(err, degradation.ErrInvalidCategory),
			errors.Is(err, degradation.ErrInvalidMatchType),
			errors.Is(err, degradation.ErrInvalidPattern):
			return resp, nil
		default:
			return nil, err
		}
	}

	detail, _ := json.Marshal(map[string]string{
		"pattern":    req.Pattern,
		"match_type": req.MatchType,
		"category":   req.Category,
	})
	recordAdminAuditLog(l.ctx, l.svcCtx, actorUserId, consts.PermAdminKeywordManage, "create keyword rule", "keyword_rule", id, string(detail))

	return &types.CreateKeywordRuleResp{
		Success: true,
		Message: "keyword rule created",
		Id:      id,
	}, nil
}
