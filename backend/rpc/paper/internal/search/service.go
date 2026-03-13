package search

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"journal/model"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	FallbackReasonNone                 = ""
	FallbackReasonEmptyQuery           = "empty_query"
	FallbackReasonBatchOneDisabled     = "batch_one_disabled"
	FallbackReasonIndexBuildFailed     = "index_build_failed"
	FallbackReasonQueryTimeout         = "query_timeout"
	FallbackReasonMissingIndexVersion  = "missing_index_version"
	FallbackReasonSegmentLoadFailed    = "segment_load_failed"
	FallbackReasonSegmentValidation    = "segment_validation_failed"
	FallbackReasonQueryParseFailed     = "query_parse_failed"
	FallbackReasonRankerError          = "ranker_error"
	FallbackReasonEngineError          = "engine_error"
	FallbackReasonShadowCompareFailure = "shadow_compare_failed"
)

var (
	errIndexBuildFailed  = errors.New("index build failed")
	errMissingVersion    = errors.New("missing index version")
	errSegmentLoadFailed = errors.New("segment load failed")
	errSegmentInvalid    = errors.New("segment validation failed")
	errQueryParseFailed  = errors.New("query parse failed")
	errRankerFailed      = errors.New("ranker failed")
)

type FulltextSearcher interface {
	Search(ctx context.Context, query, discipline string, page, pageSize int) ([]*model.Paper, int64, error)
}

type DocumentLoader interface {
	ListSearchDocuments(ctx context.Context, limit int) ([]*model.Paper, error)
}

type Service struct {
	cfg            Config
	loader         DocumentLoader
	fulltext       FulltextSearcher
	runSearch      func(ctx context.Context, snapshot *Snapshot, req Request, cfg Config, now time.Time) (Response, error)
	now            func() time.Time
	lexicon        []string
	synonyms       synonymMap
	buildMu        sync.Mutex
	mu             sync.RWMutex
	active         *indexArtifact
	lastSuccessful *indexArtifact
}

type indexArtifact struct {
	snapshot *Snapshot
	metadata BuildMetadata
}

func NewService(cfg Config, loader DocumentLoader, fulltext FulltextSearcher) *Service {
	normalized := cfg.Normalized()
	lexicon, err := loadLexicon(normalized.BatchOne.LexiconPath)
	if err != nil {
		logx.Errorf("search lexicon load failed, fallback to built-in dictionary: %v", err)
		lexicon, _ = loadLexicon("")
	}
	synonyms, err := loadSynonyms(normalized.BatchTwo.SynonymPath)
	if err != nil {
		logx.Errorf("search synonyms load failed, fallback to built-in dictionary: %v", err)
		synonyms, _ = loadSynonyms("")
	}
	return &Service{
		cfg:      normalized,
		loader:   loader,
		fulltext: fulltext,
		runSearch: func(ctx context.Context, snapshot *Snapshot, req Request, cfg Config, now time.Time) (Response, error) {
			return snapshot.Search(ctx, req, cfg, now)
		},
		now:      time.Now,
		lexicon:  lexicon,
		synonyms: synonyms,
	}
}

