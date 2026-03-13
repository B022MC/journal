package search

import (
	"context"
	"fmt"
	"sync"
	"time"

	"journal/model"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/metadata"
)

type searchRequestIDKey struct{}

type searchMetrics struct {
	requests        *prometheus.CounterVec
	latency         *prometheus.HistogramVec
	buildDuration   *prometheus.HistogramVec
	buildFailures   prometheus.Counter
	shadowDelta     *prometheus.CounterVec
	activeVersion   *prometheus.GaugeVec
	activeVersionMu sync.Mutex
	activeVersionID string
}

var (
	searchMetricsOnce sync.Once
	searchMetricSet   *searchMetrics
)

type searchLogFields struct {
	RequestID      string
	Engine         string
	Mode           string
	CompareEngine  string
	IndexVersion   string
	NormalizedQuery string
	TokenList      []string
	TopResultIDs   []int64
	FallbackReason string
	CompareReason  string
	ShadowCompared bool
}

func getSearchMetrics() *searchMetrics {
	searchMetricsOnce.Do(func() {
		searchMetricSet = &searchMetrics{
			requests: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "search_requests_total",
				Help: "Total search requests by answer engine, requested mode, and result.",
			}, []string{"engine", "mode", "result"}),
			latency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Name:    "search_latency_ms",
				Help:    "Search request latency in milliseconds by answer engine and requested mode.",
				Buckets: []float64{1, 5, 10, 25, 50, 100, 200, 500, 1000, 2000, 5000},
			}, []string{"engine", "mode"}),
			buildDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Name:    "search_index_build_duration_seconds",
				Help:    "Search index build duration by logical shard.",
				Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10},
			}, []string{"shard"}),
			buildFailures: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "search_index_build_failures_total",
				Help: "Total failed search index build attempts.",
			}),
			shadowDelta: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "search_shadow_delta_total",
				Help: "Total shadow comparison deltas by kind.",
			}, []string{"kind"}),
			activeVersion: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "search_active_index_version",
				Help: "Active search index version exposed as a single-version gauge.",
			}, []string{"version"}),
		}
		prometheus.MustRegister(
			searchMetricSet.requests,
			searchMetricSet.latency,
			searchMetricSet.buildDuration,
			searchMetricSet.buildFailures,
			searchMetricSet.shadowDelta,
			searchMetricSet.activeVersion,
		)
	})
	return searchMetricSet
}

func observeSearchRequest(engine, mode, result string, duration time.Duration) {
	metrics := getSearchMetrics()
	metrics.requests.WithLabelValues(engine, mode, result).Inc()
	metrics.latency.WithLabelValues(engine, mode).Observe(float64(duration.Milliseconds()))
}

func observeBuildSuccess(meta BuildMetadata) {
	metrics := getSearchMetrics()
	if len(meta.Segments) == 0 {
		metrics.buildDuration.WithLabelValues("all").Observe(meta.BuildDuration.Seconds())
	} else {
		for _, segment := range meta.Segments {
			metrics.buildDuration.WithLabelValues(segment.Name).Observe(meta.BuildDuration.Seconds())
		}
	}
	metrics.activeVersionMu.Lock()
	defer metrics.activeVersionMu.Unlock()
	if metrics.activeVersionID != "" && metrics.activeVersionID != meta.Version {
		metrics.activeVersion.DeleteLabelValues(metrics.activeVersionID)
	}
	if meta.Version != "" {
		metrics.activeVersion.WithLabelValues(meta.Version).Set(1)
		metrics.activeVersionID = meta.Version
	}
}

func observeBuildFailure() {
	getSearchMetrics().buildFailures.Inc()
}

func observeShadowDelta(kind string) {
	getSearchMetrics().shadowDelta.WithLabelValues(kind).Inc()
}

func withSearchRequestID(ctx context.Context) context.Context {
	if current := searchRequestID(ctx); current != "" {
		return context.WithValue(ctx, searchRequestIDKey{}, current)
	}
	return context.WithValue(ctx, searchRequestIDKey{}, uuid.NewString())
}

func searchRequestID(ctx context.Context) string {
	if value := ctx.Value(searchRequestIDKey{}); value != nil {
		if id, ok := value.(string); ok {
			return id
		}
	}
	if incoming, ok := metadata.FromIncomingContext(ctx); ok {
		for _, key := range []string{"x-request-id", "request-id", "request_id"} {
			if values := incoming.Get(key); len(values) > 0 {
				return values[0]
			}
		}
	}
	return ""
}

