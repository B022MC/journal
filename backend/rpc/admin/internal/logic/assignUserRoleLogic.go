package logic

import (
	"context"

	"journal/rpc/admin/admin"
	"journal/rpc/admin/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AssignUserRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAssignUserRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignUserRoleLogic {
	return &AssignUserRoleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AssignUserRoleLogic) AssignUserRole(in *admin.AssignUserRoleReq) (*admin.CommonResp, error) {
	err := l.svcCtx.AdminRBACModel.AssignUserRole(l.ctx, in.UserId, in.RoleId)
	if err != nil {
		return nil, err
	}
	return &admin.CommonResp{Success: true, Message: "role assigned"}, nil
}
