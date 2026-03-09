package paperlogic

import (
	"context"

	"journal/rpc/paper/internal/svc"
	"journal/rpc/paper/paper"

	"github.com/zeromicro/go-zero/core/logx"
)

type IncrViewCountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIncrViewCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IncrViewCountLogic {
	return &IncrViewCountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *IncrViewCountLogic) IncrViewCount(in *paper.GetPaperReq) (*paper.CommonResp, error) {
	err := l.svcCtx.PaperModel.IncrViewCount(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &paper.CommonResp{Success: true, Message: "ok"}, nil
}
