package search

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"journal/model"
)

func TestBuildSnapshotStableAcrossRuns(t *testing.T) {
	cfg := Config{
		DefaultEngine: EngineHybrid,
		BatchOne: BatchOneConfig{
			Enabled:     true,
			WorkerCount: 3,
			Explain:     true,
			EnableIK:    true,
			EnableJieba: true,
		},
		BatchTwo: BatchTwoConfig{
			TrieEnabled:           true,
			SynonymEnabled:        true,
			FusionEnabled:         true,
			FusionBM25Weight:      0.75,
			FusionFreshnessWeight: 0.15,
			FusionQualityWeight:   0.10,
		},
	}.Normalized()

	first, err := BuildSnapshot(context.Background(), sampleDocs(), cfg, builtInLexicon, builtInSynonyms)
	if err != nil {
		t.Fatalf("BuildSnapshot first failed: %v", err)
	}
	second, err := BuildSnapshot(context.Background(), sampleDocs(), cfg, builtInLexicon, builtInSynonyms)
	if err != nil {
		t.Fatalf("BuildSnapshot second failed: %v", err)
	}

	if first.Metadata().Signature != second.Metadata().Signature {
		t.Fatalf("expected stable signature, got %s and %s", first.Metadata().Signature, second.Metadata().Signature)
	}
	if first.Metadata().TermCount == 0 {
		t.Fatal("expected non-empty inverted index")
	}
}

