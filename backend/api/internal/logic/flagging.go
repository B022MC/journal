package logic

import (
	"context"
	"errors"

	"journal/api/internal/flagging"
	"journal/api/internal/svc"
	"journal/api/internal/types"
	"journal/common/errorx"
)

func submitFlag(ctx context.Context, svcCtx *svc.ServiceContext, targetType string, targetId int64, req *types.FlagReq) (*types.FlagActionResp, error) {
	userId := currentUserID(ctx)
	flagId, status, err := svcCtx.FlagService.SubmitFlag(ctx, targetType, targetId, userId, req.Reason, req.Detail)
	if err != nil {
		resp := &types.FlagActionResp{
			Success: false,
			Message: errorx.CodeMsg(errorx.ErrInternal),
			Status: types.FlagStatusResp{
				TargetType: targetType,
				TargetId:   targetId,
			},
		}

		switch {
		case errors.Is(err, flagging.ErrTargetNotFound):
			resp.Message = errorx.CodeMsg(errorx.ErrNotFound)
			resp.Status.Exists = false
			return resp, nil
		case errors.Is(err, flagging.ErrAlreadyFlagged):
			resp.Message = errorx.CodeMsg(errorx.ErrAlreadyFlagged)
			if currentStatus, statusErr := svcCtx.FlagService.GetFlagStatus(ctx, targetType, targetId); statusErr == nil {
				resp.Status = toFlagStatusResp(currentStatus)
			} else {
				resp.Status.Exists = true
			}
			return resp, nil
		case errors.Is(err, flagging.ErrSelfFlag):
			resp.Message = errorx.CodeMsg(errorx.ErrSelfFlag)
			resp.Status.Exists = true
			return resp, nil
		case errors.Is(err, flagging.ErrInvalidReason):
			resp.Message = errorx.CodeMsg(errorx.ErrInvalidParam)
			resp.Status.Exists = true
			return resp, nil
		default:
			return nil, err
		}
	}

	resp := &types.FlagActionResp{
		Success: true,
		Message: "flag submitted",
		FlagId:  flagId,
		Status:  toFlagStatusResp(status),
	}
	return resp, nil
}

func getFlagStatus(ctx context.Context, svcCtx *svc.ServiceContext, targetType string, targetId int64) (*types.FlagStatusResp, error) {
	status, err := svcCtx.FlagService.GetFlagStatus(ctx, targetType, targetId)
	if err != nil {
		if errors.Is(err, flagging.ErrTargetNotFound) {
			return &types.FlagStatusResp{
				Exists:     false,
				TargetType: targetType,
				TargetId:   targetId,
			}, nil
		}
		return nil, err
	}
	resp := toFlagStatusResp(status)
	return &resp, nil
}

func toFlagStatusResp(status *flagging.TargetStatus) types.FlagStatusResp {
	if status == nil {
		return types.FlagStatusResp{}
	}

	return types.FlagStatusResp{
		Exists:           status.Exists,
		TargetType:       status.TargetType,
		TargetId:         status.TargetId,
		FlagCount:        status.FlagCount,
		PendingCount:     status.PendingCount,
		WeightedSum:      status.WeightedSum,
		Quorum:           status.Quorum,
		DegradationLevel: status.DegradationLevel,
	}
}
