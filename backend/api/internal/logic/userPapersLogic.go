package logic

import (
	"context"
	"encoding/json"

	"journal/api/internal/svc"
	"journal/api/internal/types"
	"journal/rpc/paper/paper"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserPapersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserPapersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserPapersLogic {
	return &UserPapersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserPapersLogic) UserPapers(req *types.PageReq) (resp *types.ListPapersResp, err error) {
	userId, _ := l.ctx.Value("userId").(json.Number).Int64()
	rpcResp, err := l.svcCtx.PaperRpc.UserPapers(l.ctx, &paper.UserPapersReq{
		AuthorId: userId,
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
	})
	if err != nil {
		return nil, err
	}
	return &types.ListPapersResp{
		Items: toPaperItems(rpcResp.Items),
		Total: rpcResp.Total,
	}, nil
}