func requestMode(engine string, shadow bool) string {
	if engine == EngineHybrid {
		return EngineHybrid
	}
	if shadow {
		return "shadow"
	}
	return EngineFulltext
}

func requestResult(resp Response, compareReason string) string {
	switch {
	case resp.Meta.UsedFallback:
		return "fallback"
	case compareReason != FallbackReasonNone:
		return "compare_error"
	default:
		return "success"
	}
}

func (s *Service) analysisForLog(req Request, resp Response) QueryAnalysis {
	if resp.QueryAnalysis.Raw != "" || len(resp.QueryAnalysis.ExpandedTerms) > 0 {
		return resp.QueryAnalysis
	}
	analysis, err := analyzeQuery(req.Query, s.cfg.BatchOne, s.lexicon, s.synonyms, s.cfg.BatchTwo.SynonymEnabled)
	if err != nil {
		return QueryAnalysis{Raw: normalizeText(req.Query)}
	}
	return analysis
}

func (s *Service) buildSearchLogFields(ctx context.Context, req Request, resp Response, mode, compareReason string) searchLogFields {
	analysis := s.analysisForLog(req, resp)
	return searchLogFields{
		RequestID:       searchRequestID(ctx),
		Engine:          resp.Meta.Engine,
		Mode:            mode,
		CompareEngine:   compareEngineForMode(mode),
		IndexVersion:    resp.Meta.Build.Version,
		NormalizedQuery: analysis.Raw,
		TokenList:       append([]string{}, analysis.ExpandedTerms...),
		TopResultIDs:    resultIDs(resp.Papers),
		FallbackReason:  resp.Meta.FallbackReason,
		CompareReason:   compareReason,
		ShadowCompared:  resp.Meta.ShadowCompared,
	}
}

func compareEngineForMode(mode string) string {
	if mode == "shadow" {
		return EngineHybrid
	}
	return "disabled"
}

func resultIDs(papers []*model.Paper) []int64 {
	ids := make([]int64, 0, len(papers))
	for _, paper := range papers {
		ids = append(ids, paper.Id)
	}
	return ids
}

func classifyShadowDelta(req Request, fulltextResp, hybridResp Response) string {
	if isCrossQueryContamination(hybridResp) {
		return "cross_query_contamination"
	}
	fulltextIDs := resultIDs(fulltextResp.Papers)
	hybridIDs := resultIDs(hybridResp.Papers)
	if len(fulltextIDs) != len(hybridIDs) {
		return "recall_diff"
	}
	leftSet := make(map[int64]struct{}, len(fulltextIDs))
	for _, id := range fulltextIDs {
		leftSet[id] = struct{}{}
	}
	for _, id := range hybridIDs {
		if _, ok := leftSet[id]; !ok {
			return "recall_diff"
		}
	}
	for idx := range fulltextIDs {
		if fulltextIDs[idx] != hybridIDs[idx] {
			return "ranking_diff"
		}
	}
	return "match"
}

func isCrossQueryContamination(resp Response) bool {
	if len(resp.QueryAnalysis.ExpandedTerms) == 0 || len(resp.Explains) == 0 {
		return false
	}
	allowed := make(map[string]struct{}, len(resp.QueryAnalysis.ExpandedTerms))
	for _, term := range resp.QueryAnalysis.ExpandedTerms {
		allowed[term] = struct{}{}
	}
	for _, explain := range resp.Explains {
		matched := false
		for _, term := range explain.MatchedTerms {
			if _, ok := allowed[term]; ok {
				matched = true
				break
			}
		}
		if !matched {
			return true
		}
	}
	return false
}

func formatSearchLog(fields searchLogFields) string {
	return fmt.Sprintf(
		"request_id=%s engine=%s mode=%s compare_engine=%s shadow_compared=%t index_version=%s normalized_query=%q tokens=%v top_ids=%v fallback_reason=%s compare_reason=%s",
		fields.RequestID,
		fields.Engine,
		fields.Mode,
		fields.CompareEngine,
		fields.ShadowCompared,
		fields.IndexVersion,
		fields.NormalizedQuery,
		fields.TokenList,
		fields.TopResultIDs,
		fields.FallbackReason,
		fields.CompareReason,
	)
}
