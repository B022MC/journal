package logic

import (
	"context"
	"encoding/json"

	"journal/api/internal/svc"
	"journal/api/internal/types"
	"journal/rpc/user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserInfoLogic) GetUserInfo() (resp *types.UserInfo, err error) {
	userId, _ := l.ctx.Value("userId").(json.Number).Int64()
	rpcResp, err := l.svcCtx.UserRpc.GetUserInfo(l.ctx, &user.UserInfoReq{UserId: userId})
	if err != nil {
		return nil, err
	}
	return &types.UserInfo{
		Id:                rpcResp.UserInfo.Id,
		Username:          rpcResp.UserInfo.Username,
		Email:             rpcResp.UserInfo.Email,
		Nickname:          rpcResp.UserInfo.Nickname,
		Avatar:            rpcResp.UserInfo.Avatar,
		Role:              rpcResp.UserInfo.Role,
		ContributionScore: rpcResp.UserInfo.ContributionScore,
		CreatedAt:         rpcResp.UserInfo.CreatedAt,
	}, nil
}
