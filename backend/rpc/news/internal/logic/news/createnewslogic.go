package newslogic

import (
	"context"
	"database/sql"

	"journal/model"
	"journal/rpc/news/internal/svc"
	"journal/rpc/news/news"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateNewsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateNewsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateNewsLogic {
	return &CreateNewsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateNewsLogic) CreateNews(in *news.CreateNewsReq) (*news.CommonResp, error) {
	isPinned := int32(0)
	if in.IsPinned {
		isPinned = 1
	}

	_, err := l.svcCtx.NewsModel.Insert(l.ctx, &model.News{
		Title:     in.Title,
		TitleEn:   in.TitleEn,
		Content:   in.Content,
		ContentEn: sql.NullString{String: in.ContentEn, Valid: in.ContentEn != ""},
		AuthorId:  in.AuthorId,
		Category:  in.Category,
		IsPinned:  isPinned,
		Status:    1,
	})
	if err != nil {
		return nil, err
	}

	return &news.CommonResp{
		Success: true,
		Message: "news created",
	}, nil
}
