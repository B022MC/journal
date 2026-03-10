package consts

// ==================== User Role ====================

const (
	UserRoleMember  int32 = 0
	UserRoleScooper int32 = 1
	UserRoleEditor  int32 = 2
	UserRoleAdmin   int32 = 3
)

// ==================== User Status ====================

const (
	UserStatusBanned int32 = 0
	UserStatusActive int32 = 1
)

// ==================== Paper Status ====================

const (
	PaperStatusDeleted int32 = 0
	PaperStatusActive  int32 = 1
	PaperStatusFlagged int32 = 2
)

// ==================== Paper Zone ====================

const (
	PaperZoneLatrine  = "latrine"
	PaperZoneSeptic   = "septic_tank"
	PaperZoneStone    = "stone"
	PaperZoneSediment = "sediment"
)

// ==================== Paper Discipline ====================

const (
	DisciplineScience     = "science"
	DisciplineHumanities  = "humanities"
	DisciplineInformation = "information"
	DisciplineTechnology  = "technology"
	DisciplineOther       = "other"
)

// ==================== Paper Degradation Level ====================

const (
	DegradationNormal   int32 = 0
	DegradationWatched  int32 = 1
	DegradationThrottle int32 = 2
	DegradationSealed   int32 = 3
)

// ==================== Flag Status ====================

const (
	FlagStatusPending  int32 = 0
	FlagStatusDegraded int32 = 1
	FlagStatusDismiss  int32 = 2
)

// ==================== Flag Target Type ====================

const (
	FlagTargetPaper  = "paper"
	FlagTargetRating = "rating"
	FlagTargetUser   = "user"
)

// ==================== Flag Reason ====================

const (
	FlagReasonAbuse        = "abuse"
	FlagReasonSpam         = "spam"
	FlagReasonPlagiarism   = "plagiarism"
	FlagReasonSensitive    = "sensitive"
	FlagReasonManipulation = "manipulation"
)

// ==================== System Reporter IDs ====================

const (
	SystemReporterSimhash           int64 = 900000000001
	SystemReporterRatingBurst       int64 = 900000000002
	SystemReporterRatingBimodality  int64 = 900000000003
	SystemReporterRatingFingerprint int64 = 900000000004
)

// ==================== News Status ====================

const (
	NewsStatusDraft     int32 = 0
	NewsStatusPublished int32 = 1
)

// ==================== News Category ====================

const (
	NewsCatAnnouncement = "announcement"
	NewsCatGovernance   = "governance"
	NewsCatMaintenance  = "maintenance"
	NewsCatFeature      = "feature"
)
