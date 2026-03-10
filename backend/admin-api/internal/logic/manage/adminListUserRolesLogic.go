package manage

import (
	"context"

	"journal/admin-api/internal/svc"
	"journal/admin-api/internal/types"
	"journal/common/consts"
	"journal/rpc/admin/adminClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminListUserRolesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminListUserRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListUserRolesLogic {
	return &AdminListUserRolesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminListUserRolesLogic) AdminListUserRoles(req *types.IdReq) (resp *types.ListUserRolesResp, err error) {
	if _, err := requireAdminPermission(l.ctx, l.svcCtx, consts.PermAdminRoleView); err != nil {
		return nil, err
	}

	rpcResp, err := l.svcCtx.AdminRpc.ListUserRoles(l.ctx, &adminClient.ListUserRolesReq{
		UserId: req.Id,
	})
	if err != nil {
		return nil, err
	}

	resp = &types.ListUserRolesResp{
		Items: make([]types.RoleItem, 0, len(rpcResp.Items)),
	}
	for _, item := range rpcResp.Items {
		resp.Items = append(resp.Items, types.RoleItem{
			Id:          item.Id,
			Code:        item.Code,
			Name:        item.Name,
			Description: item.Description,
			IsSuper:     item.IsSuper,
			Status:      item.Status,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		})
	}

	return resp, nil
}
