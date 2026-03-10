package logic

import (
	"context"

	"journal/rpc/admin/admin"
	"journal/rpc/admin/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListRolesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRolesLogic {
	return &ListRolesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListRolesLogic) ListRoles(in *admin.ListRolesReq) (*admin.ListRolesResp, error) {
	roles, err := l.svcCtx.AdminRBACModel.ListRoles(l.ctx)
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

	return &admin.ListRolesResp{Items: items}, nil
}
