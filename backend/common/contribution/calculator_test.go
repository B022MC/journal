package contribution

import "testing"

func TestReviewerWeightForContribution(t *testing.T) {
	testCases := []struct {
		name  string
		score float64
		want  float64
	}{
		{name: "zero score", score: 0, want: 0.10},
		{name: "small score", score: 10, want: 0.15},
		{name: "mid score", score: 50, want: 0.35},
		{name: "cap score", score: 500, want: 1.00},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ReviewerWeightForContribution(tc.score); got != tc.want {
				t.Fatalf("ReviewerWeightForContribution(%v) = %v, want %v", tc.score, got, tc.want)
			}
		})
	}
}
