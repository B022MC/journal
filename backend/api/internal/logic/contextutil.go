package logic

import (
	"context"
	"encoding/json"
)

func currentUserID(ctx context.Context) int64 {
	if idNumber, ok := ctx.Value("userId").(json.Number); ok {
		uid, _ := idNumber.Int64()
		return uid
	}
	if idFloat, ok := ctx.Value("userId").(float64); ok {
		return int64(idFloat)
	}
	return 0
}
