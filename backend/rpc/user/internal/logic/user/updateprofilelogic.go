package userlogic

import (
	"context"

	"journal/rpc/user/internal/svc"
	"journal/rpc/user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProfileLogic {
	return &UpdateProfileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateProfileLogic) UpdateProfile(in *user.UpdateProfileReq) (*user.CommonResp, error) {
	err := l.svcCtx.UserModel.UpdateProfile(l.ctx, in.UserId, in.Nickname, in.Avatar)
	if err != nil {
		return nil, err
	}

	return &user.CommonResp{
		Success: true,
		Message: "profile updated",
	}, nil
}
