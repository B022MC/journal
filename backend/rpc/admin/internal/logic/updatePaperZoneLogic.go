package logic

import (
	"context"

	"journal/rpc/admin/admin"
	"journal/rpc/admin/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdatePaperZoneLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdatePaperZoneLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePaperZoneLogic {
	return &UpdatePaperZoneLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdatePaperZoneLogic) UpdatePaperZone(in *admin.UpdatePaperZoneReq) (*admin.CommonResp, error) {
	err := l.svcCtx.PaperModel.UpdatePaperZoneAdmin(l.ctx, in.PaperId, in.Zone)
	if err != nil {
		return &admin.CommonResp{Success: false, Message: err.Error()}, err
	}
	return &admin.CommonResp{Success: true, Message: "success"}, nil
}
