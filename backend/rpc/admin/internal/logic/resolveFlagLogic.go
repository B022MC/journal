package logic

import (
	"context"

	"journal/rpc/admin/admin"
	"journal/rpc/admin/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ResolveFlagLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewResolveFlagLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResolveFlagLogic {
	return &ResolveFlagLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ResolveFlagLogic) ResolveFlag(in *admin.ResolveFlagReq) (*admin.CommonResp, error) {
	err := l.svcCtx.FlagModel.UpdateStatus(l.ctx, in.FlagId, in.Status)
	if err != nil {
		return &admin.CommonResp{Success: false, Message: err.Error()}, err
	}
	return &admin.CommonResp{Success: true, Message: "success"}, nil
}
