package logic

import (
	"context"

	"journal/rpc/admin/admin"
	"journal/rpc/admin/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RecordAuditLogLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRecordAuditLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecordAuditLogLogic {
	return &RecordAuditLogLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Audit
func (l *RecordAuditLogLogic) RecordAuditLog(in *admin.RecordAuditLogReq) (*admin.CommonResp, error) {
	query := "INSERT INTO `adm_audit_log` (`actor_user_id`, `permission_code`, `action`, `target_type`, `target_id`, `detail`) VALUES (?, ?, ?, ?, ?, ?)"
	_, err := l.svcCtx.DBConn.ExecCtx(l.ctx, query, in.ActorUserId, in.PermissionCode, in.Action, in.TargetType, in.TargetId, in.Detail)
	if err != nil {
		return &admin.CommonResp{Success: false, Message: err.Error()}, err
	}
	return &admin.CommonResp{Success: true, Message: "success"}, nil
}
