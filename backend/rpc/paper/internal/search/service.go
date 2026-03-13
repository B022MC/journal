package search

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"journal/model"

	"github.com/zeromicro/go-zero/core/logx"
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
		now:      time.Now,
		lexicon:  lexicon,
		synonyms: synonyms,
	}
}

func (s *Service) Search(ctx context.Context, req Request) (Response, error) {
	if s.fulltext == nil {
		return Response{}, fmt.Errorf("fulltext searcher is not configured")
	}
	if strings.TrimSpace(req.Query) == "" {
		return s.searchFulltext(ctx, req, "empty_query")
	}
	if !s.cfg.BatchOne.Enabled || s.loader == nil {
		return s.searchFulltext(ctx, req, "batch_one_disabled")
	}

	engine := s.resolveEngine(req.Engine)
	shadowCompare := s.resolveShadowCompare(engine, req.Shadow)

	switch engine {
	case EngineHybrid:
		response, err := s.searchNewEngine(ctx, req)
		if err != nil {
			return s.searchFulltext(ctx, req, "hybrid_error:"+err.Error())
		}
		return response, nil
	default:
		fulltextResp, err := s.searchFulltext(ctx, req, "")
		if err != nil {
			return Response{}, err
		}
		if shadowCompare {
			if shadowResp, shadowErr := s.searchNewEngine(ctx, req); shadowErr != nil {
				logx.WithContext(ctx).Infof("search shadow compare fallback query=%q reason=%s", req.Query, shadowErr.Error())
			} else {
				s.logShadowComparison(ctx, req, fulltextResp, shadowResp)
				if len(fulltextResp.Suggestions) == 0 {
					fulltextResp.Suggestions = shadowResp.Suggestions
				}
			}
			fulltextResp.Meta.ShadowCompared = true
		}
		if len(fulltextResp.Suggestions) == 0 {
			fulltextResp.Suggestions = s.suggestionsForRequest(ctx, req)
		}
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
	response := artifact.snapshot.Search(req, s.cfg, s.now())
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
		return nil, err
	}
	if len(docs) > s.cfg.MaxDocuments {
		return nil, fmt.Errorf("search index aborted: %d documents exceed max %d", len(docs), s.cfg.MaxDocuments)
	}
	snapshot, err := BuildSnapshot(buildCtx, docs, s.cfg, s.lexicon, s.synonyms)
	if err != nil {
		return nil, err
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
	fulltextIDs := make([]int64, 0, len(fulltextResp.Papers))
	hybridIDs := make([]int64, 0, len(hybridResp.Papers))
	for _, paper := range fulltextResp.Papers {
		fulltextIDs = append(fulltextIDs, paper.Id)
	}
	for _, paper := range hybridResp.Papers {
		hybridIDs = append(hybridIDs, paper.Id)
	}
	logx.WithContext(ctx).Infof(
		"search shadow compare query=%q fulltext=%v hybrid=%v expanded=%v",
		req.Query,
		fulltextIDs,
		hybridIDs,
		hybridResp.QueryAnalysis.ExpandedTerms,
	)
}

func (s *Service) logExplain(ctx context.Context, req Request, resp Response) {
	if len(resp.Explains) == 0 {
		logx.WithContext(ctx).Infof("search explain query=%q no_hits expanded=%v", req.Query, resp.QueryAnalysis.ExpandedTerms)
		return
	}
	maxExplains := len(resp.Explains)
	if maxExplains > 3 {
		maxExplains = 3
	}
	for _, explain := range resp.Explains[:maxExplains] {
		logx.WithContext(ctx).Infof(
			"search explain query=%q doc=%d ik=%v jieba=%v expanded=%v matched=%v bm25=%.4f freshness=%.4f quality=%.4f final=%.4f",
			req.Query,
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
		logx.WithContext(ctx).Infof("search suggestions unavailable query=%q reason=%s", req.Query, err.Error())
		return nil
	}
	return artifact.snapshot.Suggest(req.Query, req.SuggestionLimit)
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
		return fmt.Errorf("search artifact load failed: snapshot unavailable")
	}

	meta := artifact.metadata
	if meta.Version == "" {
		return fmt.Errorf("active search artifact missing version")
	}
	if meta.Checksum == "" {
		return fmt.Errorf("active search artifact missing checksum")
	}
	if meta.Signature == "" && meta.DocumentCount > 0 {
		return fmt.Errorf("active search artifact missing signature")
	}
	if meta.DocumentCount != len(artifact.snapshot.documents) {
		return fmt.Errorf("search artifact document count mismatch: meta=%d actual=%d", meta.DocumentCount, len(artifact.snapshot.documents))
	}
	if meta.TermCount != len(artifact.snapshot.postings) {
		return fmt.Errorf("search artifact term count mismatch: meta=%d actual=%d", meta.TermCount, len(artifact.snapshot.postings))
	}
	if meta.Signature != artifact.snapshot.metadata.Signature {
		return fmt.Errorf("search artifact signature mismatch")
	}
	if meta.SegmentCount != len(meta.Segments) {
		return fmt.Errorf("search artifact segment count mismatch")
	}
	if meta.DocumentCount > 0 && len(meta.Segments) == 0 {
		return fmt.Errorf("active search artifact missing segments")
	}
	totalDocs := 0
	for _, segment := range meta.Segments {
		if segment.Name == "" {
			return fmt.Errorf("search artifact segment missing name")
		}
		if segment.Checksum == "" {
			return fmt.Errorf("search artifact segment missing checksum")
		}
		totalDocs += segment.DocumentCount
	}
	if meta.DocumentCount > 0 && totalDocs != meta.DocumentCount {
		return fmt.Errorf("search artifact segment coverage mismatch: meta=%d segments=%d", meta.DocumentCount, totalDocs)
	}
	if metadataChecksum(meta) != meta.Checksum {
		return fmt.Errorf("search artifact checksum mismatch")
	}
	return nil
}
