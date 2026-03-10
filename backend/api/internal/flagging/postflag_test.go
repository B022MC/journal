package flagging

import "testing"

func TestPostFlagEventValidate(t *testing.T) {
	t.Parallel()

	validCases := []PostFlagEvent{
		{TargetType: "paper", TargetId: 1},
		{TargetType: "rating", TargetId: 2},
	}
	for _, tc := range validCases {
		if err := tc.Validate(); err != nil {
			t.Fatalf("Validate() returned error for valid event %+v: %v", tc, err)
		}
	}

	invalidCases := []PostFlagEvent{
		{TargetType: "paper", TargetId: 0},
		{TargetType: "user", TargetId: 1},
	}
	for _, tc := range invalidCases {
		if err := tc.Validate(); err == nil {
			t.Fatalf("Validate() returned nil for invalid event %+v", tc)
		}
	}
}

func TestDecodePostFlagEvent(t *testing.T) {
	t.Parallel()

	event, err := decodePostFlagEvent(`{"flag_id":7,"target_type":"paper","target_id":11,"reporter_id":13,"attempt":1}`)
	if err != nil {
		t.Fatalf("decodePostFlagEvent() returned error: %v", err)
	}

	if event.FlagId != 7 || event.TargetType != "paper" || event.TargetId != 11 || event.ReporterId != 13 || event.Attempt != 1 {
		t.Fatalf("decodePostFlagEvent() = %+v", event)
	}
}
