package types

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestSearchPapersReqFormContract(t *testing.T) {
	t.Parallel()

	reqType := reflect.TypeOf(SearchPapersReq{})
	cases := map[string]string{
		"Query":           `form:"query"`,
		"Discipline":      `form:"discipline,optional"`,
		"Sort":            `form:"sort,optional,default=relevance"`,
		"Engine":          `form:"engine,optional,default=auto"`,
		"ShadowCompare":   `form:"shadow_compare,optional"`,
		"SuggestionLimit": `form:"suggestion_limit,optional,default=6"`,
		"Page":            `form:"page,optional,default=1"`,
		"PageSize":        `form:"page_size,optional,default=20"`,
	}

	for fieldName, wantTag := range cases {
		field, ok := reqType.FieldByName(fieldName)
		if !ok {
			t.Fatalf("missing field %s", fieldName)
		}
		if got := string(field.Tag); got != wantTag {
			t.Fatalf("field %s tag mismatch: got %q want %q", fieldName, got, wantTag)
		}
	}
}

func TestSearchPapersRespJSONContract(t *testing.T) {
	t.Parallel()

	payload := SearchPapersResp{
		Items: []PaperItem{
			{
				Id:         1,
				Title:      "人工智能论文推荐系统",
				Discipline: "cs",
				Zone:       "stone",
			},
		},
		Total:       1,
		Suggestions: []string{"人工智能论文推荐系统"},
		Meta: SearchMeta{
			Engine:         "fulltext",
			UsedFallback:   true,
			FallbackReason: "manual_override",
			ShadowCompared: true,
			IndexedDocs:    3,
			IndexedTerms:   257,
			IndexSignature: "ac73621534970cbfed28488b073d53ec2f646ef3",
			ExpandedTerms:  []string{"人工智能", "paper"},
		},
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal contract payload: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal contract payload: %v", err)
	}

	for _, key := range []string{"items", "total", "suggestions", "meta"} {
		if _, ok := decoded[key]; !ok {
			t.Fatalf("missing top-level key %q in %s", key, string(raw))
		}
	}

	meta, ok := decoded["meta"].(map[string]any)
	if !ok {
		t.Fatalf("meta should be a JSON object, got %T", decoded["meta"])
	}
	for _, key := range []string{
		"engine",
		"used_fallback",
		"fallback_reason",
		"shadow_compared",
		"indexed_docs",
		"indexed_terms",
		"index_signature",
		"expanded_terms",
	} {
		if _, ok := meta[key]; !ok {
			t.Fatalf("missing meta key %q in %s", key, string(raw))
		}
	}
}
