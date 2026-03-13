package search

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

var latinTokenPattern = regexp.MustCompile(`[\p{L}\p{N}]+`)

var builtInLexicon = []string{
	"人工智能",
	"机器学习",
	"深度学习",
	"神经网络",
	"推荐系统",
	"信息检索",
	"量子计算",
}

type QueryAnalysis struct {
	Raw           string
	IKTokens      []string
	JiebaTokens   []string
	LexiconTokens []string
	ExpandedTerms []string
}

func analyzeQuery(raw string, batchOne BatchOneConfig, lexicon []string, synonyms synonymMap, enableSynonyms bool) (QueryAnalysis, error) {
	if !utf8.ValidString(raw) {
		return QueryAnalysis{}, fmt.Errorf("%w: invalid utf8 query", errQueryParseFailed)
	}
	query := normalizeText(raw)
	analysis := QueryAnalysis{Raw: query}
	if batchOne.EnableIK {
		analysis.IKTokens = tokenizeIK(query)
	}
	if batchOne.EnableJieba {
		analysis.JiebaTokens = tokenizeJieba(query)
	}
	analysis.LexiconTokens = lexiconTokens(query, lexicon)
	base := make([]string, 0, len(analysis.IKTokens)+len(analysis.JiebaTokens)+len(analysis.LexiconTokens))
	base = append(base, analysis.IKTokens...)
	base = append(base, analysis.JiebaTokens...)
	base = append(base, analysis.LexiconTokens...)
	analysis.ExpandedTerms = uniquePreserveOrder(base)
	if enableSynonyms {
		for _, token := range analysis.ExpandedTerms {
			analysis.ExpandedTerms = append(analysis.ExpandedTerms, synonyms.expand(token)...)
		}
		analysis.ExpandedTerms = uniquePreserveOrder(analysis.ExpandedTerms)
	}
	if query != "" && len(analysis.ExpandedTerms) == 0 {
		return QueryAnalysis{}, fmt.Errorf("%w: no searchable tokens", errQueryParseFailed)
	}
	return analysis, nil
}

func tokenizeDocument(raw string, batchOne BatchOneConfig, lexicon []string) []string {
	text := normalizeText(raw)
	tokens := make([]string, 0, 32)
	if batchOne.EnableIK {
		tokens = append(tokens, tokenizeIK(text)...)
	}
	if batchOne.EnableJieba {
		tokens = append(tokens, tokenizeJieba(text)...)
	}
	tokens = append(tokens, lexiconTokens(text, lexicon)...)
	return filterEmptyTokens(tokens)
}

func normalizeText(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if raw == "" {
		return ""
	}
	return strings.Join(strings.Fields(raw), " ")
}

func tokenizeIK(text string) []string {
	if text == "" {
		return nil
	}
	tokens := latinTokenPattern.FindAllString(text, -1)
	for _, seq := range hanSequences(text) {
		runes := []rune(seq)
		for _, r := range runes {
			tokens = append(tokens, string(r))
		}
		if len(runes) == 1 {
			continue
		}
		for size := 2; size <= min(3, len(runes)); size++ {
			for i := 0; i+size <= len(runes); i++ {
				tokens = append(tokens, string(runes[i:i+size]))
			}
		}
	}
	return filterEmptyTokens(tokens)
}

func tokenizeJieba(text string) []string {
	if text == "" {
		return nil
	}
	tokens := latinTokenPattern.FindAllString(text, -1)
	for _, seq := range hanSequences(text) {
		runes := []rune(seq)
		tokens = append(tokens, seq)
		if len(runes) == 1 {
			continue
		}
		for size := min(4, len(runes)); size >= 2; size-- {
			for i := 0; i+size <= len(runes); i++ {
				tokens = append(tokens, string(runes[i:i+size]))
			}
		}
	}
	return filterEmptyTokens(tokens)
}

func lexiconTokens(text string, lexicon []string) []string {
	if text == "" || len(lexicon) == 0 {
		return nil
	}
	tokens := make([]string, 0, len(lexicon))
	for _, term := range lexicon {
		if term == "" {
			continue
		}
		if strings.Contains(text, term) {
			tokens = append(tokens, term)
		}
	}
	return filterEmptyTokens(tokens)
}

func loadLexicon(path string) ([]string, error) {
	terms := append([]string{}, builtInLexicon...)
	if strings.TrimSpace(path) == "" {
		slices.Sort(terms)
		return terms, nil
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		term := normalizeText(scanner.Text())
		if term == "" {
			continue
		}
		terms = append(terms, term)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return uniqueSorted(terms), nil
}

func hanSequences(text string) []string {
	var sequences []string
	var builder strings.Builder
	flush := func() {
		if builder.Len() == 0 {
			return
		}
		sequences = append(sequences, builder.String())
		builder.Reset()
	}

	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			builder.WriteRune(r)
			continue
		}
		flush()
	}
	flush()
	return sequences
}

func filterEmptyTokens(tokens []string) []string {
	filtered := make([]string, 0, len(tokens))
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		filtered = append(filtered, token)
	}
	return filtered
}

func uniquePreserveOrder(tokens []string) []string {
	seen := make(map[string]struct{}, len(tokens))
	out := make([]string, 0, len(tokens))
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		if _, ok := seen[token]; ok {
			continue
		}
		seen[token] = struct{}{}
		out = append(out, token)
	}
	return out
}

func uniqueSorted(tokens []string) []string {
	items := uniquePreserveOrder(tokens)
	slices.Sort(items)
	return items
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
