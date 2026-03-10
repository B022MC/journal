package paperlogic

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"journal/common/consts"
	"journal/common/degradation"
	"journal/model"
	"journal/rpc/paper/internal/svc"
	"journal/rpc/paper/paper"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	content := strings.Join([]string{
		in.Title,
		in.TitleEn,
		in.Abstract,
		in.AbstractEn,
		in.Keywords,
		in.Content,
	}, "\n")

	match, err := l.svcCtx.KeywordFilter.Check(l.ctx, content)
	if err != nil {
		return nil, err
	}
	if match != nil {
		return nil, status.Errorf(codes.InvalidArgument, "paper content blocked by keyword blacklist category=%s match_type=%s", match.Category, match.MatchType)
	}

	simhash := degradation.ComputeSimHash(content)
	similarPapers, err := l.svcCtx.PaperModel.FindSimilarBySimhash(l.ctx, simhash, 3, 5)
	if err != nil {
		return nil, err
	}

	paperStatus := int32(consts.PaperStatusActive)
	if len(similarPapers) > 0 {
		paperStatus = consts.PaperStatusFlagged
	}

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
		Simhash:    simhash,
		Status:     paperStatus,
	}

	id, err := l.svcCtx.PaperModel.Insert(l.ctx, p)
	if err != nil {
		return nil, err
	}

	if err := l.svcCtx.UserModel.UpdateLastActive(l.ctx, in.AuthorId); err != nil {
		return nil, err
	}
	if err := l.svcCtx.AchievementService.SyncUser(l.ctx, in.AuthorId); err != nil {
		return nil, err
	}

	// Generate DOI
	doi := fmt.Sprintf("10.S.H.I.T/%d.%d", time.Now().Year(), id)
	if err := l.svcCtx.PaperModel.UpdateDoi(l.ctx, id, doi); err != nil {
		return nil, err
	}

	if len(similarPapers) > 0 {
		flag := &model.Flag{
			TargetType:           consts.FlagTargetPaper,
			TargetId:             id,
			ReporterId:           consts.SystemReporterSimhash,
			Reason:               consts.FlagReasonPlagiarism,
			Detail:               buildSimhashFlagDetail(similarPapers),
			ReporterContribution: 0,
		}
		if _, _, flagErr := l.svcCtx.DegradationEngine.ProcessFlag(l.ctx, flag); flagErr != nil {
			l.Errorf("submit paper %d simhash auto-flag failed: %v", id, flagErr)
		}
	}

	return &paper.SubmitPaperResp{
		Id:  id,
		Doi: doi,
	}, nil
}

func buildSimhashFlagDetail(items []*model.PaperSimilarity) string {
	parts := make([]string, 0, len(items))
	for _, item := range items {
		parts = append(parts, fmt.Sprintf("paper#%d(distance=%d)", item.Id, item.Distance))
	}
	return "simhash suspected duplicate: " + strings.Join(parts, ", ")
}
