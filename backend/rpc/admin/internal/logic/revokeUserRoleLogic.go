package logic

import (
	"context"

	"journal/rpc/admin/admin"
	"journal/rpc/admin/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RevokeUserRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRevokeUserRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RevokeUserRoleLogic {
	return &RevokeUserRoleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RevokeUserRoleLogic) RevokeUserRole(in *admin.RevokeUserRoleReq) (*admin.CommonResp, error) {
	err := l.svcCtx.AdminRBACModel.RevokeUserRole(l.ctx, in.UserId, in.RoleId)
	if err != nil {
		return nil, err
	}
	return &admin.CommonResp{Success: true, Message: "role revoked"}, nil
}
