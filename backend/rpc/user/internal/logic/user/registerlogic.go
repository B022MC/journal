package userlogic

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"journal/model"
	"journal/rpc/user/internal/svc"
	"journal/rpc/user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegisterLogic) Register(in *user.RegisterReq) (*user.RegisterResp, error) {
	// Check username uniqueness
	// 走主库检查唯一性（避免主从延迟导致重复注册）
	exists, err := l.svcCtx.UserModel.ExistsByUsernamePrimary(l.ctx, in.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("username already taken")
	}

	// Check email uniqueness
	exists, err = l.svcCtx.UserModel.ExistsByEmailPrimary(l.ctx, in.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("email already taken")
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	nickname := in.Nickname
	if nickname == "" {
		nickname = in.Username
	}

	id, err := l.svcCtx.UserModel.Insert(l.ctx, &model.User{
		Username:     in.Username,
		Email:        in.Email,
		PasswordHash: string(hash),
		Nickname:     nickname,
		Role:         0,
		Status:       1,
	})
	if err != nil {
		return nil, err
	}

	return &user.RegisterResp{Id: id}, nil
}
