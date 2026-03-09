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

type AdminUpdateUserRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminUpdateUserRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUpdateUserRoleLogic {
	return &AdminUpdateUserRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminUpdateUserRoleLogic) AdminUpdateUserRole(req *types.UpdateUserRoleReq) (resp *types.CommonResp, err error) {
	rpcResp, err := l.svcCtx.AdminRpc.UpdateUserRole(l.ctx, &adminClient.UpdateUserRoleReq{
		UserId: req.UserId,
		Role:   req.Role,
	})
	if err != nil {
		return nil, err
	}

	return &types.CommonResp{
		Success: rpcResp.Success,
		Message: rpcResp.Message,
	}, nil
}