func (s *Service) Search(ctx context.Context, req Request) (Response, error) {
	ctx = withSearchRequestID(ctx)
	startedAt := time.Now()
	if s.fulltext == nil {
		return Response{}, fmt.Errorf("fulltext searcher is not configured")
	}
	mode := requestMode(s.resolveEngine(req.Engine), req.Shadow)
	if strings.TrimSpace(req.Query) == "" {
		resp, err := s.searchFulltext(ctx, req, FallbackReasonEmptyQuery)
		if err == nil {
			s.logSearchOutcome(ctx, req, resp, mode, FallbackReasonNone)
			observeSearchRequest(resp.Meta.Engine, mode, requestResult(resp, FallbackReasonNone), time.Since(startedAt))
		}
		return resp, err
	}
	if !s.cfg.BatchOne.Enabled || s.loader == nil {
		resp, err := s.searchFulltext(ctx, req, FallbackReasonBatchOneDisabled)
		if err == nil {
			s.logSearchOutcome(ctx, req, resp, mode, FallbackReasonNone)
			observeSearchRequest(resp.Meta.Engine, mode, requestResult(resp, FallbackReasonNone), time.Since(startedAt))
		}
		return resp, err
	}

	engine := s.resolveEngine(req.Engine)
	shadowCompare := s.resolveShadowCompare(engine, req.Shadow)

	switch engine {
	case EngineHybrid:
		response, err := s.searchNewEngine(ctx, req)
		if err != nil {
			reason := classifyFallbackReason(err)
			resp, fallbackErr := s.searchFulltext(ctx, req, reason)
			if fallbackErr == nil {
				s.logSearchOutcome(ctx, req, resp, mode, FallbackReasonNone)
				observeSearchRequest(resp.Meta.Engine, mode, requestResult(resp, FallbackReasonNone), time.Since(startedAt))
			}
			return resp, fallbackErr
		}
		s.logSearchOutcome(ctx, req, response, mode, FallbackReasonNone)
		observeSearchRequest(response.Meta.Engine, mode, requestResult(response, FallbackReasonNone), time.Since(startedAt))
		return response, nil
	default:
		fulltextResp, err := s.searchFulltext(ctx, req, "")
		if err != nil {
			return Response{}, err
		}
		compareReason := FallbackReasonNone
		if shadowCompare {
			if shadowResp, shadowErr := s.searchNewEngine(ctx, req); shadowErr != nil {
				compareReason = classifyFallbackReason(shadowErr)
				logx.WithContext(ctx).Infof(
					"search compare path request_id=%s normalized_query=%q mode=%s compare_engine=%s compare_reason=%s index_version=%s",
					searchRequestID(ctx),
					normalizeText(req.Query),
					mode,
					EngineHybrid,
					compareReason,
					s.Metadata().Version,
				)
			} else {
				s.logShadowComparison(ctx, req, fulltextResp, shadowResp)
				if len(fulltextResp.Suggestions) == 0 {
					fulltextResp.Suggestions = shadowResp.Suggestions
				}
				fulltextResp.Meta.ShadowCompared = true
			}
		}
		if len(fulltextResp.Suggestions) == 0 {
			fulltextResp.Suggestions = s.suggestionsForRequest(ctx, req)
		}
		s.logSearchOutcome(ctx, req, fulltextResp, mode, compareReason)
		observeSearchRequest(fulltextResp.Meta.Engine, mode, requestResult(fulltextResp, compareReason), time.Since(startedAt))
		return fulltextResp, nil
	}
}

func (s *Service) Suggest(prefix string, limit int) []string {
	artifact := s.currentArtifact()
	if artifact == nil || artifact.snapshot == nil {
		return nil
	}
	return artifact.snapshot.Suggest(prefix, limit)
}

func (s *Service) Metadata() BuildMetadata {
	artifact := s.currentArtifact()
	if artifact == nil {
		return BuildMetadata{}
	}
	return artifact.metadata.clone()
}

func (s *Service) searchNewEngine(ctx context.Context, req Request) (Response, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(s.cfg.QueryTimeoutMs)*time.Millisecond)
	defer cancel()

	artifact, err := s.ensureActiveArtifact(timeoutCtx)
	if err != nil {
		return Response{}, err
	}
	response, err := s.runSearch(timeoutCtx, artifact.snapshot, req, s.cfg, s.now())
	if err != nil {
		return Response{}, err
	}
	response.Meta.Engine = EngineHybrid
	response.Meta.Build = artifact.metadata.clone()
	if s.cfg.BatchOne.Explain {
		s.logExplain(timeoutCtx, req, response)
	}
	return response, nil
}

