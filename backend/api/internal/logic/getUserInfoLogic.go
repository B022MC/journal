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
	uid := currentUserID(l.ctx)

	rpcResp, err := l.svcCtx.UserRpc.GetUserInfo(l.ctx, &user.UserInfoReq{UserId: uid})
	if err != nil {
		return nil, err
	}

	permissions, err := l.svcCtx.AdminRBAC.ListPermissionCodesByUserId(l.ctx, uid)
	if err != nil {
		return nil, err
	}

	if cacheErr := l.svcCtx.Cache.SetContributionScore(l.ctx, rpcResp.UserInfo.Id, rpcResp.UserInfo.ContributionScore); cacheErr != nil {
		l.Errorf("warm contribution cache failed for user=%d: %v", rpcResp.UserInfo.Id, cacheErr)
	}

	achievements, err := l.svcCtx.AchievementService.SyncAndList(l.ctx, uid)
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
		AdminPermissions:  permissions,
		Achievements:      toAchievementBadges(achievements),
	}, nil
}
