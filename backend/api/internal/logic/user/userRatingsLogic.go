// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"journal/api/internal/svc"
	"journal/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserRatingsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserRatingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserRatingsLogic {
	return &UserRatingsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserRatingsLogic) UserRatings(req *types.PageReq) error {
	// todo: add your logic here and delete this line

	return nil
}