func (s *Service) ensureActiveArtifact(ctx context.Context) (*indexArtifact, error) {
	if artifact, err := s.loadActiveArtifact(); artifact != nil || err != nil {
		return artifact, err
	}

	s.buildMu.Lock()
	defer s.buildMu.Unlock()

	if artifact, err := s.loadActiveArtifact(); artifact != nil || err != nil {
		return artifact, err
	}

	artifact, err := s.buildArtifact(ctx)
	if err != nil {
		return nil, err
	}
	if err := validateArtifact(artifact); err != nil {
		return nil, err
	}
	s.publishArtifact(artifact)
	return artifact, nil
}

func (s *Service) buildArtifact(ctx context.Context) (*indexArtifact, error) {
	buildCtx, cancel := context.WithTimeout(ctx, time.Duration(s.cfg.BuildTimeoutMs)*time.Millisecond)
	defer cancel()

	docs, err := s.loader.ListSearchDocuments(buildCtx, s.cfg.MaxDocuments+1)
	if err != nil {
		observeBuildFailure()
		return nil, fmt.Errorf("%w: %v", errIndexBuildFailed, err)
	}
	if len(docs) > s.cfg.MaxDocuments {
		observeBuildFailure()
		return nil, fmt.Errorf("%w: %d documents exceed max %d", errIndexBuildFailed, len(docs), s.cfg.MaxDocuments)
	}
	snapshot, err := BuildSnapshot(buildCtx, docs, s.cfg, s.lexicon, s.synonyms)
	if err != nil {
		observeBuildFailure()
		return nil, fmt.Errorf("%w: %v", errIndexBuildFailed, err)
	}
	artifact := &indexArtifact{
		snapshot: snapshot,
		metadata: snapshot.Metadata(),
	}
	logx.WithContext(ctx).Infof(
		"search index candidate built version=%s checksum=%s docs=%d terms=%d workers=%d duration=%s segments=%d trie=%t lexicon=%d signature=%s",
		artifact.metadata.Version,
		artifact.metadata.Checksum,
		artifact.metadata.DocumentCount,
		artifact.metadata.TermCount,
		artifact.metadata.WorkerCount,
		artifact.metadata.BuildDuration,
		artifact.metadata.SegmentCount,
		artifact.metadata.TrieEnabled,
		artifact.metadata.LexiconSize,
		artifact.metadata.Signature,
	)
	return artifact, nil
}

func (s *Service) searchFulltext(ctx context.Context, req Request, fallbackReason string) (Response, error) {
	papers, total, err := s.fulltext.Search(ctx, req.Query, req.Discipline, req.Page, req.PageSize)
	if err != nil {
		return Response{}, err
	}
	return Response{
		Papers:      papers,
		Total:       total,
		Suggestions: s.suggestionsForRequest(ctx, req),
		Meta: ResponseMeta{
			Engine:         EngineFulltext,
			UsedFallback:   fallbackReason != "",
			FallbackReason: fallbackReason,
			Build:          s.Metadata(),
		},
	}, nil
}

func (s *Service) logShadowComparison(ctx context.Context, req Request, fulltextResp, hybridResp Response) {
	deltaKind := classifyShadowDelta(req, fulltextResp, hybridResp)
	observeShadowDelta(deltaKind)
	logx.WithContext(ctx).Infof(
		"search shadow compare request_id=%s mode=shadow index_version=%s normalized_query=%q tokens=%v fulltext=%v hybrid=%v delta_kind=%s fallback_reason=%s",
		searchRequestID(ctx),
		hybridResp.Meta.Build.Version,
		hybridResp.QueryAnalysis.Raw,
		hybridResp.QueryAnalysis.ExpandedTerms,
		resultIDs(fulltextResp.Papers),
		resultIDs(hybridResp.Papers),
		deltaKind,
		hybridResp.Meta.FallbackReason,
	)
}

