package degradation

import (
	"strings"
	"sync"
)

// KeywordFilter provides content filtering based on keyword blacklists.
// It supports hot-reloading of keywords without service restart.
type KeywordFilter struct {
	mu       sync.RWMutex
	keywords map[string]string // keyword → category
}

// Default keyword categories
const (
	CategoryAbuse     = "abuse"
	CategorySensitive = "sensitive"
	CategorySpam      = "spam"
)

// NewKeywordFilter creates a filter with default keywords
func NewKeywordFilter() *KeywordFilter {
	f := &KeywordFilter{
		keywords: make(map[string]string),
	}
	// Load default minimal keywords (in production, load from config/DB)
	f.loadDefaults()
	return f
}

// Check scans content against keyword blacklist.
// Returns (matched bool, category string, matchedKeyword string)
func (f *KeywordFilter) Check(content string) (bool, string, string) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	lower := strings.ToLower(content)
	for keyword, category := range f.keywords {
		if strings.Contains(lower, keyword) {
			return true, category, keyword
		}
	}
	return false, "", ""
}

// AddKeyword adds a keyword to the blacklist (thread-safe)
func (f *KeywordFilter) AddKeyword(keyword, category string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.keywords[strings.ToLower(keyword)] = category
}

// RemoveKeyword removes a keyword from the blacklist
func (f *KeywordFilter) RemoveKeyword(keyword string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.keywords, strings.ToLower(keyword))
}

// ReloadKeywords replaces the entire keyword set (for hot-reload from config)
func (f *KeywordFilter) ReloadKeywords(newKeywords map[string]string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.keywords = make(map[string]string, len(newKeywords))
	for k, v := range newKeywords {
		f.keywords[strings.ToLower(k)] = v
	}
}

// loadDefaults adds a minimal set of default filter keywords
// In production, these should be loaded from database or config file
func (f *KeywordFilter) loadDefaults() {
	// Intentionally minimal — real deployment should load from external source
	defaults := map[string]string{
		// Spam patterns
		"click here to buy":  CategorySpam,
		"free download":      CategorySpam,
		"limited time offer": CategorySpam,
	}
	for k, v := range defaults {
		f.keywords[k] = v
	}
}

// Count returns the number of keywords in the blacklist
func (f *KeywordFilter) Count() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return len(f.keywords)
}
