// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"journal/api/internal/svc"
	"journal/api/internal/types"
	"journal/rpc/admin/adminClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminResolveFlagLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminResolveFlagLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminResolveFlagLogic {
	return &AdminResolveFlagLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminResolveFlagLogic) AdminResolveFlag(req *types.ResolveFlagReq) (resp *types.CommonResp, err error) {
	if _, err := requireAdminPermission(l.ctx, l.svcCtx, permissionAdminPaperZoneUpdate); err != nil {
		return nil, err
	}

	rpcResp, err := l.svcCtx.AdminRpc.ResolveFlag(l.ctx, &adminClient.ResolveFlagReq{
		FlagId: req.FlagId,
		Status: req.Status,
	})
	if err != nil {
		return nil, err
	}

	return &types.CommonResp{
		Success: rpcResp.Success,
		Message: rpcResp.Message,
	}, nil
}
