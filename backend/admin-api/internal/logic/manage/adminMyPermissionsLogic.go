package manage

import (
	"context"

	"journal/admin-api/internal/svc"
	"journal/admin-api/internal/types"
	"journal/common/consts"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminMyPermissionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminMyPermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminMyPermissionsLogic {
	return &AdminMyPermissionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminMyPermissionsLogic) AdminMyPermissions() (resp *types.MyPermissionsResp, err error) {
	userId, err := currentUserID(l.ctx)
	if err != nil {
		return nil, err
	}

	isSuper, err := isSuperAdmin(l.ctx, l.svcCtx, userId)
	if err != nil {
		return nil, err
	}

	codes, err := listAdminPermissionCodes(l.ctx, l.svcCtx, userId)
	if err != nil {
		return nil, err
	}

	_ = consts.PermAdminDashboardView // ensure import

	return &types.MyPermissionsResp{
		Permissions: codes,
		IsSuper:     isSuper,
	}, nil
}
