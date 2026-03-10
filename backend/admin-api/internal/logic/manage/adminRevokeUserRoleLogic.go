package manage

import (
	"context"

	"journal/admin-api/internal/svc"
	"journal/admin-api/internal/types"
	"journal/common/consts"
	"journal/rpc/admin/adminClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminRevokeUserRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminRevokeUserRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminRevokeUserRoleLogic {
	return &AdminRevokeUserRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminRevokeUserRoleLogic) AdminRevokeUserRole(req *types.RevokeUserRoleReq) (resp *types.CommonResp, err error) {
	if _, err := requireAdminPermission(l.ctx, l.svcCtx, consts.PermAdminRoleManage); err != nil {
		return nil, err
	}

	rpcResp, err := l.svcCtx.AdminRpc.RevokeUserRole(l.ctx, &adminClient.RevokeUserRoleReq{
		UserId: req.UserId,
		RoleId: req.RoleId,
	})
	if err != nil {
		return nil, err
	}

	return &types.CommonResp{
		Success: rpcResp.Success,
		Message: rpcResp.Message,
	}, nil
}
