package manage

import (
	"context"

	"journal/admin-api/internal/svc"
	"journal/admin-api/internal/types"
	"journal/common/consts"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminListKeywordRulesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminListKeywordRulesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListKeywordRulesLogic {
	return &AdminListKeywordRulesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminListKeywordRulesLogic) AdminListKeywordRules() (*types.ListKeywordRulesResp, error) {
	if _, err := requireAdminPermission(l.ctx, l.svcCtx, consts.PermAdminKeywordView); err != nil {
		return nil, err
	}

	rules, err := l.svcCtx.KeywordFilter.ListRules(l.ctx, false)
	if err != nil {
		return nil, err
	}

	resp := &types.ListKeywordRulesResp{
		Items: make([]types.KeywordRuleItem, 0, len(rules)),
	}
	for _, item := range rules {
		resp.Items = append(resp.Items, types.KeywordRuleItem{
			Id:            item.Id,
			Pattern:       item.Pattern,
			MatchType:     item.MatchType,
			Category:      item.Category,
			Enabled:       item.Enabled,
			CreatorUserId: item.CreatorUserId,
			CreatedAt:     item.CreatedAt.Unix(),
			UpdatedAt:     item.UpdatedAt.Unix(),
		})
	}

	return resp, nil
}
