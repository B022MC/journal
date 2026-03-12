package search

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"journal/model"
)

const (
	bm25K1 = 1.2
	bm25B  = 0.75
)

type Request struct {
	Query           string
	Discipline      string
	Page            int
	PageSize        int
	Sort            string
	Engine          string
	Shadow          bool
	SuggestionLimit int
}

type Response struct {
	Papers        []*model.Paper
	Total         int64
	Suggestions   []string
	QueryAnalysis QueryAnalysis
	Explains      []ExplainMatch
	Meta          ResponseMeta
}

type ResponseMeta struct {
	Engine         string
	UsedFallback   bool
	FallbackReason string
	ShadowCompared bool
	Build          BuildMetadata
}

type BuildMetadata struct {
	DocumentCount int
	TermCount     int
	WorkerCount   int
	BuildDuration time.Duration
	Signature     string
	TrieEnabled   bool
	LexiconSize   int
}

type ExplainMatch struct {
	DocumentID     int64
	MatchedTerms   []string
	TermDetails    []ExplainTerm
	BM25Score      float64
	FreshnessScore float64
	QualityScore   float64
	FinalScore     float64
}

type ExplainTerm struct {
	Term      string
	TF        int
	DF        int
	IDF       float64
	TermScore float64
}

type Snapshot struct {
	documents      map[int64]*model.Paper
	docLengths     map[int64]int
	docTokenCounts map[int64]map[string]int
	postings       map[string][]posting
	termDocFreq    map[string]int
	avgDocLength   float64
	metadata       BuildMetadata
	trie           *trieNode
	lexicon        []string
	synonyms       synonymMap
}

type posting struct {
	DocumentID int64
	TF         int
}

type buildDocResult struct {
	paper       *model.Paper
	tokenCounts map[string]int
	docLength   int
	err         error
}

