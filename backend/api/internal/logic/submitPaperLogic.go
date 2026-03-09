package logic

import (
	"context"
	"encoding/json"

	"journal/api/internal/svc"
	"journal/api/internal/types"
	"journal/rpc/paper/paper"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubmitPaperLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSubmitPaperLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitPaperLogic {
	return &SubmitPaperLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubmitPaperLogic) SubmitPaper(req *types.SubmitPaperReq) (resp *types.SubmitPaperResp, err error) {
	userId, _ := l.ctx.Value("userId").(json.Number).Int64()
	username, _ := l.ctx.Value("username").(string)

	rpcResp, err := l.svcCtx.PaperRpc.SubmitPaper(l.ctx, &paper.SubmitPaperReq{
		Title:      req.Title,
		TitleEn:    req.TitleEn,
		Abstract:   req.Abstract,
		AbstractEn: req.AbstractEn,
		Content:    req.Content,
		AuthorId:   userId,
		AuthorName: username,
		Discipline: req.Discipline,
		Keywords:   req.Keywords,
	})
	if err != nil {
		return nil, err
	}
	return &types.SubmitPaperResp{Id: rpcResp.Id, Doi: rpcResp.Doi}, nil
}
