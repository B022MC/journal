package paperlogic

import (
	"context"
	"fmt"

	"journal/rpc/paper/internal/svc"
	"journal/rpc/paper/paper"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateZoneLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateZoneLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateZoneLogic {
	return &UpdateZoneLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateZoneLogic) UpdateZone(in *paper.UpdateZoneReq) (*paper.CommonResp, error) {
	validZones := map[string]bool{
		"latrine": true, "septic_tank": true, "stone": true, "sediment": true,
	}
	if !validZones[in.Zone] {
		return nil, fmt.Errorf("invalid zone: %s", in.Zone)
	}

	err := l.svcCtx.PaperModel.UpdateZone(l.ctx, in.Id, in.Zone)
	if err != nil {
		return nil, err
	}

	return &paper.CommonResp{
		Success: true,
		Message: "zone updated to " + in.Zone,
	}, nil
}
