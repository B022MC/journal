package logic

import (
	"context"

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
	userId := currentUserID(l.ctx)
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

	if cacheErr := l.svcCtx.Cache.InvalidateHotPapers(l.ctx, "latrine"); cacheErr != nil {
		l.Errorf("failed to invalidate hot papers cache after submit paper=%d: %v", rpcResp.Id, cacheErr)
	}
	if cacheErr := l.svcCtx.Cache.DeletePaperModeration(l.ctx, rpcResp.Id); cacheErr != nil {
		l.Errorf("failed to invalidate paper moderation cache after submit paper=%d: %v", rpcResp.Id, cacheErr)
	}
	return &types.SubmitPaperResp{Id: rpcResp.Id, Doi: rpcResp.Doi}, nil
}
