package manage

import (
	"context"
	"encoding/json"
	"testing"
)

func TestCurrentUserIDSupportsLegacyAndSnakeCaseClaims(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want int64
		val  any
	}{
		{name: "legacy json number", key: "userId", val: json.Number("7"), want: 7},
		{name: "legacy string", key: "userId", val: "8", want: 8},
		{name: "snake case int64", key: "user_id", val: int64(9), want: 9},
		{name: "snake case int", key: "user_id", val: int(10), want: 10},
		{name: "snake case float64", key: "user_id", val: float64(11), want: 11},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), tt.key, tt.val)

			got, err := currentUserID(ctx)
			if err != nil {
				t.Fatalf("currentUserID returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("currentUserID = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCurrentUserIDReturnsErrorWithoutSupportedClaim(t *testing.T) {
	if _, err := currentUserID(context.Background()); err == nil {
		t.Fatal("currentUserID should fail when no supported claim is present")
	}
}
