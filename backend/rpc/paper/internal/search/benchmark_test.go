package search

import (
	"context"
	"fmt"
	"testing"
	"time"

	"journal/model"
)

func BenchmarkBuildSnapshot(b *testing.B) {
	cfg := Config{
		DefaultEngine: EngineHybrid,
		BatchOne: BatchOneConfig{
			Enabled:     true,
			WorkerCount: 4,
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
	docs := benchmarkDocs(256)
	b.ReportAllocs()
	for b.Loop() {
		snapshot, err := BuildSnapshot(context.Background(), docs, cfg, builtInLexicon, builtInSynonyms)
		if err != nil {
			b.Fatalf("BuildSnapshot failed: %v", err)
		}
		if snapshot.Metadata().TermCount == 0 {
			b.Fatal("expected non-empty term count")
		}
	}
}

func BenchmarkServiceSearch(b *testing.B) {
	cfg := Config{
		DefaultEngine: EngineHybrid,
		BatchOne: BatchOneConfig{
			Enabled:     true,
			WorkerCount: 4,
			Explain:     false,
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
	store := stubStore{
		docs:     benchmarkDocs(256),
		fulltext: sampleDocs()[:1],
	}
	service := NewService(cfg, store, store)
	service.now = func() time.Time { return time.Unix(1_742_000_000, 0) }

	b.ReportAllocs()
	for b.Loop() {
		resp, err := service.Search(context.Background(), Request{
			Query:    "人工智能论文",
			Page:     1,
			PageSize: 10,
		})
		if err != nil {
			b.Fatalf("Search failed: %v", err)
		}
		if len(resp.Papers) == 0 {
			b.Fatal("expected non-empty hits")
		}
	}
}

func benchmarkDocs(size int) []*model.Paper {
	docs := make([]*model.Paper, 0, size)
	base := time.Unix(1_741_500_000, 0)
	for idx := range size {
		docs = append(docs, &model.Paper{
			Id:         int64(idx + 1),
			Title:      fmt.Sprintf("人工智能论文推荐系统 %d", idx),
			Abstract:   fmt.Sprintf("第 %d 篇文档讨论机器学习、信息检索和 Explain 输出。", idx),
			Keywords:   "人工智能,机器学习,搜索,Explain",
			Discipline: "cs",
			ShitScore:  5 + float64(idx%7),
			CreatedAt:  base.Add(-time.Duration(idx) * time.Hour),
			UpdatedAt:  base.Add(-time.Duration(idx/2) * time.Hour),
		})
	}
	return docs
}
