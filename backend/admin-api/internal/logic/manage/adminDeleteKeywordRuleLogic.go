package manage

import (
	"context"
	"errors"

	"journal/admin-api/internal/svc"
	"journal/admin-api/internal/types"
	"journal/common/consts"
	"journal/common/degradation"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminDeleteKeywordRuleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminDeleteKeywordRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminDeleteKeywordRuleLogic {
	return &AdminDeleteKeywordRuleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminDeleteKeywordRuleLogic) AdminDeleteKeywordRule(req *types.IdReq) (*types.CommonResp, error) {
	actorUserId, err := requireKeywordRuleManageAccess(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}

	if err := l.svcCtx.KeywordFilter.DeleteRule(l.ctx, req.Id); err != nil {
		if errors.Is(err, degradation.ErrKeywordRuleGone) {
			return &types.CommonResp{
				Success: false,
				Message: err.Error(),
			}, nil
		}
		return nil, err
	}

	recordAdminAuditLog(l.ctx, l.svcCtx, actorUserId, consts.PermAdminKeywordManage, "delete keyword rule", "keyword_rule", req.Id, "")

	return &types.CommonResp{
		Success: true,
		Message: "keyword rule deleted",
	}, nil
}
