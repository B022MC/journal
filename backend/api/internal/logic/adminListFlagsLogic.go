// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"journal/api/internal/svc"
	"journal/api/internal/types"
	"journal/rpc/admin/adminClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminListFlagsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminListFlagsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListFlagsLogic {
	return &AdminListFlagsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminListFlagsLogic) AdminListFlags(req *types.ListFlagsReq) (resp *types.ListFlagsResp, err error) {
	if _, err := requireAdminPermission(l.ctx, l.svcCtx, permissionAdminPaperView); err != nil {
		return nil, err
	}

	rpcResp, err := l.svcCtx.AdminRpc.ListFlags(l.ctx, &adminClient.ListFlagsReq{
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
		Status:   req.Status,
	})
	if err != nil {
		return nil, err
	}

	resp = &types.ListFlagsResp{
		Total: rpcResp.Total,
		Items: make([]types.FlagItem, 0, len(rpcResp.Items)),
	}

	for _, item := range rpcResp.Items {
		resp.Items = append(resp.Items, types.FlagItem{
			Id:         item.Id,
			TargetType: item.TargetType,
			TargetId:   item.TargetId,
			ReporterId: item.ReporterId,
			Reason:     item.Reason,
			Detail:     item.Detail,
			Status:     item.Status,
			CreatedAt:  item.CreatedAt,
		})
	}

	return resp, nil
}
