package logic

import (
	"context"

	"journal/rpc/admin/admin"
	"journal/rpc/admin/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserStatusLogic {
	return &UpdateUserStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateUserStatusLogic) UpdateUserStatus(in *admin.UpdateUserStatusReq) (*admin.CommonResp, error) {
	err := l.svcCtx.UserModel.UpdateUserStatus(l.ctx, in.UserId, in.Status)
	if err != nil {
		return &admin.CommonResp{Success: false, Message: err.Error()}, err
	}
	return &admin.CommonResp{Success: true, Message: "success"}, nil
}
