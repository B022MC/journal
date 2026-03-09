package logic

import (
	"context"

	"journal/rpc/admin/admin"
	"journal/rpc/admin/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListFlagsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListFlagsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFlagsLogic {
	return &ListFlagsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Flag
func (l *ListFlagsLogic) ListFlags(in *admin.ListFlagsReq) (*admin.ListFlagsResp, error) {
	flags, total, err := l.svcCtx.FlagModel.ListFlagsPaginated(l.ctx, int(in.Page), int(in.PageSize), in.Status)
	if err != nil {
		return nil, err
	}

	var items []*admin.FlagItem
	for _, f := range flags {
		items = append(items, &admin.FlagItem{
			Id:         f.Id,
			TargetType: f.TargetType,
			TargetId:   f.TargetId,
			ReporterId: f.ReporterId,
			Reason:     f.Reason,
			Detail:     f.Detail,
			Status:     f.Status,
			CreatedAt:  f.CreatedAt.Unix(),
		})
	}

	return &admin.ListFlagsResp{
		Items: items,
		Total: total,
	}, nil
}
