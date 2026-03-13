package search

import (
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestSearchTelemetryRecordsRequestAndActiveVersion(t *testing.T) {
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
	metrics := getSearchMetrics()

	before := testutil.ToFloat64(metrics.requests.WithLabelValues(EngineHybrid, EngineHybrid, "success"))
	resp, err := service.Search(context.Background(), Request{Query: "机器学习", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	after := testutil.ToFloat64(metrics.requests.WithLabelValues(EngineHybrid, EngineHybrid, "success"))
	if after != before+1 {
		t.Fatalf("expected request counter to increment, got before=%v after=%v", before, after)
	}
	if version := resp.Meta.Build.Version; version == "" {
		t.Fatal("expected active version in response metadata")
	} else if got := testutil.ToFloat64(metrics.activeVersion.WithLabelValues(version)); got != 1 {
		t.Fatalf("expected active version gauge to be 1, got %v", got)
	}
}

func TestSearchTelemetryRecordsShadowDeltaKind(t *testing.T) {
	cfg := Config{
		DefaultEngine: EngineFulltext,
		ShadowCompare: true,
		BatchOne: BatchOneConfig{
			Enabled:     true,
			WorkerCount: 2,
			EnableIK:    true,
			EnableJieba: true,
		},
	}.Normalized()
	docs := sampleDocs()
	store := stubStore{
		docs:     docs,
		fulltext: docs[:1],
	}
	service := NewService(cfg, store, store)
	metrics := getSearchMetrics()

	before := testutil.ToFloat64(metrics.shadowDelta.WithLabelValues("recall_diff"))
	resp, err := service.Search(context.Background(), Request{Query: "机器学习", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if !resp.Meta.ShadowCompared {
		t.Fatal("expected successful shadow compare")
	}
	after := testutil.ToFloat64(metrics.shadowDelta.WithLabelValues("recall_diff"))
	if after != before+1 {
		t.Fatalf("expected recall_diff counter to increment, got before=%v after=%v", before, after)
	}
}

func TestSearchTelemetryRecordsBuildFailure(t *testing.T) {
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
	metrics := getSearchMetrics()

	before := testutil.ToFloat64(metrics.buildFailures)
	resp, err := service.Search(context.Background(), Request{Query: "机器学习", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if resp.Meta.FallbackReason != FallbackReasonIndexBuildFailed {
		t.Fatalf("expected build failure fallback reason, got %q", resp.Meta.FallbackReason)
	}
	after := testutil.ToFloat64(metrics.buildFailures)
	if after != before+1 {
		t.Fatalf("expected build failure counter to increment, got before=%v after=%v", before, after)
	}
}

func TestBuildSearchLogFieldsIncludesRequiredFields(t *testing.T) {
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

	ctx := withSearchRequestID(context.Background())
	req := Request{Query: "机器学习", Page: 1, PageSize: 10}
	resp, err := service.Search(ctx, req)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	fields := service.buildSearchLogFields(ctx, req, resp, requestMode(EngineHybrid, false), FallbackReasonNone)
	if fields.RequestID == "" {
		t.Fatal("expected request id in log fields")
	}
	if fields.Engine != EngineHybrid || fields.Mode != EngineHybrid {
		t.Fatalf("expected hybrid engine and mode, got %+v", fields)
	}
	if fields.IndexVersion == "" || fields.NormalizedQuery == "" {
		t.Fatalf("expected version and normalized query, got %+v", fields)
	}
	if len(fields.TokenList) == 0 || len(fields.TopResultIDs) == 0 {
		t.Fatalf("expected token list and top ids, got %+v", fields)
	}
}
