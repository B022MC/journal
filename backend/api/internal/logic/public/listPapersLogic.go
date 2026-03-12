// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package public

import (
	"context"

	"journal/api/internal/svc"
	"journal/api/internal/types"
	"journal/rpc/paper/paper"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPapersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListPapersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPapersLogic {
	return &ListPapersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListPapersLogic) ListPapers(req *types.ListPapersReq) (resp *types.ListPapersResp, err error) {
	rpcResp, err := l.svcCtx.PaperRpc.ListPapers(l.ctx, &paper.ListPapersReq{
		Zone:       req.Zone,
		Discipline: req.Discipline,
		Sort:       req.Sort,
		Page:       int32(req.Page),
		PageSize:   int32(req.PageSize),
	})
	if err != nil {
		return nil, err
	}
	return &types.ListPapersResp{
		Items: toPaperItems(rpcResp.Items),
		Total: rpcResp.Total,
	}, nil
}
