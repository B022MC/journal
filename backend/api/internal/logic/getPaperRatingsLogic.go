package logic

import (
	"context"

	"journal/api/internal/svc"
	"journal/api/internal/types"
	"journal/rpc/rating/rating"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPaperRatingsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPaperRatingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPaperRatingsLogic {
	return &GetPaperRatingsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPaperRatingsLogic) GetPaperRatings(paperId int64, req *types.PageReq) (resp *types.PaperRatingsResp, err error) {
	rpcResp, err := l.svcCtx.RatingRpc.GetPaperRatings(l.ctx, &rating.PaperRatingsReq{
		PaperId:  paperId,
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
	})
	if err != nil {
		return nil, err
	}

	items := make([]types.RatingItem, 0, len(rpcResp.Items))
	for _, r := range rpcResp.Items {
		items = append(items, types.RatingItem{
			Id:        r.Id,
			PaperId:   r.PaperId,
			UserId:    r.UserId,
			Username:  r.Username,
			Nickname:  r.Nickname,
			Score:     r.Score,
			Comment:   r.Comment,
			CreatedAt: r.CreatedAt,
		})
	}

	return &types.PaperRatingsResp{
		Items:    items,
		Total:    rpcResp.Total,
		AvgScore: rpcResp.AvgScore,
	}, nil
}
