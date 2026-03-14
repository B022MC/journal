package logic

import (
	"context"
	"encoding/json"
	"testing"
)

func TestCurrentUserIDSupportsCommonValueTypes(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  int64
	}{
		{name: "json number", value: json.Number("42"), want: 42},
		{name: "int64", value: int64(43), want: 43},
		{name: "int", value: int(44), want: 44},
		{name: "float64", value: float64(45), want: 45},
		{name: "missing", value: nil, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.value != nil {
				ctx = context.WithValue(ctx, "userId", tt.value)
			}
			if got := currentUserID(ctx); got != tt.want {
				t.Fatalf("currentUserID() = %d, want %d", got, tt.want)
			}
		})
	}
}
