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
	cfg      Config
	loader   DocumentLoader
	fulltext FulltextSearcher
	now      func() time.Time
	lexicon  []string
	synonyms synonymMap
	mu       sync.RWMutex
	snapshot *Snapshot
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

	switch s.cfg.DefaultEngine {
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
		if s.cfg.ShadowCompare {
			if shadowResp, shadowErr := s.searchNewEngine(ctx, req); shadowErr != nil {
				logx.WithContext(ctx).Infof("search shadow compare fallback query=%q reason=%s", req.Query, shadowErr.Error())
			} else {
				s.logShadowComparison(ctx, req, fulltextResp, shadowResp)
			}
			fulltextResp.Meta.ShadowCompared = true
		}
		return fulltextResp, nil
	}
}

func (s *Service) Suggest(prefix string, limit int) []string {
	s.mu.RLock()
	snapshot := s.snapshot
	s.mu.RUnlock()
	if snapshot == nil {
		return nil
	}
	return snapshot.Suggest(prefix, limit)
}

func (s *Service) Metadata() BuildMetadata {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.snapshot == nil {
		return BuildMetadata{}
	}
	return s.snapshot.Metadata()
}

func (s *Service) searchNewEngine(ctx context.Context, req Request) (Response, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(s.cfg.QueryTimeoutMs)*time.Millisecond)
	defer cancel()

	snapshot, err := s.ensureSnapshot(timeoutCtx)
	if err != nil {
		return Response{}, err
	}
	response := snapshot.Search(req, s.cfg, s.now())
	response.Meta.Engine = EngineHybrid
	response.Meta.Build = snapshot.Metadata()
	if s.cfg.BatchOne.Explain {
		s.logExplain(timeoutCtx, req, response)
	}
	return response, nil
}

func (s *Service) ensureSnapshot(ctx context.Context) (*Snapshot, error) {
	s.mu.RLock()
	if s.snapshot != nil {
		defer s.mu.RUnlock()
		return s.snapshot, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.snapshot != nil {
		return s.snapshot, nil
	}

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
	s.snapshot = snapshot
	logx.WithContext(ctx).Infof(
		"search index built docs=%d terms=%d workers=%d duration=%s trie=%t lexicon=%d signature=%s",
		snapshot.metadata.DocumentCount,
		snapshot.metadata.TermCount,
		snapshot.metadata.WorkerCount,
		snapshot.metadata.BuildDuration,
		snapshot.metadata.TrieEnabled,
		snapshot.metadata.LexiconSize,
		snapshot.metadata.Signature,
	)
	return snapshot, nil
}

func (s *Service) searchFulltext(ctx context.Context, req Request, fallbackReason string) (Response, error) {
	papers, total, err := s.fulltext.Search(ctx, req.Query, req.Discipline, req.Page, req.PageSize)
	if err != nil {
		return Response{}, err
	}
	return Response{
		Papers: papers,
		Total:  total,
		Meta: ResponseMeta{
			Engine:         EngineFulltext,
			UsedFallback:   fallbackReason != "",
			FallbackReason: fallbackReason,
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
