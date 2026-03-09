package logic

import (
	"context"

	"journal/api/internal/svc"
	"journal/api/internal/types"
	"journal/rpc/paper/paper"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminUpdateZoneLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminUpdateZoneLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUpdateZoneLogic {
	return &AdminUpdateZoneLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminUpdateZoneLogic) AdminUpdateZone(paperId int64, req *types.UpdateZoneReq) (resp *types.CommonResp, err error) {
	_, err = l.svcCtx.PaperRpc.UpdateZone(l.ctx, &paper.UpdateZoneReq{
		Id:   paperId,
		Zone: req.Zone,
	})
	if err != nil {
		return nil, err
	}
	return &types.CommonResp{Success: true, Message: "zone updated"}, nil
}
