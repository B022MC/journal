package degradation

import (
	"testing"

	"journal/model"
)

func TestCompileKeywordRulesMatchesKeywordRegexAndPinyin(t *testing.T) {
	rules, err := compileKeywordRules([]*model.KeywordRule{
		{Id: 1, Pattern: "free download", MatchType: MatchTypeKeyword, Category: CategorySpam, Enabled: 1},
		{Id: 2, Pattern: "(?i)drop\\s+table", MatchType: MatchTypeRegex, Category: CategorySensitive, Enabled: 1},
		{Id: 3, Pattern: "xueshuzuobi", MatchType: MatchTypePinyin, Category: CategoryPlagiarism, Enabled: 1},
	})
	if err != nil {
		t.Fatalf("compile rules: %v", err)
	}

	filter := &KeywordFilter{}
	filter.cached.rules = rules
	filter.cached.version = "test"

	tests := []struct {
		name      string
		content   string
		wantRule  int64
		wantMatch string
	}{
		{name: "keyword", content: "This paper is a FREE DOWNLOAD now", wantRule: 1, wantMatch: "free download"},
		{name: "regex", content: "please Drop   TABLE users;", wantRule: 2, wantMatch: "Drop   TABLE"},
		{name: "pinyin-full", content: "这篇文章存在学术作弊行为", wantRule: 3, wantMatch: "xueshuzuobi"},
		{name: "pinyin-initials", content: "有人怀疑这里是学术作弊", wantRule: 3, wantMatch: "xueshuzuobi"},
	}

	for _, tc := range tests {
		match, err := filter.Check(t.Context(), tc.content)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", tc.name, err)
		}
		if match == nil {
			t.Fatalf("%s: expected match", tc.name)
		}
		if match.RuleId != tc.wantRule {
			t.Fatalf("%s: expected rule %d, got %d", tc.name, tc.wantRule, match.RuleId)
		}
		if match.MatchedBy != tc.wantMatch {
			t.Fatalf("%s: expected matchedBy %q, got %q", tc.name, tc.wantMatch, match.MatchedBy)
		}
	}
}

func TestValidateKeywordRuleRejectsInvalidConfig(t *testing.T) {
	tests := []struct {
		name string
		rule *model.KeywordRule
		want error
	}{
		{name: "empty", rule: &model.KeywordRule{}, want: ErrKeywordRuleEmpty},
		{name: "bad-type", rule: &model.KeywordRule{Pattern: "foo", MatchType: "glob", Category: CategorySpam}, want: ErrInvalidMatchType},
		{name: "bad-category", rule: &model.KeywordRule{Pattern: "foo", MatchType: MatchTypeKeyword, Category: "ads"}, want: ErrInvalidCategory},
		{name: "bad-regex", rule: &model.KeywordRule{Pattern: "(", MatchType: MatchTypeRegex, Category: CategorySensitive}, want: ErrInvalidPattern},
		{name: "bad-pinyin", rule: &model.KeywordRule{Pattern: "学术作弊", MatchType: MatchTypePinyin, Category: CategoryPlagiarism}, want: ErrInvalidPattern},
	}

	for _, tc := range tests {
		if err := validateKeywordRule(tc.rule); err != tc.want {
			t.Fatalf("%s: expected %v, got %v", tc.name, tc.want, err)
		}
	}
}
