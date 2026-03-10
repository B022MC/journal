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

	permissions, err := l.svcCtx.AdminRBAC.ListPermissionCodesByUserId(l.ctx, rpcResp.UserInfo.Id)
	if err != nil {
		return nil, err
	}

	if cacheErr := l.svcCtx.Cache.SetContributionScore(l.ctx, rpcResp.UserInfo.Id, rpcResp.UserInfo.ContributionScore); cacheErr != nil {
		l.Errorf("warm contribution cache failed for user=%d: %v", rpcResp.UserInfo.Id, cacheErr)
	}

	achievements, err := l.svcCtx.AchievementService.SyncAndList(l.ctx, rpcResp.UserInfo.Id)
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
			AdminPermissions:  permissions,
			Achievements:      toAchievementBadges(achievements),
		},
	}, nil
}
