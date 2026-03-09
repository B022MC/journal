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

type AdminListUsersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminListUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListUsersLogic {
	return &AdminListUsersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminListUsersLogic) AdminListUsers(req *types.PageReq) (resp *types.ListUsersRespAdmin, err error) {
	rpcResp, err := l.svcCtx.AdminRpc.ListUsers(l.ctx, &adminClient.ListUsersReq{
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
	})
	if err != nil {
		return nil, err
	}

	resp = &types.ListUsersRespAdmin{
		Total: rpcResp.Total,
		Items: make([]types.UserItemAdmin, 0, len(rpcResp.Items)),
	}

	for _, item := range rpcResp.Items {
		resp.Items = append(resp.Items, types.UserItemAdmin{
			Id:                item.Id,
			Username:          item.Username,
			Email:             item.Email,
			Nickname:          item.Nickname,
			Role:              item.Role,
			ContributionScore: item.ContributionScore,
			Status:            item.Status,
			CreatedAt:         item.CreatedAt,
		})
	}

	return resp, nil
}
