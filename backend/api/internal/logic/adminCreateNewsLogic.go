package logic

import (
	"context"

	"journal/api/internal/svc"
	"journal/api/internal/types"
	"journal/rpc/news/news"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminCreateNewsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminCreateNewsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminCreateNewsLogic {
	return &AdminCreateNewsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminCreateNewsLogic) AdminCreateNews(req *types.CreateNewsReq) (resp *types.CommonResp, err error) {
	authorId, err := requireAdminPermission(l.ctx, l.svcCtx, permissionAdminNewsCreate)
	if err != nil {
		return nil, err
	}
	_, err = l.svcCtx.NewsRpc.CreateNews(l.ctx, &news.CreateNewsReq{
		Title:     req.Title,
		TitleEn:   req.TitleEn,
		Content:   req.Content,
		ContentEn: req.ContentEn,
		AuthorId:  authorId,
		Category:  req.Category,
		IsPinned:  req.IsPinned,
	})
	if err != nil {
		return nil, err
	}
	return &types.CommonResp{Success: true, Message: "news created"}, nil
}
