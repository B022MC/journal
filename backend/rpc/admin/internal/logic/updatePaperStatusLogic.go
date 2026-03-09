package logic

import (
	"context"

	"journal/rpc/admin/admin"
	"journal/rpc/admin/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdatePaperStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdatePaperStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePaperStatusLogic {
	return &UpdatePaperStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdatePaperStatusLogic) UpdatePaperStatus(in *admin.UpdatePaperStatusReq) (*admin.CommonResp, error) {
	err := l.svcCtx.PaperModel.UpdatePaperStatusAdmin(l.ctx, in.PaperId, in.Status)
	if err != nil {
		return &admin.CommonResp{Success: false, Message: err.Error()}, err
	}
	return &admin.CommonResp{Success: true, Message: "success"}, nil
}
