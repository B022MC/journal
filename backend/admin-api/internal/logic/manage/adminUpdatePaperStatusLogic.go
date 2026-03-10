package manage

import (
	"context"

	"journal/admin-api/internal/svc"
	"journal/admin-api/internal/types"
	"journal/common/consts"
	"journal/rpc/admin/adminClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminUpdatePaperStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminUpdatePaperStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUpdatePaperStatusLogic {
	return &AdminUpdatePaperStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminUpdatePaperStatusLogic) AdminUpdatePaperStatus(req *types.UpdatePaperStatusReq) (resp *types.CommonResp, err error) {
	if _, err := requireAdminPermission(l.ctx, l.svcCtx, consts.PermAdminPaperView); err != nil {
		return nil, err
	}

	rpcResp, err := l.svcCtx.AdminRpc.UpdatePaperStatus(l.ctx, &adminClient.UpdatePaperStatusReq{
		PaperId: req.PaperId,
		Status:  req.Status,
	})
	if err != nil {
		return nil, err
	}

	return &types.CommonResp{
		Success: rpcResp.Success,
		Message: rpcResp.Message,
	}, nil
}
