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

type AdminListPapersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminListPapersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListPapersLogic {
	return &AdminListPapersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminListPapersLogic) AdminListPapers(req *types.ListPapersReq) (resp *types.ListPapersResp, err error) {
	rpcResp, err := l.svcCtx.AdminRpc.ListPapers(l.ctx, &adminClient.ListPapersReq{
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
		Zone:     req.Zone,
	})
	if err != nil {
		return nil, err
	}

	resp = &types.ListPapersResp{
		Total: rpcResp.Total,
		Items: make([]types.PaperItem, 0, len(rpcResp.Items)),
	}

	for _, item := range rpcResp.Items {
		resp.Items = append(resp.Items, types.PaperItem{
			Id:               item.Id,
			Title:            item.Title,
			AuthorId:         item.AuthorId,
			AuthorName:       item.AuthorName,
			Zone:             item.Zone,
			Status:           item.Status,
			DegradationLevel: item.DegradationLevel,
			ShitScore:        item.ShitScore,
			CreatedAt:        item.CreatedAt,
		})
	}

	return resp, nil
}
