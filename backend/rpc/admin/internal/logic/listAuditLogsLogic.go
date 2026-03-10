package logic

import (
	"context"

	"journal/rpc/admin/admin"
	"journal/rpc/admin/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAuditLogsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAuditLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAuditLogsLogic {
	return &ListAuditLogsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListAuditLogsLogic) ListAuditLogs(in *admin.ListAuditLogsReq) (*admin.ListAuditLogsResp, error) {
	logs, total, err := l.svcCtx.AdminRBACModel.ListAuditLogs(l.ctx, int(in.Page), int(in.PageSize))
	if err != nil {
		return nil, err
	}

	var items []*admin.AuditLogItem
	for _, log := range logs {
		items = append(items, &admin.AuditLogItem{
			Id:             log.Id,
			ActorUserId:    log.ActorUserId,
			PermissionCode: log.PermissionCode,
			Action:         log.Action,
			TargetType:     log.TargetType,
			TargetId:       log.TargetId,
			Detail:         log.Detail,
			CreatedAt:      log.CreatedAt.Unix(),
		})
	}

	return &admin.ListAuditLogsResp{Items: items, Total: total}, nil
}
