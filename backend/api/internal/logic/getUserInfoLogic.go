package logic

import (
	"context"

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
	userId, err := currentUserID(l.ctx)
	if err != nil {
		return nil, err
	}
	rpcResp, err := l.svcCtx.UserRpc.GetUserInfo(l.ctx, &user.UserInfoReq{UserId: userId})
	if err != nil {
		return nil, err
	}

	adminPermissions, permErr := listAdminPermissionCodes(l.ctx, l.svcCtx, userId)
	if permErr != nil {
		l.Errorf("load admin permissions for user %d: %v", userId, permErr)
		adminPermissions = []string{}
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
		AdminPermissions:  adminPermissions,
	}, nil
}
