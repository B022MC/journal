package logic

import (
	"context"
	"encoding/json"

	"journal/api/internal/svc"
	"journal/api/internal/types"
	"journal/rpc/rating/rating"

	"github.com/zeromicro/go-zero/core/logx"
)

type RatePaperLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRatePaperLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RatePaperLogic {
	return &RatePaperLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RatePaperLogic) RatePaper(req *types.RatePaperReq, paperId int64) (resp *types.CommonResp, err error) {
	userId, _ := l.ctx.Value("userId").(json.Number).Int64()
	_, err = l.svcCtx.RatingRpc.RatePaper(l.ctx, &rating.RatePaperReq{
		PaperId: paperId,
		UserId:  userId,
		Score:   req.Score,
		Comment: req.Comment,
	})
	if err != nil {
		return nil, err
	}
	return &types.CommonResp{Success: true, Message: "rating submitted"}, nil
}
