package logic

import (
	"context"

	"journal/rpc/admin/admin"
	"journal/rpc/admin/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteRoleLogic {
	return &DeleteRoleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteRoleLogic) DeleteRole(in *admin.DeleteRoleReq) (*admin.CommonResp, error) {
	err := l.svcCtx.AdminRBACModel.DeleteRole(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &admin.CommonResp{Success: true, Message: "role deleted"}, nil
}
