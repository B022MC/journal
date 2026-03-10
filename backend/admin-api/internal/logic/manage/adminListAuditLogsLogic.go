package manage

import (
	"context"

	"journal/admin-api/internal/svc"
	"journal/admin-api/internal/types"
	"journal/common/consts"
	"journal/rpc/admin/adminClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminListAuditLogsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminListAuditLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListAuditLogsLogic {
	return &AdminListAuditLogsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminListAuditLogsLogic) AdminListAuditLogs(req *types.PageReq) (resp *types.ListAuditLogsResp, err error) {
	if _, err := requireAdminPermission(l.ctx, l.svcCtx, consts.PermAdminAuditView); err != nil {
		return nil, err
	}

	rpcResp, err := l.svcCtx.AdminRpc.ListAuditLogs(l.ctx, &adminClient.ListAuditLogsReq{
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
	})
	if err != nil {
		return nil, err
	}

	resp = &types.ListAuditLogsResp{
		Total: rpcResp.Total,
		Items: make([]types.AuditLogItem, 0, len(rpcResp.Items)),
	}
	for _, item := range rpcResp.Items {
		resp.Items = append(resp.Items, types.AuditLogItem{
			Id:             item.Id,
			ActorUserId:    item.ActorUserId,
			PermissionCode: item.PermissionCode,
			Action:         item.Action,
			TargetType:     item.TargetType,
			TargetId:       item.TargetId,
			Detail:         item.Detail,
			CreatedAt:      item.CreatedAt,
		})
	}

	return resp, nil
}
