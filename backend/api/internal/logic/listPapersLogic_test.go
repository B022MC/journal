package logic

import (
	"testing"

	"journal/api/internal/types"
)

func TestShouldUseHotPapersCache(t *testing.T) {
	tests := []struct {
		name string
		req  *types.ListPapersReq
		want bool
	}{
		{
			name: "highest rated first page",
			req: &types.ListPapersReq{
				Sort:     "highest_rated",
				Page:     1,
				PageSize: 20,
			},
			want: true,
		},
		{
			name: "with discipline filter bypasses cache",
			req: &types.ListPapersReq{
				Sort:       "highest_rated",
				Discipline: "cs",
				Page:       1,
				PageSize:   20,
			},
			want: false,
		},
		{
			name: "page beyond top 100 bypasses cache",
			req: &types.ListPapersReq{
				Sort:     "highest_rated",
				Page:     6,
				PageSize: 20,
			},
			want: false,
		},
		{
			name: "request spanning beyond top 100 bypasses cache",
			req: &types.ListPapersReq{
				Sort:     "highest_rated",
				Page:     4,
				PageSize: 30,
			},
			want: false,
		},
		{
			name: "newest sort bypasses cache",
			req: &types.ListPapersReq{
				Sort:     "newest",
				Page:     1,
				PageSize: 20,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		page, pageSize := normalizePagination(tt.req.Page, tt.req.PageSize)
		if got := shouldUseHotPapersCache(tt.req, page, pageSize); got != tt.want {
			t.Fatalf("%s: shouldUseHotPapersCache() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestSliceHotPapers(t *testing.T) {
	payload := &types.ListPapersResp{
		Items: []types.PaperItem{
			{Id: 1},
			{Id: 2},
			{Id: 3},
			{Id: 4},
			{Id: 5},
		},
		Total: 42,
	}

	resp := sliceHotPapers(payload, 2, 2)
	if resp.Total != 42 {
		t.Fatalf("sliceHotPapers() total = %d, want 42", resp.Total)
	}
	if len(resp.Items) != 2 || resp.Items[0].Id != 3 || resp.Items[1].Id != 4 {
		t.Fatalf("sliceHotPapers() items = %+v, want ids [3 4]", resp.Items)
	}

	empty := sliceHotPapers(payload, 4, 2)
	if len(empty.Items) != 0 {
		t.Fatalf("sliceHotPapers() out of range items = %+v, want empty", empty.Items)
	}
}
