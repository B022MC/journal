package logic

import (
	"context"

	"journal/api/internal/svc"
	"journal/api/internal/types"
	"journal/rpc/user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPublicUserProfileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPublicUserProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPublicUserProfileLogic {
	return &GetPublicUserProfileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPublicUserProfileLogic) GetPublicUserProfile(req *types.IdReq) (resp *types.PublicUserProfile, err error) {
	rpcResp, err := l.svcCtx.UserRpc.GetUserInfo(l.ctx, &user.UserInfoReq{UserId: req.Id})
	if err != nil {
		return nil, err
	}

	if cacheErr := l.svcCtx.Cache.SetContributionScore(l.ctx, rpcResp.UserInfo.Id, rpcResp.UserInfo.ContributionScore); cacheErr != nil {
		l.Errorf("warm contribution cache failed for user=%d: %v", rpcResp.UserInfo.Id, cacheErr)
	}

	achievements, err := l.svcCtx.AchievementService.SyncAndList(l.ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &types.PublicUserProfile{
		Id:                rpcResp.UserInfo.Id,
		Username:          rpcResp.UserInfo.Username,
		Nickname:          rpcResp.UserInfo.Nickname,
		Avatar:            rpcResp.UserInfo.Avatar,
		Role:              rpcResp.UserInfo.Role,
		ContributionScore: rpcResp.UserInfo.ContributionScore,
		CreatedAt:         rpcResp.UserInfo.CreatedAt,
		Achievements:      toAchievementBadges(achievements),
	}, nil
}
