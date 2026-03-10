package degradation

import "testing"

func TestComputeSimHashIdenticalTextsMatchExactly(t *testing.T) {
	text := "量子计算 论文 评审 系统 quantum review system"
	left := ComputeSimHash(text)
	right := ComputeSimHash(text)

	if HammingDistance(left, right) != 0 {
		t.Fatalf("expected identical texts to have zero hamming distance")
	}
}

func TestComputeSimHashKeepsNearDuplicateDistanceLow(t *testing.T) {
	base := "这篇论文讨论量子计算在匿名评审系统中的应用，并分析评审分布与争议度。"
	edited := "这篇论文讨论量子计算在匿名评审系统中的应用，分析评审分布、争议度，以及投稿治理机制。"
	different := "足球联赛的主客场积分模型与门将扑救概率统计。"

	baseHash := ComputeSimHash(base)
	editedHash := ComputeSimHash(edited)
	differentHash := ComputeSimHash(different)

	nearDistance := HammingDistance(baseHash, editedHash)
	farDistance := HammingDistance(baseHash, differentHash)

	if nearDistance > 10 {
		t.Fatalf("expected near duplicate distance <= 10, got %d", nearDistance)
	}
	if farDistance <= nearDistance {
		t.Fatalf("expected unrelated text distance > near duplicate distance, got near=%d far=%d", nearDistance, farDistance)
	}
}
