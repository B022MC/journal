package newslogic

import (
	"context"
	"database/sql"
	"strings"

	"journal/model"
	"journal/rpc/news/internal/svc"
	"journal/rpc/news/news"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	content := strings.Join([]string{
		in.Title,
		in.TitleEn,
		in.Content,
		in.ContentEn,
	}, "\n")

	match, err := l.svcCtx.KeywordFilter.Check(l.ctx, content)
	if err != nil {
		return nil, err
	}
	if match != nil {
		return nil, status.Errorf(codes.InvalidArgument, "news content blocked by keyword blacklist category=%s match_type=%s", match.Category, match.MatchType)
	}

	isPinned := int32(0)
	if in.IsPinned {
		isPinned = 1
	}

	_, err = l.svcCtx.NewsModel.Insert(l.ctx, &model.News{
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
