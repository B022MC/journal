package degradation

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"

	"journal/model"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/mozillazg/go-pinyin"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	CategoryAbuse        = "abuse"
	CategorySensitive    = "sensitive"
	CategorySpam         = "spam"
	CategoryPlagiarism   = "plagiarism"
	CategoryManipulation = "manipulation"

	MatchTypeKeyword = "keyword"
	MatchTypeRegex   = "regex"
	MatchTypePinyin  = "pinyin"
)

const (
	keywordRuleDataKey    = "keyword_filter:rules:data"
	keywordRuleVersionKey = "keyword_filter:rules:version"
)

var (
	ErrDuplicateRule    = errors.New("keyword rule already exists")
	ErrKeywordRuleEmpty = errors.New("keyword rule pattern is required")
	ErrKeywordRuleGone  = errors.New("keyword rule not found")
	ErrInvalidCategory  = errors.New("invalid keyword rule category")
	ErrInvalidMatchType = errors.New("invalid keyword rule match type")
	ErrInvalidPattern   = errors.New("invalid keyword rule pattern")
)

type KeywordMatch struct {
	RuleId    int64
	Pattern   string
	MatchType string
	Category  string
	MatchedBy string
}

type compiledKeywordRule struct {
	id                int64
	pattern           string
	matchType         string
	category          string
	lowerPattern      string
	normalizedPattern string
	regex             *regexp.Regexp
}

type compiledKeywordRuleSet struct {
	version string
	rules   []compiledKeywordRule
}

type KeywordFilter struct {
	mu     sync.RWMutex
	model  *model.KeywordRuleModel
	store  *redis.Redis
	cached compiledKeywordRuleSet
}

func NewKeywordFilter(ruleModel *model.KeywordRuleModel, store *redis.Redis) *KeywordFilter {
	return &KeywordFilter{
		model: ruleModel,
		store: store,
	}
}

func (f *KeywordFilter) Check(ctx context.Context, content string) (*KeywordMatch, error) {
	rules, err := f.loadRules(ctx)
	if err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return nil, nil
	}

	lowerContent := strings.ToLower(content)
	fullPinyin := ""
	initials := ""

	for _, rule := range rules {
		switch rule.matchType {
		case MatchTypeKeyword:
			if strings.Contains(lowerContent, rule.lowerPattern) {
				return &KeywordMatch{
					RuleId:    rule.id,
					Pattern:   rule.pattern,
					MatchType: rule.matchType,
					Category:  rule.category,
					MatchedBy: rule.pattern,
				}, nil
			}
		case MatchTypeRegex:
			matched := rule.regex.FindString(content)
			if matched != "" {
				return &KeywordMatch{
					RuleId:    rule.id,
					Pattern:   rule.pattern,
					MatchType: rule.matchType,
					Category:  rule.category,
					MatchedBy: matched,
				}, nil
			}
		case MatchTypePinyin:
			if fullPinyin == "" && initials == "" {
				fullPinyin, initials = normalizePinyinContent(content)
			}
			if strings.Contains(fullPinyin, rule.normalizedPattern) || strings.Contains(initials, rule.normalizedPattern) {
				return &KeywordMatch{
					RuleId:    rule.id,
					Pattern:   rule.pattern,
					MatchType: rule.matchType,
					Category:  rule.category,
					MatchedBy: rule.pattern,
				}, nil
			}
		}
	}

	return nil, nil
}

func (f *KeywordFilter) ListRules(ctx context.Context, enabledOnly bool) ([]*model.KeywordRule, error) {
	return f.model.List(ctx, enabledOnly)
}

func (f *KeywordFilter) CreateRule(ctx context.Context, rule *model.KeywordRule) (int64, error) {
	if err := validateKeywordRule(rule); err != nil {
		return 0, err
	}
	if rule.Enabled == 0 {
		rule.Enabled = 1
	}

	id, err := f.model.Insert(ctx, normalizeRule(rule))
	if err != nil {
		var mysqlErr *mysqlDriver.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return 0, ErrDuplicateRule
		}
		return 0, err
	}

	if err := f.Refresh(ctx); err != nil {
		return 0, err
	}

	return id, nil
}

func (f *KeywordFilter) DeleteRule(ctx context.Context, id int64) error {
	if _, err := f.model.FindByIdPrimary(ctx, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrKeywordRuleGone
		}
		return err
	}

	if err := f.model.Delete(ctx, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrKeywordRuleGone
		}
		return err
	}

	return f.Refresh(ctx)
}

func (f *KeywordFilter) Refresh(ctx context.Context) error {
	rules, err := f.model.ListPrimary(ctx, true)
	if err != nil {
		return err
	}
	return f.replaceCache(ctx, rules)
}

func (f *KeywordFilter) loadRules(ctx context.Context) ([]compiledKeywordRule, error) {
	version, versionErr := f.currentVersion(ctx)
	if versionErr == nil {
		f.mu.RLock()
		if version != "" && version == f.cached.version {
			rules := append([]compiledKeywordRule(nil), f.cached.rules...)
			f.mu.RUnlock()
			return rules, nil
		}
		f.mu.RUnlock()

		if version != "" {
			rules, err := f.loadFromRedis(ctx, version)
			if err == nil {
				return rules, nil
			}
		}
	}

	f.mu.RLock()
	if len(f.cached.rules) > 0 {
		rules := append([]compiledKeywordRule(nil), f.cached.rules...)
		f.mu.RUnlock()
		return rules, nil
	}
	f.mu.RUnlock()

	if err := f.Refresh(ctx); err != nil {
		return nil, err
	}

	f.mu.RLock()
	defer f.mu.RUnlock()
	return append([]compiledKeywordRule(nil), f.cached.rules...), nil
}

