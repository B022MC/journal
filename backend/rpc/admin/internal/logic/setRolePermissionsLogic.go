package logic

import (
	"context"

	"journal/rpc/admin/admin"
	"journal/rpc/admin/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetRolePermissionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetRolePermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetRolePermissionsLogic {
	return &SetRolePermissionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetRolePermissionsLogic) SetRolePermissions(in *admin.SetRolePermissionsReq) (*admin.CommonResp, error) {
	err := l.svcCtx.AdminRBACModel.SetRolePermissions(l.ctx, in.RoleId, in.PermissionIds)
	if err != nil {
		return nil, err
	}
	return &admin.CommonResp{Success: true, Message: "role permissions updated"}, nil
}
