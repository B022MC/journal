package search

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	"journal/model"
)

type goldenFixture struct {
	Documents []goldenDocument `json:"documents"`
	Queries   []goldenQuery    `json:"queries"`
}

type goldenDocument struct {
	ID         int64   `json:"id"`
	Title      string  `json:"title"`
	TitleEn    string  `json:"title_en"`
	Abstract   string  `json:"abstract"`
	Keywords   string  `json:"keywords"`
	Discipline string  `json:"discipline"`
	ShitScore  float64 `json:"shit_score"`
}

type goldenQuery struct {
	Name       string  `json:"name"`
	Query      string  `json:"query"`
	Discipline string  `json:"discipline"`
	RelevantID []int64 `json:"relevant_ids"`
}

type goldenStore struct {
	cfg  Config
	docs []*model.Paper
}

func TestGoldenSearchBenchmarks(t *testing.T) {
	fixture := loadGoldenFixture(t)
	cfg := Config{
		DefaultEngine: EngineHybrid,
		BatchOne: BatchOneConfig{
			Enabled:     true,
			WorkerCount: 2,
			// Keep explain payload in responses, but exclude explain logging from latency probes.
			Explain:     false,
			EnableIK:    true,
			EnableJieba: true,
		},
		BatchTwo: BatchTwoConfig{
			TrieEnabled:    true,
			SynonymEnabled: true,
		},
	}.Normalized()
	store := goldenStore{cfg: cfg, docs: fixture.documents()}
	service := NewService(cfg, store, store)

	// Warm the snapshot before measuring latency.
	if _, err := service.Search(context.Background(), Request{
		Query:           fixture.Queries[0].Query,
		Page:            1,
		PageSize:        10,
		Engine:          EngineHybrid,
		SuggestionLimit: 0,
	}); err != nil {
		t.Fatalf("warm-up hybrid search failed: %v", err)
	}

	hybridRecalls := make([]float64, 0, len(fixture.Queries))
	fulltextRecalls := make([]float64, 0, len(fixture.Queries))
	hybridLatencies := make([]float64, 0, len(fixture.Queries)*5)
	fulltextLatencies := make([]float64, 0, len(fixture.Queries)*5)
	explainHits := 0
	explainTotal := 0

	for _, query := range fixture.Queries {
		relevant := make(map[int64]struct{}, len(query.RelevantID))
		for _, id := range query.RelevantID {
			relevant[id] = struct{}{}
		}

		hybridResp, err := service.Search(context.Background(), Request{
			Query:           query.Query,
			Discipline:      query.Discipline,
			Page:            1,
			PageSize:        10,
			Engine:          EngineHybrid,
			SuggestionLimit: 0,
		})
		if err != nil {
			t.Fatalf("hybrid search failed for %s: %v", query.Name, err)
		}
		fulltextResp, err := service.Search(context.Background(), Request{
			Query:           query.Query,
			Discipline:      query.Discipline,
			Page:            1,
			PageSize:        10,
			Engine:          EngineFulltext,
			SuggestionLimit: 0,
		})
		if err != nil {
			t.Fatalf("fulltext search failed for %s: %v", query.Name, err)
		}

		hybridRecalls = append(hybridRecalls, recallAtK(resultIDs(hybridResp.Papers), relevant, 10))
		fulltextRecalls = append(fulltextRecalls, recallAtK(resultIDs(fulltextResp.Papers), relevant, 10))

		for _, explain := range hybridResp.Explains {
			explainTotal++
			if len(explain.MatchedTerms) > 0 {
				explainHits++
			}
		}

		for range 5 {
			started := time.Now()
			if _, err := service.Search(context.Background(), Request{
				Query:           query.Query,
				Discipline:      query.Discipline,
				Page:            1,
				PageSize:        10,
				Engine:          EngineHybrid,
				SuggestionLimit: 0,
			}); err != nil {
				t.Fatalf("hybrid latency probe failed for %s: %v", query.Name, err)
			}
			hybridLatencies = append(hybridLatencies, durationMs(time.Since(started)))

			started = time.Now()
			if _, err := service.Search(context.Background(), Request{
				Query:           query.Query,
				Discipline:      query.Discipline,
				Page:            1,
				PageSize:        10,
				Engine:          EngineFulltext,
				SuggestionLimit: 0,
			}); err != nil {
				t.Fatalf("fulltext latency probe failed for %s: %v", query.Name, err)
			}
			fulltextLatencies = append(fulltextLatencies, durationMs(time.Since(started)))
		}
	}

	service.invalidateActiveArtifact()
	first, err := service.Search(context.Background(), Request{
		Query:           fixture.Queries[0].Query,
		Page:            1,
		PageSize:        10,
		Engine:          EngineHybrid,
		SuggestionLimit: 0,
	})
	if err != nil {
		t.Fatalf("first rebuild failed: %v", err)
	}
	service.invalidateActiveArtifact()
	second, err := service.Search(context.Background(), Request{
		Query:           fixture.Queries[0].Query,
		Page:            1,
		PageSize:        10,
		Engine:          EngineHybrid,
		SuggestionLimit: 0,
	})
	if err != nil {
		t.Fatalf("second rebuild failed: %v", err)
	}

	hybridRecall := average(hybridRecalls)
	fulltextRecall := average(fulltextRecalls)
	hybridP50 := percentile(hybridLatencies, 0.50)
	hybridP95 := percentile(hybridLatencies, 0.95)
	fulltextP50 := percentile(fulltextLatencies, 0.50)
	fulltextP95 := percentile(fulltextLatencies, 0.95)
	explainCompleteness := 0.0
	if explainTotal > 0 {
		explainCompleteness = float64(explainHits) / float64(explainTotal)
	}
	stableBuild := first.Meta.Build.Version == second.Meta.Build.Version && len(first.Papers) == len(second.Papers)

	t.Logf(
		"golden_bench fixture=%d queries=%d recall_at_10 hybrid=%.2f fulltext=%.2f latency_ms_p50 hybrid=%.3f fulltext=%.3f latency_ms_p95 hybrid=%.3f fulltext=%.3f explain_completeness=%.2f rebuild_stable=%t version=%s",
		len(fixture.Documents),
		len(fixture.Queries),
		hybridRecall,
		fulltextRecall,
		hybridP50,
		fulltextP50,
		hybridP95,
		fulltextP95,
		explainCompleteness,
		stableBuild,
		first.Meta.Build.Version,
	)

	if hybridRecall < fulltextRecall-0.05 {
		t.Fatalf("hybrid recall %.2f is worse than fulltext %.2f beyond 5%%", hybridRecall, fulltextRecall)
	}
	if hybridP95 > fulltextP95*1.2 {
		t.Fatalf("hybrid p95 %.3f ms exceeds fulltext p95 %.3f ms by more than 20%%", hybridP95, fulltextP95)
	}
	if explainCompleteness < 1 {
		t.Fatalf("expected explain completeness to be 1.00, got %.2f", explainCompleteness)
	}
	if !stableBuild {
		t.Fatalf("expected repeated rebuilds to keep version and result count stable, got %s/%s and %d/%d", first.Meta.Build.Version, second.Meta.Build.Version, len(first.Papers), len(second.Papers))
	}
}

