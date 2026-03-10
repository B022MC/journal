package manage

import (
	"context"

	"journal/admin-api/internal/svc"
	"journal/admin-api/internal/types"
	"journal/common/consts"
	"journal/rpc/admin/adminClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminAssignUserRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminAssignUserRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminAssignUserRoleLogic {
	return &AdminAssignUserRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminAssignUserRoleLogic) AdminAssignUserRole(req *types.AssignUserRoleReq) (resp *types.CommonResp, err error) {
	if _, err := requireAdminPermission(l.ctx, l.svcCtx, consts.PermAdminRoleManage); err != nil {
		return nil, err
	}

	rpcResp, err := l.svcCtx.AdminRpc.AssignUserRole(l.ctx, &adminClient.AssignUserRoleReq{
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