func (s *Service) logExplain(ctx context.Context, req Request, resp Response) {
	if len(resp.Explains) == 0 {
		logx.WithContext(ctx).Infof(
			"search explain request_id=%s engine=%s mode=%s index_version=%s normalized_query=%q tokens=%v top_ids=%v fallback_reason=%s no_hits=true",
			searchRequestID(ctx),
			resp.Meta.Engine,
			requestMode(resp.Meta.Engine, false),
			resp.Meta.Build.Version,
			resp.QueryAnalysis.Raw,
			resp.QueryAnalysis.ExpandedTerms,
			resultIDs(resp.Papers),
			resp.Meta.FallbackReason,
		)
		return
	}
	maxExplains := len(resp.Explains)
	if maxExplains > 3 {
		maxExplains = 3
	}
	for _, explain := range resp.Explains[:maxExplains] {
		logx.WithContext(ctx).Infof(
			"search explain request_id=%s engine=%s mode=%s index_version=%s normalized_query=%q tokens=%v top_ids=%v fallback_reason=%s doc=%d ik=%v jieba=%v expanded=%v matched=%v bm25=%.4f freshness=%.4f quality=%.4f final=%.4f",
			searchRequestID(ctx),
			resp.Meta.Engine,
			requestMode(resp.Meta.Engine, false),
			resp.Meta.Build.Version,
			resp.QueryAnalysis.Raw,
			resp.QueryAnalysis.ExpandedTerms,
			resultIDs(resp.Papers),
			resp.Meta.FallbackReason,
			explain.DocumentID,
			resp.QueryAnalysis.IKTokens,
			resp.QueryAnalysis.JiebaTokens,
			resp.QueryAnalysis.ExpandedTerms,
			explain.MatchedTerms,
			explain.BM25Score,
			explain.FreshnessScore,
			explain.QualityScore,
			explain.FinalScore,
		)
	}
}

func (s *Service) resolveEngine(override string) string {
	switch strings.ToLower(strings.TrimSpace(override)) {
	case EngineFulltext:
		return EngineFulltext
	case EngineHybrid:
		return EngineHybrid
	default:
		return s.cfg.DefaultEngine
	}
}

func (s *Service) resolveShadowCompare(engine string, requested bool) bool {
	if engine == EngineHybrid {
		return false
	}
	if requested {
		return true
	}
	return s.cfg.ShadowCompare
}

func (s *Service) suggestionsForRequest(ctx context.Context, req Request) []string {
	if req.SuggestionLimit <= 0 || !s.cfg.BatchTwo.TrieEnabled || !s.cfg.BatchOne.Enabled || s.loader == nil {
		return nil
	}
	artifact, err := s.ensureActiveArtifact(ctx)
	if err != nil {
		logx.WithContext(ctx).Infof(
			"search suggestions unavailable request_id=%s normalized_query=%q reason=%s index_version=%s",
			searchRequestID(ctx),
			normalizeText(req.Query),
			classifyFallbackReason(err),
			s.Metadata().Version,
		)
		return nil
	}
	return artifact.snapshot.Suggest(req.Query, req.SuggestionLimit)
}

func (s *Service) logSearchOutcome(ctx context.Context, req Request, resp Response, mode, compareReason string) {
	fields := s.buildSearchLogFields(ctx, req, resp, mode, compareReason)
	logx.WithContext(ctx).Infof("search answer path %s", formatSearchLog(fields))
}

func (s *Service) currentArtifact() *indexArtifact {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.active != nil {
		return s.active
	}
	return s.lastSuccessful
}

func (s *Service) loadActiveArtifact() (*indexArtifact, error) {
	s.mu.RLock()
	artifact := s.active
	s.mu.RUnlock()
	if artifact == nil {
		return nil, nil
	}
	if err := validateArtifact(artifact); err != nil {
		s.invalidateActiveArtifact()
		return nil, err
	}
	return artifact, nil
}