func BuildSnapshot(ctx context.Context, docs []*model.Paper, cfg Config, lexicon []string, synonyms synonymMap) (*Snapshot, error) {
	startedAt := time.Now()
	if len(docs) == 0 {
		return &Snapshot{
			documents:      map[int64]*model.Paper{},
			docLengths:     map[int64]int{},
			docTokenCounts: map[int64]map[string]int{},
			postings:       map[string][]posting{},
			termDocFreq:    map[string]int{},
			metadata: BuildMetadata{
				WorkerCount: cfg.BatchOne.WorkerCount,
				LexiconSize: len(lexicon),
				TrieEnabled: cfg.BatchTwo.TrieEnabled,
			},
			lexicon:  append([]string{}, lexicon...),
			synonyms: synonyms,
		}, nil
	}

	workerCount := cfg.BatchOne.WorkerCount
	if workerCount > len(docs) {
		workerCount = len(docs)
	}
	if workerCount <= 0 {
		workerCount = 1
	}

	jobs := make(chan *model.Paper)
	results := make(chan buildDocResult, len(docs))

	var workers sync.WaitGroup
	for range workerCount {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for paper := range jobs {
				select {
				case <-ctx.Done():
					results <- buildDocResult{err: ctx.Err()}
					return
				default:
				}
				tokenCounts, docLength := buildDocumentTermCounts(paper, cfg.BatchOne, lexicon)
				results <- buildDocResult{
					paper:       paper,
					tokenCounts: tokenCounts,
					docLength:   docLength,
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, doc := range docs {
			select {
			case <-ctx.Done():
				return
			case jobs <- doc:
			}
		}
	}()

	go func() {
		workers.Wait()
		close(results)
	}()

	snapshot := &Snapshot{
		documents:      make(map[int64]*model.Paper, len(docs)),
		docLengths:     make(map[int64]int, len(docs)),
		docTokenCounts: make(map[int64]map[string]int, len(docs)),
		postings:       map[string][]posting{},
		termDocFreq:    map[string]int{},
		lexicon:        append([]string{}, lexicon...),
		synonyms:       synonyms,
	}

	totalDocLength := 0
	for result := range results {
		if result.err != nil {
			return nil, result.err
		}
		if result.paper == nil {
			continue
		}
		snapshot.documents[result.paper.Id] = result.paper
		snapshot.docLengths[result.paper.Id] = result.docLength
		snapshot.docTokenCounts[result.paper.Id] = result.tokenCounts
		totalDocLength += result.docLength
		for term, tf := range result.tokenCounts {
			snapshot.postings[term] = append(snapshot.postings[term], posting{
				DocumentID: result.paper.Id,
				TF:         tf,
			})
			snapshot.termDocFreq[term]++
		}
	}

	if len(snapshot.documents) > 0 {
		snapshot.avgDocLength = float64(totalDocLength) / float64(len(snapshot.documents))
	}
	for term := range snapshot.postings {
		sort.Slice(snapshot.postings[term], func(i, j int) bool {
			return snapshot.postings[term][i].DocumentID < snapshot.postings[term][j].DocumentID
		})
	}
	if cfg.BatchTwo.TrieEnabled {
		snapshot.trie = newTrie()
		for term, df := range snapshot.termDocFreq {
			snapshot.trie.Insert(term, df)
		}
	}
	snapshot.metadata = BuildMetadata{
		DocumentCount: len(snapshot.documents),
		TermCount:     len(snapshot.postings),
		WorkerCount:   workerCount,
		BuildDuration: time.Since(startedAt),
		Signature:     snapshot.signature(),
		TrieEnabled:   cfg.BatchTwo.TrieEnabled,
		LexiconSize:   len(lexicon),
	}
	return snapshot, nil
}

func buildDocumentTermCounts(paper *model.Paper, batchOne BatchOneConfig, lexicon []string) (map[string]int, int) {
	tokenCounts := map[string]int{}
	addFieldTokens := func(value string, weight int) {
		for _, token := range tokenizeDocument(value, batchOne, lexicon) {
			tokenCounts[token] += weight
		}
	}
	addFieldTokens(paper.Title, 3)
	addFieldTokens(paper.TitleEn, 2)
	addFieldTokens(paper.Keywords, 2)
	addFieldTokens(paper.Abstract, 1)
	addFieldTokens(paper.GetAbstractEn(), 1)
	docLength := 0
	for _, count := range tokenCounts {
		docLength += count
	}
	return tokenCounts, docLength
}

func (s *Snapshot) Search(req Request, cfg Config, now time.Time) Response {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	suggestionLimit := req.SuggestionLimit
	if suggestionLimit <= 0 {
		suggestionLimit = 6
	}

	analysis := analyzeQuery(req.Query, cfg.BatchOne, s.lexicon, s.synonyms, cfg.BatchTwo.SynonymEnabled)
	accumulators := map[int64]*scoreAccumulator{}
	totalDocs := float64(len(s.documents))

	for _, term := range analysis.ExpandedTerms {
		postings := s.postings[term]
		if len(postings) == 0 {
			continue
		}
		df := len(postings)
		idf := math.Log(1 + (totalDocs-float64(df)+0.5)/(float64(df)+0.5))
		for _, posting := range postings {
			paper := s.documents[posting.DocumentID]
			if paper == nil {
				continue
			}
			if req.Discipline != "" && paper.Discipline != req.Discipline {
				continue
			}
			docLength := float64(s.docLengths[posting.DocumentID])
			score := idf * (float64(posting.TF) * (bm25K1 + 1)) / (float64(posting.TF) + bm25K1*(1-bm25B+bm25B*docLength/max(1, s.avgDocLength)))
			acc := accumulators[posting.DocumentID]
			if acc == nil {
				acc = &scoreAccumulator{
					paper:        paper,
					matchedTerms: map[string]ExplainTerm{},
				}
				accumulators[posting.DocumentID] = acc
			}
			acc.bm25 += score
			explain := acc.matchedTerms[term]
			explain.Term = term
			explain.TF += posting.TF
			explain.DF = df
			explain.IDF = idf
			explain.TermScore += score
			acc.matchedTerms[term] = explain
		}
	}

	results := make([]searchResult, 0, len(accumulators))
	maxBM25 := 0.0
	maxQuality := 0.0
	for _, acc := range accumulators {
		if acc.bm25 > maxBM25 {
			maxBM25 = acc.bm25
		}
		if acc.paper.ShitScore > maxQuality {
			maxQuality = acc.paper.ShitScore
		}
	}

	for _, acc := range accumulators {
		freshness := freshnessScore(acc.paper, now)
		quality := normalizePositive(acc.paper.ShitScore, maxQuality)
		final := acc.bm25
		if cfg.BatchTwo.FusionEnabled {
			final = cfg.BatchTwo.FusionBM25Weight*normalizePositive(acc.bm25, maxBM25) +
				cfg.BatchTwo.FusionFreshnessWeight*freshness +
				cfg.BatchTwo.FusionQualityWeight*quality
		}
		results = append(results, searchResult{
			paper: acc.paper,
			explain: ExplainMatch{
				DocumentID:     acc.paper.Id,
				MatchedTerms:   sortedExplainTerms(acc.matchedTerms),
				TermDetails:    sortedExplainDetails(acc.matchedTerms),
				BM25Score:      round4(acc.bm25),
				FreshnessScore: round4(freshness),
				QualityScore:   round4(quality),
				FinalScore:     round4(final),
			},
			finalScore: final,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return compareSearchResults(results[i], results[j], req.Sort)
	})

	total := int64(len(results))
	offset := (page - 1) * pageSize
	if offset > len(results) {
		offset = len(results)
	}
	end := offset + pageSize
	if end > len(results) {
		end = len(results)
	}

	response := Response{
		Papers:        make([]*model.Paper, 0, end-offset),
		Explains:      make([]ExplainMatch, 0, end-offset),
		Suggestions:   s.Suggest(req.Query, suggestionLimit),
		Total:         total,
		QueryAnalysis: analysis,
		Meta: ResponseMeta{
			Engine: EngineHybrid,
			Build:  s.metadata,
		},
	}
	for _, result := range results[offset:end] {
		response.Papers = append(response.Papers, result.paper)
		response.Explains = append(response.Explains, result.explain)
	}
	return response
}

func (s *Snapshot) Suggest(prefix string, limit int) []string {
	if s.trie == nil || strings.TrimSpace(prefix) == "" || limit <= 0 {
		return nil
	}
	return s.trie.Suggest(normalizeText(prefix), limit)
}

func (s *Snapshot) Metadata() BuildMetadata {
	return s.metadata
}

type scoreAccumulator struct {
	paper        *model.Paper
	bm25         float64
	matchedTerms map[string]ExplainTerm
}

type searchResult struct {
	paper      *model.Paper
	explain    ExplainMatch
	finalScore float64
}

func compareSearchResults(left, right searchResult, sortMode string) bool {
	switch normalizeSort(sortMode) {
	case "newest":
		if left.paper.UpdatedAt.Equal(right.paper.UpdatedAt) {
			if left.finalScore == right.finalScore {
				return left.paper.Id > right.paper.Id
			}
			return left.finalScore > right.finalScore
		}
		return left.paper.UpdatedAt.After(right.paper.UpdatedAt)
	case "quality":
		if left.paper.ShitScore == right.paper.ShitScore {
			if left.finalScore == right.finalScore {
				return left.paper.Id > right.paper.Id
			}
			return left.finalScore > right.finalScore
		}
		return left.paper.ShitScore > right.paper.ShitScore
	default:
		if left.finalScore == right.finalScore {
			if left.paper.UpdatedAt.Equal(right.paper.UpdatedAt) {
				return left.paper.Id > right.paper.Id
			}
			return left.paper.UpdatedAt.After(right.paper.UpdatedAt)
		}
		return left.finalScore > right.finalScore
	}
}

func normalizeSort(sortMode string) string {
	switch strings.ToLower(strings.TrimSpace(sortMode)) {
	case "newest", "quality":
		return strings.ToLower(strings.TrimSpace(sortMode))
	default:
		return "relevance"
	}
}

func sortedExplainTerms(details map[string]ExplainTerm) []string {
	keys := make([]string, 0, len(details))
	for term := range details {
		keys = append(keys, term)
	}
	sort.Strings(keys)
	return keys
}

func sortedExplainDetails(details map[string]ExplainTerm) []ExplainTerm {
	terms := make([]ExplainTerm, 0, len(details))
	for _, term := range sortedExplainTerms(details) {
		terms = append(terms, details[term])
	}
	return terms
}

func freshnessScore(paper *model.Paper, now time.Time) float64 {
	if paper == nil {
		return 0
	}
	reference := paper.UpdatedAt
	if reference.IsZero() {
		reference = paper.CreatedAt
	}
	if reference.IsZero() {
		return 0
	}
	days := now.Sub(reference).Hours() / 24
	if days < 0 {
		days = 0
	}
	return round4(1 / (1 + days/30))
}

func normalizePositive(value, maxValue float64) float64 {
	if value <= 0 || maxValue <= 0 {
		return 0
	}
	return value / maxValue
}

func round4(value float64) float64 {
	return math.Round(value*10000) / 10000
}

func (s *Snapshot) signature() string {
	docIDs := make([]int64, 0, len(s.documents))
	for id := range s.documents {
		docIDs = append(docIDs, id)
	}
	sort.Slice(docIDs, func(i, j int) bool { return docIDs[i] < docIDs[j] })

	hash := sha1.New()
	for _, id := range docIDs {
		_, _ = hash.Write([]byte(fmt.Sprintf("%d|", id)))
		terms := make([]string, 0, len(s.docTokenCounts[id]))
		for term := range s.docTokenCounts[id] {
			terms = append(terms, term)
		}
		sort.Strings(terms)
		for _, term := range terms {
			_, _ = hash.Write([]byte(fmt.Sprintf("%s:%d;", term, s.docTokenCounts[id][term])))
		}
	}
	return hex.EncodeToString(hash.Sum(nil))
}

type trieNode struct {
	children map[rune]*trieNode
	terminal bool
	term     string
	weight   int
}

func newTrie() *trieNode {
	return &trieNode{children: map[rune]*trieNode{}}
}

func (n *trieNode) Insert(term string, weight int) {
	node := n
	for _, r := range term {
		if node.children[r] == nil {
			node.children[r] = newTrie()
		}
		node = node.children[r]
	}
	node.terminal = true
	node.term = term
	node.weight = weight
}

func (n *trieNode) Suggest(prefix string, limit int) []string {
	node := n
	for _, r := range prefix {
		node = node.children[r]
		if node == nil {
			return nil
		}
	}

	type candidate struct {
		term   string
		weight int
	}
	candidates := make([]candidate, 0, limit)
	var walk func(current *trieNode)
	walk = func(current *trieNode) {
		if current == nil {
			return
		}
		if current.terminal {
			candidates = append(candidates, candidate{
				term:   current.term,
				weight: current.weight,
			})
		}
		keys := make([]rune, 0, len(current.children))
		for r := range current.children {
			keys = append(keys, r)
		}
		sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
		for _, key := range keys {
			walk(current.children[key])
		}
	}
	walk(node)

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].weight == candidates[j].weight {
			return candidates[i].term < candidates[j].term
		}
		return candidates[i].weight > candidates[j].weight
	})
	if limit > len(candidates) {
		limit = len(candidates)
	}
	suggestions := make([]string, 0, limit)
	for _, item := range candidates[:limit] {
		suggestions = append(suggestions, item.term)
	}
	return suggestions
}
