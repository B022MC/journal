package logic

import (
	"context"

	"journal/api/internal/svc"
	"journal/api/internal/types"
	"journal/rpc/paper/paper"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPublicUserPapersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPublicUserPapersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPublicUserPapersLogic {
	return &GetPublicUserPapersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPublicUserPapersLogic) GetPublicUserPapers(req *types.UserPageReq) (resp *types.ListPapersResp, err error) {
	page, pageSize := normalizePagination(req.Page, req.PageSize)

	rpcResp, err := l.svcCtx.PaperRpc.UserPapers(l.ctx, &paper.UserPapersReq{
		AuthorId: req.Id,
		Page:     int32(page),
		PageSize: int32(pageSize),
	})
	if err != nil {
		return nil, err
	}

	return &types.ListPapersResp{
		Items: toPaperItems(rpcResp.Items),
		Total: rpcResp.Total,
	}, nil
}
