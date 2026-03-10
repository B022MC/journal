package manage

import (
	"context"

	"journal/admin-api/internal/svc"
	"journal/admin-api/internal/types"
	"journal/common/consts"
	"journal/rpc/admin/adminClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminGetRolePermissionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminGetRolePermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminGetRolePermissionsLogic {
	return &AdminGetRolePermissionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminGetRolePermissionsLogic) AdminGetRolePermissions(req *types.IdReq) (resp *types.GetRolePermissionsResp, err error) {
	if _, err := requireAdminPermission(l.ctx, l.svcCtx, consts.PermAdminRoleView); err != nil {
		return nil, err
	}

	rpcResp, err := l.svcCtx.AdminRpc.GetRolePermissions(l.ctx, &adminClient.GetRolePermissionsReq{
		RoleId: req.Id,
	})
	if err != nil {
		return nil, err
	}

	return &types.GetRolePermissionsResp{
		PermissionIds: rpcResp.PermissionIds,
	}, nil
}
