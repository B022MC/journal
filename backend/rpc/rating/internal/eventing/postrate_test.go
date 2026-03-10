package eventing

import "testing"

func TestPostRateEventValidate(t *testing.T) {
	valid := PostRateEvent{
		PaperId:    1,
		ReviewerId: 2,
		AuthorId:   3,
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() returned error for valid event: %v", err)
	}

	if err := (PostRateEvent{ReviewerId: 2}).Validate(); err == nil {
		t.Fatalf("Validate() expected error for missing paper id")
	}

	if err := (PostRateEvent{PaperId: 1}).Validate(); err == nil {
		t.Fatalf("Validate() expected error for missing reviewer id")
	}
}

func TestDecodePostRateEvent(t *testing.T) {
	event, err := decodePostRateEvent(`{"paper_id":11,"reviewer_id":22,"author_id":33,"attempt":1}`)
	if err != nil {
		t.Fatalf("decodePostRateEvent() error = %v", err)
	}
	if event.PaperId != 11 || event.ReviewerId != 22 || event.AuthorId != 33 || event.Attempt != 1 {
		t.Fatalf("decodePostRateEvent() = %+v", event)
	}

	if _, err := decodePostRateEvent(`{"paper_id":0,"reviewer_id":22}`); err == nil {
		t.Fatalf("decodePostRateEvent() expected validation error")
	}
}
