package logic

import (
	"context"

	"journal/rpc/admin/admin"
	"journal/rpc/admin/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUserRolesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListUserRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserRolesLogic {
	return &ListUserRolesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListUserRolesLogic) ListUserRoles(in *admin.ListUserRolesReq) (*admin.ListUserRolesResp, error) {
	roles, err := l.svcCtx.AdminRBACModel.ListUserRolesByUserId(l.ctx, in.UserId)
	if err != nil {
		return nil, err
	}

	var items []*admin.RoleItem
	for _, r := range roles {
		items = append(items, &admin.RoleItem{
			Id:          r.Id,
			Code:        r.Code,
			Name:        r.Name,
			Description: r.Description,
			IsSuper:     r.IsSuper,
			Status:      r.Status,
			CreatedAt:   r.CreatedAt.Unix(),
			UpdatedAt:   r.UpdatedAt.Unix(),
		})
	}

	return &admin.ListUserRolesResp{Items: items}, nil
}
