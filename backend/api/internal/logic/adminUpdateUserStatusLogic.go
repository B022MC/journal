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

type AdminUpdateUserStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminUpdateUserStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUpdateUserStatusLogic {
	return &AdminUpdateUserStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminUpdateUserStatusLogic) AdminUpdateUserStatus(req *types.UpdateUserStatusReq) (resp *types.CommonResp, err error) {
	rpcResp, err := l.svcCtx.AdminRpc.UpdateUserStatus(l.ctx, &adminClient.UpdateUserStatusReq{
		UserId: req.UserId,
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