func TestSearchReturnsChineseExplainAndSynonyms(t *testing.T) {
	cfg := Config{
		DefaultEngine: EngineHybrid,
		BatchOne: BatchOneConfig{
			Enabled:     true,
			WorkerCount: 2,
			Explain:     true,
			EnableIK:    true,
			EnableJieba: true,
		},
		BatchTwo: BatchTwoConfig{
			TrieEnabled:           true,
			SynonymEnabled:        true,
			FusionEnabled:         true,
			FusionBM25Weight:      0.75,
			FusionFreshnessWeight: 0.15,
			FusionQualityWeight:   0.10,
		},
	}.Normalized()

	store := stubStore{docs: sampleDocs(), fulltext: sampleDocs()[:1]}
	service := NewService(cfg, store, store)
	service.now = func() time.Time { return time.Unix(1_742_000_000, 0) }

	resp, err := service.Search(context.Background(), Request{
		Query:    "人工智能论文",
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if resp.Meta.Engine != EngineHybrid {
		t.Fatalf("expected hybrid engine, got %s", resp.Meta.Engine)
	}
	if len(resp.Papers) == 0 {
		t.Fatal("expected search hits")
	}
	if len(resp.QueryAnalysis.IKTokens) == 0 || len(resp.QueryAnalysis.JiebaTokens) == 0 {
		t.Fatalf("expected Chinese query analysis, got %+v", resp.QueryAnalysis)
	}
	if !slices.Contains(resp.QueryAnalysis.ExpandedTerms, "paper") {
		t.Fatalf("expected synonym expansion to include paper, got %+v", resp.QueryAnalysis.ExpandedTerms)
	}
	if len(resp.Explains) == 0 || len(resp.Explains[0].MatchedTerms) == 0 {
		t.Fatalf("expected explain output, got %+v", resp.Explains)
	}
}

func TestSuggestHonorsTrieFlag(t *testing.T) {
	cfg := Config{
		DefaultEngine: EngineHybrid,
		BatchOne: BatchOneConfig{
			Enabled:     true,
			WorkerCount: 2,
			EnableIK:    true,
			EnableJieba: true,
		},
		BatchTwo: BatchTwoConfig{
			TrieEnabled: true,
		},
	}.Normalized()
	store := stubStore{docs: sampleDocs(), fulltext: sampleDocs()[:1]}
	service := NewService(cfg, store, store)

	if _, err := service.Search(context.Background(), Request{Query: "机器学习", Page: 1, PageSize: 10}); err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	suggestions := service.Suggest("机", 5)
	if len(suggestions) == 0 {
		t.Fatal("expected trie suggestions")
	}

	cfg.BatchTwo.TrieEnabled = false
	noTrie := NewService(cfg, store, store)
	if _, err := noTrie.Search(context.Background(), Request{Query: "机器学习", Page: 1, PageSize: 10}); err != nil {
		t.Fatalf("Search failed without trie: %v", err)
	}
	if got := noTrie.Suggest("机", 5); len(got) != 0 {
		t.Fatalf("expected no suggestions when trie is disabled, got %v", got)
	}
}

func TestSearchFallsBackToFulltextWhenIndexBuildFails(t *testing.T) {
	cfg := Config{
		DefaultEngine: EngineHybrid,
		MaxDocuments:  1,
		BatchOne: BatchOneConfig{
			Enabled:     true,
			WorkerCount: 2,
			EnableIK:    true,
			EnableJieba: true,
		},
	}.Normalized()
	store := stubStore{docs: sampleDocs(), fulltext: sampleDocs()[:1]}
	service := NewService(cfg, store, store)

	resp, err := service.Search(context.Background(), Request{Query: "机器学习", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if resp.Meta.Engine != EngineFulltext {
		t.Fatalf("expected fulltext fallback, got %s", resp.Meta.Engine)
	}
	if !resp.Meta.UsedFallback {
		t.Fatal("expected fallback metadata")
	}
}

func TestSearchShadowCompareKeepsFulltextPrimary(t *testing.T) {
	cfg := Config{
		DefaultEngine: EngineFulltext,
		ShadowCompare: true,
		BatchOne: BatchOneConfig{
			Enabled:     true,
			WorkerCount: 2,
			EnableIK:    true,
			EnableJieba: true,
		},
		BatchTwo: BatchTwoConfig{
			TrieEnabled:    true,
			SynonymEnabled: true,
		},
	}.Normalized()
	fulltextOnly := []*model.Paper{sampleDocs()[2]}
	store := stubStore{docs: sampleDocs(), fulltext: fulltextOnly}
	service := NewService(cfg, store, store)

	resp, err := service.Search(context.Background(), Request{Query: "搜索", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if resp.Meta.Engine != EngineFulltext {
		t.Fatalf("expected fulltext primary, got %s", resp.Meta.Engine)
	}
	if !resp.Meta.ShadowCompared {
		t.Fatal("expected shadow comparison flag")
	}
	if len(resp.Papers) != 1 || resp.Papers[0].Id != fulltextOnly[0].Id {
		t.Fatalf("expected fulltext response to win, got %+v", resp.Papers)
	}
}

func TestSearchRebuildKeepsActiveVersionStable(t *testing.T) {
	cfg := Config{
		DefaultEngine: EngineHybrid,
		BatchOne: BatchOneConfig{
			Enabled:     true,
			WorkerCount: 2,
			EnableIK:    true,
			EnableJieba: true,
		},
	}.Normalized()
	store := stubStore{docs: sampleDocs(), fulltext: sampleDocs()[:1]}
	service := NewService(cfg, store, store)

	first, err := service.Search(context.Background(), Request{Query: "机器学习", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("first search failed: %v", err)
	}
	if first.Meta.Build.Version == "" || first.Meta.Build.Checksum == "" {
		t.Fatalf("expected versioned build metadata, got %+v", first.Meta.Build)
	}
	if len(first.Meta.Build.Segments) == 0 {
		t.Fatalf("expected segment metadata, got %+v", first.Meta.Build)
	}

	service.invalidateActiveArtifact()

	second, err := service.Search(context.Background(), Request{Query: "机器学习", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("second search failed: %v", err)
	}
	if second.Meta.Build.Version != first.Meta.Build.Version {
		t.Fatalf("expected stable version across rebuilds, got %s then %s", first.Meta.Build.Version, second.Meta.Build.Version)
	}
	if second.Meta.Build.DocumentCount != first.Meta.Build.DocumentCount {
		t.Fatalf("expected stable document count across rebuilds, got %d then %d", first.Meta.Build.DocumentCount, second.Meta.Build.DocumentCount)
	}
}

func TestSearchFallsBackToFulltextWhenActiveArtifactChecksumMismatch(t *testing.T) {
	cfg := Config{
		DefaultEngine: EngineHybrid,
		BatchOne: BatchOneConfig{
			Enabled:     true,
			WorkerCount: 2,
			EnableIK:    true,
			EnableJieba: true,
		},
	}.Normalized()
	store := stubStore{docs: sampleDocs(), fulltext: sampleDocs()[:1]}
	service := NewService(cfg, store, store)

	first, err := service.Search(context.Background(), Request{Query: "机器学习", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("initial search failed: %v", err)
	}
	service.active.metadata.Checksum = "corrupted"

	resp, err := service.Search(context.Background(), Request{Query: "机器学习", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("fallback search failed: %v", err)
	}
	if resp.Meta.Engine != EngineFulltext || !resp.Meta.UsedFallback {
		t.Fatalf("expected fulltext fallback on checksum mismatch, got %+v", resp.Meta)
	}
	if !strings.Contains(resp.Meta.FallbackReason, "checksum mismatch") {
		t.Fatalf("expected checksum mismatch in fallback reason, got %q", resp.Meta.FallbackReason)
	}
	if resp.Meta.Build.Version != first.Meta.Build.Version {
		t.Fatalf("expected cached successful version to remain available, got %+v", resp.Meta.Build)
	}
	if service.active != nil {
		t.Fatal("expected invalid active artifact to be cleared")
	}
	if service.lastSuccessful == nil {
		t.Fatal("expected last successful artifact to be retained")
	}
}

func TestSearchFallsBackToFulltextWhenActiveArtifactCannotLoad(t *testing.T) {
	cfg := Config{
		DefaultEngine: EngineHybrid,
		BatchOne: BatchOneConfig{
			Enabled:     true,
			WorkerCount: 2,
			EnableIK:    true,
			EnableJieba: true,
		},
	}.Normalized()
	store := stubStore{docs: sampleDocs(), fulltext: sampleDocs()[:1]}
	service := NewService(cfg, store, store)

	first, err := service.Search(context.Background(), Request{Query: "机器学习", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("initial search failed: %v", err)
	}
	service.active.snapshot = nil

	resp, err := service.Search(context.Background(), Request{Query: "机器学习", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("fallback search failed: %v", err)
	}
	if resp.Meta.Engine != EngineFulltext || !resp.Meta.UsedFallback {
		t.Fatalf("expected fulltext fallback on load failure, got %+v", resp.Meta)
	}
	if !strings.Contains(resp.Meta.FallbackReason, "snapshot unavailable") {
		t.Fatalf("expected load failure in fallback reason, got %q", resp.Meta.FallbackReason)
	}
	if resp.Meta.Build.Version != first.Meta.Build.Version {
		t.Fatalf("expected cached successful version to remain available, got %+v", resp.Meta.Build)
	}
	if service.lastSuccessful == nil {
		t.Fatal("expected last successful artifact to be retained")
	}
}

type stubStore struct {
	docs            []*model.Paper
	fulltext        []*model.Paper
	fulltextByQuery map[string][]*model.Paper
}

func (s stubStore) ListSearchDocuments(ctx context.Context, limit int) ([]*model.Paper, error) {
	if limit > len(s.docs) {
		limit = len(s.docs)
	}
	return slices.Clone(s.docs[:limit]), nil
}

func (s stubStore) Search(ctx context.Context, query, discipline string, page, pageSize int) ([]*model.Paper, int64, error) {
	source := s.fulltext
	if items, ok := s.fulltextByQuery[query]; ok {
		source = items
	}
	items := make([]*model.Paper, 0, len(source))
	for _, doc := range source {
		if discipline != "" && doc.Discipline != discipline {
			continue
		}
		items = append(items, doc)
	}
	return items, int64(len(items)), nil
}

func sampleDocs() []*model.Paper {
	base := time.Unix(1_741_500_000, 0)
	return []*model.Paper{
		{
			Id:         1,
			Title:      "人工智能论文推荐系统",
			Abstract:   "面向中文论文检索的机器学习与排序融合研究。",
			Keywords:   "人工智能,机器学习,推荐系统",
			Discipline: "cs",
			ShitScore:  7.2,
			CreatedAt:  base.Add(-48 * time.Hour),
			UpdatedAt:  base.Add(-12 * time.Hour),
			AuthorName: "alice",
		},
		{
			Id:         2,
			Title:      "机器学习在量子计算中的应用",
			Abstract:   "使用深度学习方法加速量子态分析与论文召回。",
			Keywords:   "机器学习,量子计算,深度学习",
			Discipline: "cs",
			ShitScore:  8.1,
			CreatedAt:  base.Add(-72 * time.Hour),
			UpdatedAt:  base.Add(-24 * time.Hour),
			AuthorName: "bob",
		},
		{
			Id:         3,
			Title:      "信息检索中的 Explain 排序调试",
			Abstract:   "讨论 BM25 explain、搜索日志与回退链路。",
			Keywords:   "搜索,Explain,BM25,检索",
			Discipline: "cs",
			ShitScore:  6.6,
			CreatedAt:  base.Add(-96 * time.Hour),
			UpdatedAt:  base.Add(-36 * time.Hour),
			AuthorName: "carol",
		},
	}
}

func TestLoadersStayDeterministic(t *testing.T) {
	store := stubStore{docs: sampleDocs(), fulltext: sampleDocs()[:1]}
	first, err := store.ListSearchDocuments(context.Background(), 3)
	if err != nil {
		t.Fatalf("ListSearchDocuments failed: %v", err)
	}
	second, err := store.ListSearchDocuments(context.Background(), 3)
	if err != nil {
		t.Fatalf("ListSearchDocuments failed: %v", err)
	}
	for idx := range first {
		if fmt.Sprintf("%+v", first[idx]) != fmt.Sprintf("%+v", second[idx]) {
			t.Fatalf("expected deterministic loader output at index %d", idx)
		}
	}
}

func TestSearchComparisonReport(t *testing.T) {
	cfg := Config{
		DefaultEngine: EngineFulltext,
		BatchOne: BatchOneConfig{
			Enabled:     true,
			WorkerCount: 2,
			Explain:     true,
			EnableIK:    true,
			EnableJieba: true,
		},
		BatchTwo: BatchTwoConfig{
			TrieEnabled:           true,
			SynonymEnabled:        true,
			FusionEnabled:         true,
			FusionBM25Weight:      0.75,
			FusionFreshnessWeight: 0.15,
			FusionQualityWeight:   0.10,
		},
	}.Normalized()
	docs := sampleDocs()
	store := stubStore{
		docs: docs,
		fulltextByQuery: map[string][]*model.Paper{
			"人工智能论文": {docs[0]},
			"搜索":     {docs[2]},
		},
	}
	service := NewService(cfg, store, store)
	service.now = func() time.Time { return time.Unix(1_742_000_000, 0) }

	for _, query := range []string{"人工智能论文", "搜索"} {
		hybridStarted := time.Now()
		hybridResp, err := service.Search(context.Background(), Request{
			Query:           query,
			Page:            1,
			PageSize:        5,
			Sort:            "relevance",
			Engine:          EngineHybrid,
			SuggestionLimit: 4,
		})
		if err != nil {
			t.Fatalf("hybrid search failed for %q: %v", query, err)
		}
		fulltextStarted := time.Now()
		fulltextResp, err := service.Search(context.Background(), Request{
			Query:           query,
			Page:            1,
			PageSize:        5,
			Sort:            "relevance",
			Engine:          EngineFulltext,
			Shadow:          true,
			SuggestionLimit: 4,
		})
		if err != nil {
			t.Fatalf("fulltext search failed for %q: %v", query, err)
		}
		t.Logf(
			"query=%s hybrid_ids=%v fulltext_ids=%v suggestions=%v latency_hybrid=%s latency_fulltext=%s",
			query,
			paperIDs(hybridResp.Papers),
			paperIDs(fulltextResp.Papers),
			hybridResp.Suggestions,
			time.Since(hybridStarted),
			time.Since(fulltextStarted),
		)
	}
}

func paperIDs(items []*model.Paper) []int64 {
	ids := make([]int64, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.Id)
	}
	return ids
}
