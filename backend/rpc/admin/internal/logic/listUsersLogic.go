package logic

import (
	"context"
	"fmt"

	"journal/rpc/admin/admin"
	"journal/rpc/admin/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUsersLogic {
	return &ListUsersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// User
func (l *ListUsersLogic) ListUsers(in *admin.ListUsersReq) (*admin.ListUsersResp, error) {
	users, total, err := l.svcCtx.UserModel.ListUsersPaginatedAdmin(l.ctx, int(in.Page), int(in.PageSize))
	if err != nil {
		return nil, err
	}

	var items []*admin.UserItem
	for _, u := range users {
		items = append(items, &admin.UserItem{
			Id:                u.Id,
			Username:          u.Username,
			Email:             u.Email,
			Nickname:          u.Nickname,
			Role:              u.Role,
			ContributionScore: fmt.Sprintf("%.2f", u.ContributionScore),
			Status:            u.Status,
			CreatedAt:         u.CreatedAt.Unix(),
		})
	}

	return &admin.ListUsersResp{
		Items: items,
		Total: total,
	}, nil
}
