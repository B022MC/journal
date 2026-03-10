package manage

import (
	"context"

	"journal/admin-api/internal/svc"
	"journal/admin-api/internal/types"
	"journal/common/consts"
	"journal/rpc/admin/adminClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminListPermissionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminListPermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListPermissionsLogic {
	return &AdminListPermissionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminListPermissionsLogic) AdminListPermissions() (resp *types.ListPermissionsResp, err error) {
	if _, err := requireAdminPermission(l.ctx, l.svcCtx, consts.PermAdminRoleView); err != nil {
		return nil, err
	}

	rpcResp, err := l.svcCtx.AdminRpc.ListPermissions(l.ctx, &adminClient.ListPermissionsReq{})
	if err != nil {
		return nil, err
	}

	resp = &types.ListPermissionsResp{
		Items: make([]types.PermissionItem, 0, len(rpcResp.Items)),
	}
	for _, item := range rpcResp.Items {
		resp.Items = append(resp.Items, types.PermissionItem{
			Id:          item.Id,
			Code:        item.Code,
			Name:        item.Name,
			Module:      item.Module,
			Resource:    item.Resource,
			Action:      item.Action,
			Description: item.Description,
			Status:      item.Status,
		})
	}

	return resp, nil
}
