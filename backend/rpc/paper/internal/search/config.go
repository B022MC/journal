package search

const (
	EngineFulltext = "fulltext"
	EngineHybrid   = "hybrid"
)

type Config struct {
	DefaultEngine  string
	ShadowCompare  bool
	QueryTimeoutMs int
	BuildTimeoutMs int
	MaxDocuments   int
	BatchOne       BatchOneConfig
	BatchTwo       BatchTwoConfig
}

type BatchOneConfig struct {
	Enabled     bool
	WorkerCount int
	Explain     bool
	EnableIK    bool
	EnableJieba bool
	LexiconPath string
}

type BatchTwoConfig struct {
	TrieEnabled           bool
	SynonymEnabled        bool
	SynonymPath           string
	FusionEnabled         bool
	FusionBM25Weight      float64
	FusionFreshnessWeight float64
	FusionQualityWeight   float64
}

func (c Config) Normalized() Config {
	cfg := c
	if cfg.DefaultEngine != EngineHybrid {
		cfg.DefaultEngine = EngineFulltext
	}
	if cfg.QueryTimeoutMs <= 0 {
		cfg.QueryTimeoutMs = 150
	}
	if cfg.BuildTimeoutMs <= 0 {
		cfg.BuildTimeoutMs = 5000
	}
	if cfg.MaxDocuments <= 0 {
		cfg.MaxDocuments = 10000
	}
	if cfg.BatchOne.WorkerCount <= 0 {
		cfg.BatchOne.WorkerCount = 4
	}
	if !cfg.BatchOne.EnableIK && !cfg.BatchOne.EnableJieba {
		cfg.BatchOne.EnableIK = true
		cfg.BatchOne.EnableJieba = true
	}
	if !cfg.BatchOne.Enabled {
		cfg.BatchOne.Explain = false
	}
	if cfg.BatchTwo.FusionBM25Weight <= 0 {
		cfg.BatchTwo.FusionBM25Weight = 0.75
	}
	if cfg.BatchTwo.FusionFreshnessWeight <= 0 {
		cfg.BatchTwo.FusionFreshnessWeight = 0.15
	}
	if cfg.BatchTwo.FusionQualityWeight <= 0 {
		cfg.BatchTwo.FusionQualityWeight = 0.10
	}
	totalWeight := cfg.BatchTwo.FusionBM25Weight + cfg.BatchTwo.FusionFreshnessWeight + cfg.BatchTwo.FusionQualityWeight
	if totalWeight <= 0 {
		cfg.BatchTwo.FusionBM25Weight = 0.75
		cfg.BatchTwo.FusionFreshnessWeight = 0.15
		cfg.BatchTwo.FusionQualityWeight = 0.10
		totalWeight = 1
	}
	cfg.BatchTwo.FusionBM25Weight /= totalWeight
	cfg.BatchTwo.FusionFreshnessWeight /= totalWeight
	cfg.BatchTwo.FusionQualityWeight /= totalWeight
	return cfg
}
