package userlogic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	sjwt "journal/common/jwt"
	"journal/rpc/user/internal/svc"
	"journal/rpc/user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogic) Login(in *user.LoginReq) (*user.LoginResp, error) {
	u, err := l.svcCtx.UserModel.FindByUsername(l.ctx, in.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	if u.Status == 0 {
		return nil, errors.New("user banned")
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(in.Password))
	if err != nil {
		return nil, errors.New("wrong password")
	}

	token, expireAt, err := sjwt.GenerateToken(
		l.svcCtx.Config.JwtSecret,
		u.Id, u.Username, u.Role,
		l.svcCtx.Config.JwtExpireHrs,
	)
	if err != nil {
		return nil, err
	}

	return &user.LoginResp{
		Id:       u.Id,
		Token:    token,
		ExpireAt: expireAt,
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
