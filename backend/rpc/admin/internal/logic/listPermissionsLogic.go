package logic

import (
	"context"

	"journal/rpc/admin/admin"
	"journal/rpc/admin/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPermissionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListPermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPermissionsLogic {
	return &ListPermissionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListPermissionsLogic) ListPermissions(in *admin.ListPermissionsReq) (*admin.ListPermissionsResp, error) {
	perms, err := l.svcCtx.AdminRBACModel.ListPermissions(l.ctx)
	if err != nil {
		return nil, err
	}

	var items []*admin.PermissionItem
	for _, p := range perms {
		items = append(items, &admin.PermissionItem{
			Id:          p.Id,
			Code:        p.Code,
			Name:        p.Name,
			Module:      p.Module,
			Resource:    p.Resource,
			Action:      p.Action,
			Description: p.Description,
			Status:      p.Status,
		})
	}

	return &admin.ListPermissionsResp{Items: items}, nil
}
