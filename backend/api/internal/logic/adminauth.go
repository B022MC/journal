package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"journal/api/internal/svc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	permissionAdminNewsCreate      = "admin.news.create"
	permissionAdminPaperZoneUpdate = "admin.paper.zone.update"
)

func currentUserID(ctx context.Context) (int64, error) {
	switch v := ctx.Value("userId").(type) {
	case json.Number:
		return v.Int64()
	case string:
		return strconv.ParseInt(v, 10, 64)
	case int64:
		return v, nil
	default:
		return 0, fmt.Errorf("userId not found in context")
	}
}

func listAdminPermissionCodes(ctx context.Context, svcCtx *svc.ServiceContext, userId int64) ([]string, error) {
	return svcCtx.AdminRBAC.ListPermissionCodesByUserId(ctx, userId)
}

func requireAdminPermission(ctx context.Context, svcCtx *svc.ServiceContext, permissionCode string) (int64, error) {
	userId, err := currentUserID(ctx)
	if err != nil {
		return 0, status.Error(codes.Unauthenticated, "unauthorized")
	}

	allowed, err := svcCtx.AdminRBAC.HasPermission(ctx, userId, permissionCode)
	if err != nil {
		return 0, status.Error(codes.Internal, "failed to verify permission")
	}
	if !allowed {
		return 0, status.Error(codes.PermissionDenied, "permission denied")
	}

	return userId, nil
}
