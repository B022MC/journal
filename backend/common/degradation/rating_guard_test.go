package degradation

import "testing"

func TestBimodalityCoefficientDetectsBimodalShape(t *testing.T) {
	bimodal := [10]int32{0, 0, 4, 1, 0, 0, 1, 4, 0, 0}
	unimodal := [10]int32{0, 1, 2, 3, 4, 5, 4, 3, 2, 1}

	bimodalScore := BimodalityCoefficient(bimodal)
	unimodalScore := BimodalityCoefficient(unimodal)

	if bimodalScore <= bimodalityThreshold {
		t.Fatalf("expected bimodal distribution coefficient > %.2f, got %.4f", bimodalityThreshold, bimodalScore)
	}
	if unimodalScore >= bimodalScore {
		t.Fatalf("expected unimodal coefficient < bimodal coefficient, got unimodal=%.4f bimodal=%.4f", unimodalScore, bimodalScore)
	}
}

func TestBimodalityCoefficientReturnsZeroForSmallSample(t *testing.T) {
	small := [10]int32{0, 0, 1, 1, 1, 1, 1, 0, 0, 0}
	if got := BimodalityCoefficient(small); got != 0 {
		t.Fatalf("expected small sample coefficient to be 0, got %.4f", got)
	}
}

func TestNormalizeRatingUserAgentCollapsesWhitespaceAndTrims(t *testing.T) {
	raw := "  Mozilla/5.0   (Macintosh; Intel Mac OS X 10_15_7) \n AppleWebKit/537.36  "
	got := NormalizeRatingUserAgent(raw)
	want := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36"
	if got != want {
		t.Fatalf("expected normalized user-agent %q, got %q", want, got)
	}
}

func TestBuildRatingDeviceFingerprintRequiresIPAndUserAgent(t *testing.T) {
	if got := BuildRatingDeviceFingerprint("", "Mozilla/5.0"); got != "" {
		t.Fatalf("expected empty fingerprint without source ip, got %q", got)
	}
	if got := BuildRatingDeviceFingerprint("1.2.3.4", ""); got != "" {
		t.Fatalf("expected empty fingerprint without user-agent, got %q", got)
	}
}

func TestBuildRatingDeviceFingerprintStableAcrossWhitespace(t *testing.T) {
	first := BuildRatingDeviceFingerprint("1.2.3.4", "Mozilla/5.0   Test")
	second := BuildRatingDeviceFingerprint("1.2.3.4", " Mozilla/5.0 Test ")
	if first == "" || second == "" {
		t.Fatal("expected non-empty fingerprints")
	}
	if first != second {
		t.Fatalf("expected stable fingerprint, got %q and %q", first, second)
	}
}
