package logic

import (
	"context"
	"encoding/json"
	"strings"

	"journal/api/internal/svc"
	"journal/api/internal/types"
	"journal/common/degradation"
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

	sourceIP := strings.TrimSpace(contextString(l.ctx, "requestIP"))
	userAgent := degradation.NormalizeRatingUserAgent(contextString(l.ctx, "requestUserAgent"))
	deviceFingerprint := degradation.BuildRatingDeviceFingerprint(sourceIP, userAgent)
	if sourceIP != "" || userAgent != "" {
		if err := l.svcCtx.RatingModel.UpdateRequestFingerprint(l.ctx, paperId, userId, sourceIP, userAgent, deviceFingerprint); err != nil {
			l.Errorf("failed to persist rating fingerprint for paper=%d user=%d: %v", paperId, userId, err)
		} else if err := l.svcCtx.RatingGuard.InspectIdentityCluster(l.ctx, paperId, sourceIP, userAgent); err != nil {
			l.Errorf("failed to inspect rating identity cluster for paper=%d user=%d: %v", paperId, userId, err)
		}
	}

	return &types.CommonResp{Success: true, Message: "rating submitted"}, nil
}

func contextString(ctx context.Context, key string) string {
	value, _ := ctx.Value(key).(string)
	return value
}
