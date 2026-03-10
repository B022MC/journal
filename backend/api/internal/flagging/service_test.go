package flagging

import "testing"

func TestEffectivePaperDegradationLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		currentLevel int32
		pendingCount int
		weightedSum  float64
		quorum       int
		want         int32
	}{
		{
			name:         "uses pending flags before async consumer persists paper level",
			currentLevel: 0,
			pendingCount: 3,
			weightedSum:  12,
			quorum:       3,
			want:         2,
		},
		{
			name:         "keeps persisted level after pending flags are resolved",
			currentLevel: 2,
			pendingCount: 0,
			weightedSum:  0,
			quorum:       3,
			want:         2,
		},
		{
			name:         "keeps higher persisted seal over lower pending level",
			currentLevel: 3,
			pendingCount: 2,
			weightedSum:  12,
			quorum:       4,
			want:         3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := effectivePaperDegradationLevel(tt.currentLevel, tt.pendingCount, tt.weightedSum, tt.quorum)
			if got != tt.want {
				t.Fatalf("effectivePaperDegradationLevel() = %d, want %d", got, tt.want)
			}
		})
	}
}