func (s *Service) publishArtifact(artifact *indexArtifact) {
	artifact.metadata.PublishedAt = s.now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active = artifact
	s.lastSuccessful = artifact.cloneForCache()
	observeBuildSuccess(artifact.metadata)
}

func (s *Service) invalidateActiveArtifact() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active = nil
}

func (a *indexArtifact) cloneForCache() *indexArtifact {
	if a == nil {
		return nil
	}
	return &indexArtifact{
		snapshot: a.snapshot,
		metadata: a.metadata.clone(),
	}
}

func validateArtifact(artifact *indexArtifact) error {
	if artifact == nil {
		return fmt.Errorf("search artifact is nil")
	}
	if artifact.snapshot == nil {
		return fmt.Errorf("%w: snapshot unavailable", errSegmentLoadFailed)
	}

	meta := artifact.metadata
	if meta.Version == "" {
		return fmt.Errorf("%w: active search artifact missing version", errMissingVersion)
	}
	if meta.Checksum == "" {
		return fmt.Errorf("%w: active search artifact missing checksum", errSegmentInvalid)
	}
	if meta.Signature == "" && meta.DocumentCount > 0 {
		return fmt.Errorf("%w: active search artifact missing signature", errSegmentInvalid)
	}
	if meta.DocumentCount != len(artifact.snapshot.documents) {
		return fmt.Errorf("%w: document count mismatch meta=%d actual=%d", errSegmentInvalid, meta.DocumentCount, len(artifact.snapshot.documents))
	}
	if meta.TermCount != len(artifact.snapshot.postings) {
		return fmt.Errorf("%w: term count mismatch meta=%d actual=%d", errSegmentInvalid, meta.TermCount, len(artifact.snapshot.postings))
	}
	if meta.Signature != artifact.snapshot.metadata.Signature {
		return fmt.Errorf("%w: signature mismatch", errSegmentInvalid)
	}
	if meta.SegmentCount != len(meta.Segments) {
		return fmt.Errorf("%w: segment count mismatch", errSegmentInvalid)
	}
	if meta.DocumentCount > 0 && len(meta.Segments) == 0 {
		return fmt.Errorf("%w: active search artifact missing segments", errSegmentInvalid)
	}
	totalDocs := 0
	for _, segment := range meta.Segments {
		if segment.Name == "" {
			return fmt.Errorf("%w: search artifact segment missing name", errSegmentInvalid)
		}
		if segment.Checksum == "" {
			return fmt.Errorf("%w: search artifact segment missing checksum", errSegmentInvalid)
		}
		totalDocs += segment.DocumentCount
	}
	if meta.DocumentCount > 0 && totalDocs != meta.DocumentCount {
		return fmt.Errorf("%w: segment coverage mismatch meta=%d segments=%d", errSegmentInvalid, meta.DocumentCount, totalDocs)
	}
	if metadataChecksum(meta) != meta.Checksum {
		return fmt.Errorf("%w: checksum mismatch", errSegmentInvalid)
	}
	return nil
}

func classifyFallbackReason(err error) string {
	switch {
	case err == nil:
		return FallbackReasonNone
	case errors.Is(err, context.DeadlineExceeded):
		return FallbackReasonQueryTimeout
	case errors.Is(err, errMissingVersion):
		return FallbackReasonMissingIndexVersion
	case errors.Is(err, errSegmentLoadFailed):
		return FallbackReasonSegmentLoadFailed
	case errors.Is(err, errSegmentInvalid):
		return FallbackReasonSegmentValidation
	case errors.Is(err, errQueryParseFailed):
		return FallbackReasonQueryParseFailed
	case errors.Is(err, errRankerFailed):
		return FallbackReasonRankerError
	case errors.Is(err, errIndexBuildFailed):
		return FallbackReasonIndexBuildFailed
	default:
		return FallbackReasonEngineError
	}
}
