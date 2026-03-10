package manage

import (
	"context"

	"journal/admin-api/internal/svc"
	"journal/admin-api/internal/types"
	"journal/common/consts"
	"journal/rpc/admin/adminClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminCreateRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminCreateRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminCreateRoleLogic {
	return &AdminCreateRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminCreateRoleLogic) AdminCreateRole(req *types.CreateRoleReq) (resp *types.CreateRoleResp, err error) {
	if _, err := requireAdminPermission(l.ctx, l.svcCtx, consts.PermAdminRoleManage); err != nil {
		return nil, err
	}

	rpcResp, err := l.svcCtx.AdminRpc.CreateRole(l.ctx, &adminClient.CreateRoleReq{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return nil, err
	}

	return &types.CreateRoleResp{Id: rpcResp.Id}, nil
}
