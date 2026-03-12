package search

import (
	"encoding/json"
	"os"
	"slices"
	"strings"
)

type synonymMap map[string][]string

var builtInSynonyms = synonymMap{
	"ai":    {"人工智能"},
	"人工智能":  {"ai", "机器学习"},
	"机器学习":  {"ml", "人工智能"},
	"ml":    {"机器学习"},
	"论文":    {"paper", "文章", "文献"},
	"paper": {"论文", "文献"},
	"文献":    {"论文", "paper"},
	"推荐系统":  {"推荐", "排序"},
	"排序":    {"ranking", "推荐"},
	"检索":    {"搜索", "召回"},
	"搜索":    {"检索", "召回"},
}

func loadSynonyms(path string) (synonymMap, error) {
	merged := synonymMap{}
	for key, values := range builtInSynonyms {
		merged[key] = append([]string{}, values...)
	}
	if strings.TrimSpace(path) == "" {
		return normalizeSynonyms(merged), nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var extra map[string][]string
	if err := json.Unmarshal(raw, &extra); err != nil {
		return nil, err
	}
	for key, values := range extra {
		key = normalizeText(key)
		if key == "" {
			continue
		}
		merged[key] = append(merged[key], values...)
	}
	return normalizeSynonyms(merged), nil
}

func normalizeSynonyms(in synonymMap) synonymMap {
	out := synonymMap{}
	for key, values := range in {
		key = normalizeText(key)
		if key == "" {
			continue
		}
		normalized := make([]string, 0, len(values))
		for _, value := range values {
			value = normalizeText(value)
			if value == "" || value == key {
				continue
			}
			normalized = append(normalized, value)
		}
		normalized = uniqueSorted(normalized)
		if len(normalized) == 0 {
			continue
		}
		out[key] = normalized
	}
	return out
}

func (m synonymMap) expand(term string) []string {
	if len(m) == 0 {
		return nil
	}
	values := append([]string{}, m[normalizeText(term)]...)
	slices.Sort(values)
	return values
}
