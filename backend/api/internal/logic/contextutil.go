package logic

import (
	"context"
	"encoding/json"
	"strconv"
)

func currentUserID(ctx context.Context) int64 {
	for _, key := range []string{"userId", "user_id"} {
		if uid, ok := parseContextUserID(ctx.Value(key)); ok {
			return uid
		}
	}
	return 0
}

func parseContextUserID(value any) (int64, bool) {
	if idNumber, ok := value.(json.Number); ok {
		uid, _ := idNumber.Int64()
		return uid, true
	}
	if idInt64, ok := value.(int64); ok {
		return idInt64, true
	}
	if idInt, ok := value.(int); ok {
		return int64(idInt), true
	}
	if idFloat, ok := value.(float64); ok {
		return int64(idFloat), true
	}
	if idString, ok := value.(string); ok {
		uid, err := strconv.ParseInt(idString, 10, 64)
		return uid, err == nil
	}
	return 0, false
}
