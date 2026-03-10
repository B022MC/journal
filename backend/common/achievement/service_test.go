package achievement

import "testing"

func TestEligibleAchievementCodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		metrics *metrics
		want    []string
	}{
		{
			name:    "no milestones",
			metrics: &metrics{},
			want:    []string{},
		},
		{
			name: "first submission only",
			metrics: &metrics{
				PaperCount: 1,
			},
			want: []string{"first_submission"},
		},
		{
			name: "submission and sediment",
			metrics: &metrics{
				PaperCount:         2,
				SedimentPaperCount: 1,
			},
			want: []string{"first_submission", "sediment_breakthrough"},
		},
		{
			name: "all milestones",
			metrics: &metrics{
				PaperCount:         3,
				SedimentPaperCount: 1,
				ReviewCount:        100,
			},
			want: []string{"first_submission", "sediment_breakthrough", "reviewer_century"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := eligibleAchievementCodes(tt.metrics)
			if len(got) != len(tt.want) {
				t.Fatalf("eligibleAchievementCodes() len = %d, want %d (%v)", len(got), len(tt.want), got)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("eligibleAchievementCodes()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
