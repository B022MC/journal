package logic

import (
	"context"

	"journal/rpc/admin/admin"
	"journal/rpc/admin/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRolePermissionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetRolePermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRolePermissionsLogic {
	return &GetRolePermissionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetRolePermissionsLogic) GetRolePermissions(in *admin.GetRolePermissionsReq) (*admin.GetRolePermissionsResp, error) {
	ids, err := l.svcCtx.AdminRBACModel.ListRolePermissionIds(l.ctx, in.RoleId)
	if err != nil {
		return nil, err
	}
	return &admin.GetRolePermissionsResp{PermissionIds: ids}, nil
}
