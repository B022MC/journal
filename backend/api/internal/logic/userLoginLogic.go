package logic

import (
	"context"

	"journal/api/internal/svc"
	"journal/api/internal/types"
	"journal/rpc/user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserLoginLogic {
	return &UserLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserLoginLogic) UserLogin(req *types.LoginReq) (resp *types.LoginResp, err error) {
	rpcResp, err := l.svcCtx.UserRpc.Login(l.ctx, &user.LoginReq{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &types.LoginResp{
		Token:    rpcResp.Token,
		ExpireAt: rpcResp.ExpireAt,
		UserInfo: types.UserInfo{
			Id:                rpcResp.UserInfo.Id,
			Username:          rpcResp.UserInfo.Username,
			Email:             rpcResp.UserInfo.Email,
			Nickname:          rpcResp.UserInfo.Nickname,
			Avatar:            rpcResp.UserInfo.Avatar,
			Role:              rpcResp.UserInfo.Role,
			ContributionScore: rpcResp.UserInfo.ContributionScore,
			CreatedAt:         rpcResp.UserInfo.CreatedAt,
		},
	}, nil
}
