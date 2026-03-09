package paperlogic

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"journal/model"
	"journal/rpc/paper/internal/svc"
	"journal/rpc/paper/paper"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubmitPaperLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSubmitPaperLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitPaperLogic {
	return &SubmitPaperLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SubmitPaperLogic) SubmitPaper(in *paper.SubmitPaperReq) (*paper.SubmitPaperResp, error) {
	p := &model.Paper{
		Title:      in.Title,
		TitleEn:    in.TitleEn,
		Abstract:   in.Abstract,
		AbstractEn: sql.NullString{String: in.AbstractEn, Valid: in.AbstractEn != ""},
		Content:    in.Content,
		AuthorId:   in.AuthorId,
		AuthorName: in.AuthorName,
		Discipline: in.Discipline,
		Zone:       "latrine",
		Keywords:   in.Keywords,
		FilePath:   in.FilePath,
		Status:     1,
	}

	id, err := l.svcCtx.PaperModel.Insert(l.ctx, p)
	if err != nil {
		return nil, err
	}

	// Generate DOI
	doi := fmt.Sprintf("10.S.H.I.T/%d.%d", time.Now().Year(), id)

	return &paper.SubmitPaperResp{
		Id:  id,
		Doi: doi,
	}, nil
}
