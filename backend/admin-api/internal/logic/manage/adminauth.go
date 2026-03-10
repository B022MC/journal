package manage

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"journal/admin-api/internal/svc"
	"journal/common/consts"
	"journal/rpc/admin/adminClient"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const keywordRuleManageMinContribution = 200.0

// currentUserID extracts the user ID from the JWT context.
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

// requireAdminPermission checks that the current user has the specified permission.
// Super admins always pass because IsSuperAdmin is called inside AdminRBACModel.HasPermission.
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
		return 0, status.Error(codes.PermissionDenied, "permission denied: "+permissionCode)
	}

	return userId, nil
}

// listAdminPermissionCodes returns all permission codes the user has.
// Super admins get all permission codes.
func listAdminPermissionCodes(ctx context.Context, svcCtx *svc.ServiceContext, userId int64) ([]string, error) {
	return svcCtx.AdminRBAC.ListPermissionCodesByUserId(ctx, userId)
}

// isSuperAdmin checks if the user is a super admin.
func isSuperAdmin(ctx context.Context, svcCtx *svc.ServiceContext, userId int64) (bool, error) {
	return svcCtx.AdminRBAC.IsSuperAdmin(ctx, userId)
}

func requireAdminContributionAtLeast(ctx context.Context, svcCtx *svc.ServiceContext, minimum float64) (int64, error) {
	userId, err := currentUserID(ctx)
	if err != nil {
		return 0, status.Error(codes.Unauthenticated, "unauthorized")
	}

	userInfo, err := svcCtx.UserModel.FindByIdPrimary(ctx, userId)
	if err != nil {
		return 0, status.Error(codes.PermissionDenied, "failed to verify contribution score")
	}
	if userInfo.ContributionScore < minimum {
		return 0, status.Error(codes.PermissionDenied, "contribution score is below required threshold")
	}

	return userId, nil
}

func requireKeywordRuleManageAccess(ctx context.Context, svcCtx *svc.ServiceContext) (int64, error) {
	userId, err := requireAdminPermission(ctx, svcCtx, consts.PermAdminKeywordManage)
	if err != nil {
		return 0, err
	}
	if _, err := requireAdminContributionAtLeast(ctx, svcCtx, keywordRuleManageMinContribution); err != nil {
		return 0, err
	}
	return userId, nil
}

func recordAdminAuditLog(ctx context.Context, svcCtx *svc.ServiceContext, actorUserId int64, permissionCode, action, targetType string, targetId int64, detail string) {
	_, err := svcCtx.AdminRpc.RecordAuditLog(ctx, &adminClient.RecordAuditLogReq{
		ActorUserId:    actorUserId,
		PermissionCode: permissionCode,
		Action:         action,
		TargetType:     targetType,
		TargetId:       targetId,
		Detail:         detail,
	})
	if err != nil {
		logx.WithContext(ctx).Errorf("record admin audit log failed: %v", err)
	}
}

// Permission code aliases using centralized constants.
// These are kept for backward compatibility with existing logic files.
var (
	_ = consts.PermAdminDashboardView
	_ = consts.PermAdminKeywordView
	_ = consts.PermAdminKeywordManage
	_ = consts.PermAdminNewsView
	_ = consts.PermAdminNewsCreate
	_ = consts.PermAdminNewsUpdate
	_ = consts.PermAdminNewsDelete
	_ = consts.PermAdminPaperView
	_ = consts.PermAdminPaperZoneUp
	_ = consts.PermAdminUserView
	_ = consts.PermAdminUserManage
	_ = consts.PermAdminRoleView
	_ = consts.PermAdminRoleManage
	_ = consts.PermAdminAuditView
)
