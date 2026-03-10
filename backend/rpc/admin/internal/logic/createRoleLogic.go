package logic

import (
	"context"

	"journal/rpc/admin/admin"
	"journal/rpc/admin/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateRoleLogic {
	return &CreateRoleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateRoleLogic) CreateRole(in *admin.CreateRoleReq) (*admin.CreateRoleResp, error) {
	id, err := l.svcCtx.AdminRBACModel.CreateRole(l.ctx, in.Code, in.Name, in.Description)
	if err != nil {
		return nil, err
	}
	return &admin.CreateRoleResp{Id: id}, nil
}
