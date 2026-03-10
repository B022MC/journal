package auth

import (
	"context"
	"errors"

	"journal/admin-api/internal/svc"
	"journal/admin-api/internal/types"
	"journal/common/consts"
	sjwt "journal/common/jwt"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type AdminLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminLoginLogic {
	return &AdminLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminLoginLogic) AdminLogin(req *types.AdminLoginReq) (resp *types.AdminLoginResp, err error) {
	adminUser, err := l.svcCtx.AdminRBAC.FindAdminUserByUsername(l.ctx, req.Username)
	if err != nil {
		l.Errorf("Failed to find admin user %s: %v", req.Username, err)
		return nil, errors.New("invalid credentials")
	}

	if adminUser.Status == 0 {
		return nil, errors.New("admin user is disabled")
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(adminUser.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Generate a new Admin Token using the admin-api's secret
	expireHrs := l.svcCtx.Config.Auth.AccessExpire / 3600
	if expireHrs == 0 {
		expireHrs = 72
	}

	token, expireAt, err := sjwt.GenerateToken(
		l.svcCtx.Config.Auth.AccessSecret,
		adminUser.Id,
		adminUser.Username,
		int32(consts.UserRoleAdmin),
		int(expireHrs),
	)
	if err != nil {
		return nil, err
	}

	return &types.AdminLoginResp{
		Token:    token,
		ExpireAt: expireAt,
	}, nil
}
