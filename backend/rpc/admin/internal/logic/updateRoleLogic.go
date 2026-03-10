package logic

import (
	"context"

	"journal/rpc/admin/admin"
	"journal/rpc/admin/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateRoleLogic {
	return &UpdateRoleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateRoleLogic) UpdateRole(in *admin.UpdateRoleReq) (*admin.CommonResp, error) {
	err := l.svcCtx.AdminRBACModel.UpdateRole(l.ctx, in.Id, in.Name, in.Description, in.Status)
	if err != nil {
		return nil, err
	}
	return &admin.CommonResp{Success: true, Message: "role updated"}, nil
}
