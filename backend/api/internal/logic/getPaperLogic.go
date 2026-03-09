package logic

import (
	"context"

	"journal/api/internal/svc"
	"journal/api/internal/types"
	"journal/rpc/paper/paper"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPaperLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPaperLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPaperLogic {
	return &GetPaperLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPaperLogic) GetPaper(req *types.IdReq) (resp *types.PaperItem, err error) {
	rpcResp, err := l.svcCtx.PaperRpc.GetPaper(l.ctx, &paper.GetPaperReq{Id: req.Id})
	if err != nil {
		return nil, err
	}
	result := toPaperType(rpcResp.Paper)
	return &result, nil
}