func (f *KeywordFilter) currentVersion(ctx context.Context) (string, error) {
	if f.store == nil {
		return "", errors.New("redis store not configured")
	}

	version, err := f.store.GetCtx(ctx, keywordRuleVersionKey)
	if err != nil && err != redis.Nil {
		return "", err
	}
	return version, nil
}

func (f *KeywordFilter) loadFromRedis(ctx context.Context, version string) ([]compiledKeywordRule, error) {
	if f.store == nil {
		return nil, errors.New("redis store not configured")
	}

	payload, err := f.store.GetCtx(ctx, keywordRuleDataKey)
	if err != nil {
		return nil, err
	}

	var rules []*model.KeywordRule
	if err := json.Unmarshal([]byte(payload), &rules); err != nil {
		return nil, err
	}

	compiled, err := compileKeywordRules(rules)
	if err != nil {
		return nil, err
	}

	f.mu.Lock()
	f.cached = compiledKeywordRuleSet{
		version: version,
		rules:   compiled,
	}
	f.mu.Unlock()

	return append([]compiledKeywordRule(nil), compiled...), nil
}

func (f *KeywordFilter) replaceCache(ctx context.Context, rules []*model.KeywordRule) error {
	compiled, err := compileKeywordRules(rules)
	if err != nil {
		return err
	}

	version := fmt.Sprintf("%d", time.Now().UnixNano())
	payload, err := json.Marshal(rules)
	if err != nil {
		return err
	}

	if f.store != nil {
		if err := f.store.SetCtx(ctx, keywordRuleDataKey, string(payload)); err != nil {
			return err
		}
		if err := f.store.SetCtx(ctx, keywordRuleVersionKey, version); err != nil {
			return err
		}
	}

	f.mu.Lock()
	f.cached = compiledKeywordRuleSet{
		version: version,
		rules:   compiled,
	}
	f.mu.Unlock()

	return nil
}

func compileKeywordRules(rules []*model.KeywordRule) ([]compiledKeywordRule, error) {
	items := make([]compiledKeywordRule, 0, len(rules))
	for _, rule := range rules {
		if rule == nil || rule.Enabled != 1 {
			continue
		}

		normalized := normalizeRule(rule)
		item := compiledKeywordRule{
			id:           normalized.Id,
			pattern:      normalized.Pattern,
			matchType:    normalized.MatchType,
			category:     normalized.Category,
			lowerPattern: strings.ToLower(normalized.Pattern),
		}

		switch normalized.MatchType {
		case MatchTypeKeyword:
		case MatchTypeRegex:
			regex, err := regexp.Compile(normalized.Pattern)
			if err != nil {
				return nil, fmt.Errorf("compile keyword rule %d: %w", normalized.Id, err)
			}
			item.regex = regex
		case MatchTypePinyin:
			item.normalizedPattern = normalizePinyinPattern(normalized.Pattern)
			if item.normalizedPattern == "" {
				return nil, ErrInvalidPattern
			}
		default:
			return nil, ErrInvalidMatchType
		}

		items = append(items, item)
	}
	return items, nil
}

func validateKeywordRule(rule *model.KeywordRule) error {
	if rule == nil {
		return ErrKeywordRuleEmpty
	}

	normalized := normalizeRule(rule)
	switch {
	case normalized.Pattern == "":
		return ErrKeywordRuleEmpty
	case !validMatchType(normalized.MatchType):
		return ErrInvalidMatchType
	case !validCategory(normalized.Category):
		return ErrInvalidCategory
	}

	switch normalized.MatchType {
	case MatchTypeRegex:
		if _, err := regexp.Compile(normalized.Pattern); err != nil {
			return ErrInvalidPattern
		}
	case MatchTypePinyin:
		if normalizePinyinPattern(normalized.Pattern) == "" {
			return ErrInvalidPattern
		}
	}

	return nil
}

func normalizeRule(rule *model.KeywordRule) *model.KeywordRule {
	if rule == nil {
		return nil
	}

	normalized := *rule
	normalized.Pattern = strings.TrimSpace(normalized.Pattern)
	normalized.MatchType = strings.ToLower(strings.TrimSpace(normalized.MatchType))
	normalized.Category = strings.ToLower(strings.TrimSpace(normalized.Category))
	return &normalized
}

func validMatchType(matchType string) bool {
	switch matchType {
	case MatchTypeKeyword, MatchTypeRegex, MatchTypePinyin:
		return true
	default:
		return false
	}
}

func validCategory(category string) bool {
	switch category {
	case CategoryAbuse, CategorySensitive, CategorySpam, CategoryPlagiarism, CategoryManipulation:
		return true
	default:
		return false
	}
}

func normalizeASCII(value string) string {
	var builder strings.Builder
	for _, r := range strings.ToLower(value) {
		switch {
		case unicode.IsLetter(r), unicode.IsNumber(r):
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func normalizePinyinPattern(value string) string {
	var builder strings.Builder
	for _, r := range strings.ToLower(value) {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func normalizePinyinContent(content string) (string, string) {
	args := pinyin.NewArgs()
	args.Style = pinyin.Normal
	syllables := pinyin.LazyPinyin(content, args)

	var full strings.Builder
	var initials strings.Builder
	for _, syllable := range syllables {
		normalized := normalizeASCII(syllable)
		if normalized == "" {
			continue
		}
		full.WriteString(normalized)
		initials.WriteByte(normalized[0])
	}
	return full.String(), initials.String()
}
