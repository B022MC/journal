package userlogic

import (
	"context"
	"fmt"

	"journal/rpc/user/internal/svc"
	"journal/rpc/user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserInfoLogic) GetUserInfo(in *user.UserInfoReq) (*user.UserInfoResp, error) {
	u, err := l.svcCtx.UserModel.FindById(l.ctx, in.UserId)
	if err != nil {
		return nil, err
	}

	return &user.UserInfoResp{
		UserInfo: &user.UserInfo{
			Id:                u.Id,
			Username:          u.Username,
			Email:             u.Email,
			Nickname:          u.Nickname,
			Avatar:            u.Avatar,
			Role:              u.Role,
			ContributionScore: fmt.Sprintf("%.2f", u.ContributionScore),
			CreatedAt:         u.CreatedAt.Unix(),
		},
	}, nil
}
