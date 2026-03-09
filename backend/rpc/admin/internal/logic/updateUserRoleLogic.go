package logic

import (
	"context"

	"journal/rpc/admin/admin"
	"journal/rpc/admin/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserRoleLogic {
	return &UpdateUserRoleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateUserRoleLogic) UpdateUserRole(in *admin.UpdateUserRoleReq) (*admin.CommonResp, error) {
	err := l.svcCtx.UserModel.UpdateUserRoleAdmin(l.ctx, in.UserId, in.Role)
	if err != nil {
		return &admin.CommonResp{Success: false, Message: err.Error()}, err
	}
	return &admin.CommonResp{Success: true, Message: "success"}, nil
}
