package manage

import (
	"context"

	"journal/admin-api/internal/svc"
	"journal/admin-api/internal/types"
	"journal/common/consts"
	"journal/rpc/admin/adminClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminDeleteRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminDeleteRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminDeleteRoleLogic {
	return &AdminDeleteRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminDeleteRoleLogic) AdminDeleteRole(req *types.IdReq) (resp *types.CommonResp, err error) {
	if _, err := requireAdminPermission(l.ctx, l.svcCtx, consts.PermAdminRoleManage); err != nil {
		return nil, err
	}

	rpcResp, err := l.svcCtx.AdminRpc.DeleteRole(l.ctx, &adminClient.DeleteRoleReq{
		Id: req.Id,
	})
	if err != nil {
		return nil, err
	}

	return &types.CommonResp{
		Success: rpcResp.Success,
		Message: rpcResp.Message,
	}, nil
}