func loadGoldenFixture(t *testing.T) goldenFixture {
	t.Helper()
	path := filepath.Join("testdata", "golden_search_fixture.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden fixture: %v", err)
	}
	var fixture goldenFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("decode golden fixture: %v", err)
	}
	return fixture
}

func (f goldenFixture) documents() []*model.Paper {
	docs := make([]*model.Paper, 0, len(f.Documents))
	base := time.Unix(1_742_000_000, 0)
	for idx, item := range f.Documents {
		docs = append(docs, &model.Paper{
			Id:         item.ID,
			Title:      item.Title,
			TitleEn:    item.TitleEn,
			Abstract:   item.Abstract,
			Keywords:   item.Keywords,
			Discipline: item.Discipline,
			ShitScore:  item.ShitScore,
			CreatedAt:  base.Add(-time.Duration(idx+1) * time.Hour),
			UpdatedAt:  base.Add(-time.Duration(idx) * time.Hour),
		})
	}
	return docs
}

func (s goldenStore) ListSearchDocuments(ctx context.Context, limit int) ([]*model.Paper, error) {
	if limit > len(s.docs) {
		limit = len(s.docs)
	}
	return slices.Clone(s.docs[:limit]), nil
}

func (s goldenStore) Search(ctx context.Context, query, discipline string, page, pageSize int) ([]*model.Paper, int64, error) {
	analysis, err := analyzeQuery(query, s.cfg.BatchOne, builtInLexicon, builtInSynonyms, false)
	if err != nil {
		return nil, 0, err
	}
	type scored struct {
		paper  *model.Paper
		score  int
		order  int
	}
	items := make([]scored, 0, len(s.docs))
	for idx, doc := range s.docs {
		if discipline != "" && doc.Discipline != discipline {
			continue
		}
		text := strings.Join([]string{
			normalizeText(doc.Title),
			normalizeText(doc.TitleEn),
			normalizeText(doc.Abstract),
			normalizeText(doc.Keywords),
		}, " ")
		score := 0
		for _, term := range analysis.ExpandedTerms {
			if strings.Contains(text, normalizeText(term)) {
				score++
			}
		}
		if score == 0 {
			continue
		}
		items = append(items, scored{paper: doc, score: score, order: idx})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].score == items[j].score {
			return items[i].order < items[j].order
		}
		return items[i].score > items[j].score
	})
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	if offset > len(items) {
		offset = len(items)
	}
	end := offset + pageSize
	if end > len(items) {
		end = len(items)
	}
	out := make([]*model.Paper, 0, end-offset)
	for _, item := range items[offset:end] {
		out = append(out, item.paper)
	}
	return out, int64(len(items)), nil
}

func recallAtK(ids []int64, relevant map[int64]struct{}, k int) float64 {
	if len(relevant) == 0 {
		return 1
	}
	if k > len(ids) {
		k = len(ids)
	}
	hits := 0
	for _, id := range ids[:k] {
		if _, ok := relevant[id]; ok {
			hits++
		}
	}
	return float64(hits) / float64(len(relevant))
}

func durationMs(duration time.Duration) float64 {
	return float64(duration) / float64(time.Millisecond)
}

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	total := 0.0
	for _, value := range values {
		total += value
	}
	return total / float64(len(values))
}

func percentile(values []float64, q float64) float64 {
	if len(values) == 0 {
		return 0
	}
	ordered := append([]float64(nil), values...)
	sort.Float64s(ordered)
	index := int(float64(len(ordered)-1) * q)
	if index < 0 {
		index = 0
	}
	if index >= len(ordered) {
		index = len(ordered) - 1
	}
	return ordered[index]
}
