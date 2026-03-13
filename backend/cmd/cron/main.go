package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"journal/common/contribution"
	"journal/common/degradation"
	"journal/model"
)

// ==================== Zone Lifecycle Config ====================

// PromotionConfigV2 defines multi-dimensional promotion thresholds
type PromotionConfigV2 struct {
	NextZone         string
	MinRatingCount   int
	MinShitScore     float64
	MinWeightedCount int
	MinReviewerAuth  float64
	MinAgeDays       int
	// Demotion
	MaxStaleAge time.Duration
	DemoteScore float64
	DemoteFlags int
}

var zoneConfigV2 = map[string]PromotionConfigV2{
	"latrine": {
		NextZone:         "septic_tank",
		MinRatingCount:   5,
		MinShitScore:     0.25,
		MinWeightedCount: 3,
		MinReviewerAuth:  0.10,
		MinAgeDays:       3,
		MaxStaleAge:      60 * 24 * time.Hour,
		DemoteScore:      0.0,
		DemoteFlags:      2,
	},
	"septic_tank": {
		NextZone:         "stone",
		MinRatingCount:   15,
		MinShitScore:     0.45,
		MinWeightedCount: 8,
		MinReviewerAuth:  0.25,
		MinAgeDays:       14,
		MaxStaleAge:      90 * 24 * time.Hour,
		DemoteScore:      0.20,
		DemoteFlags:      3,
	},
	"stone": {
		NextZone:         "sediment",
		MinRatingCount:   30,
		MinShitScore:     0.65,
		MinWeightedCount: 20,
		MinReviewerAuth:  0.40,
		MinAgeDays:       30,
		MaxStaleAge:      120 * 24 * time.Hour,
		DemoteScore:      0.40,
		DemoteFlags:      5,
	},
}

func main() {
	dsn := "journal:banishmentB022.@tcp(127.0.0.1:13306)/journal?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("failed to connect mysql:", err)
	}
	defer db.Close()

	paperModel := model.NewPaperModelFromDB(db)
	userModel := model.NewUserModelFromDB(db)
	ratingModel := model.NewRatingModelFromDB(db)
	flagModel := model.NewFlagModelFromDB(db)

	calc := contribution.NewCalculator(userModel, paperModel, ratingModel)
	degradeEngine := degradation.NewEngine(flagModel, paperModel, userModel)

	ctx := context.Background()

	log.Println("=== S.H.I.T Unified Cron Started ===")
	log.Println("[cron] Jobs:")
	log.Println("  - @every 1h  : Zone lifecycle (promotion/demotion)")
	log.Println("  - @daily 3:00: Contribution decay + Role audit")
	log.Println("  - @daily 4:00: Degradation sweep")
	log.Println("  - @daily 5:00: Cold data archive")

	// Run lifecycle immediately on startup
	runLifecycleV2(ctx, paperModel)

	// Simple scheduler using tickers
	lifecycleTicker := time.NewTicker(1 * time.Hour)
	defer lifecycleTicker.Stop()

	// Calculate next daily run times
	go func() {
		for {
			now := time.Now()
			// Next 3:00 AM
			next3AM := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location())
			if now.After(next3AM) {
				next3AM = next3AM.Add(24 * time.Hour)
			}
			time.Sleep(time.Until(next3AM))

			log.Println("[cron] === Daily contribution decay + role audit ===")
			runContributionDecay(ctx, userModel)
			time.Sleep(5 * time.Minute)
			runRoleAudit(ctx, userModel, calc)
		}
	}()

	go func() {
		for {
			now := time.Now()
			// Next 4:00 AM
			next4AM := time.Date(now.Year(), now.Month(), now.Day(), 4, 0, 0, 0, now.Location())
			if now.After(next4AM) {
				next4AM = next4AM.Add(24 * time.Hour)
			}
			time.Sleep(time.Until(next4AM))

			log.Println("[cron] === Daily degradation sweep ===")
			runDegradationSweep(ctx, degradeEngine, flagModel)
		}
	}()

	// Main loop: hourly lifecycle
	go func() {
		for {
			now := time.Now()
			// Next 5:00 AM
			next5AM := time.Date(now.Year(), now.Month(), now.Day(), 5, 0, 0, 0, now.Location())
			if now.After(next5AM) {
				next5AM = next5AM.Add(24 * time.Hour)
			}
			time.Sleep(time.Until(next5AM))

			log.Println("[cron] === Daily cold data archive ===")
			runColdDataSweep(ctx, paperModel)
		}
	}()

	for range lifecycleTicker.C {
		runLifecycleV2(ctx, paperModel)
	}
}

// ==================== Zone Lifecycle V2 ====================

func runLifecycleV2(ctx context.Context, paperModel *model.PaperModel) {
	log.Println("[lifecycle] Running zone transition check (v2)...")

	for _, zone := range []string{"latrine", "septic_tank", "stone"} {
		cfg, ok := zoneConfigV2[zone]
		if !ok {
			continue
		}

		// Promotion with multi-dimensional thresholds
		promoted, err := paperModel.GetPapersForPromotionV2(ctx, zone,
			cfg.MinRatingCount, cfg.MinShitScore, cfg.MinWeightedCount,
			cfg.MinReviewerAuth, cfg.MinAgeDays)
		if err != nil {
			log.Printf("[lifecycle] error finding promotion candidates in %s: %v", zone, err)
			continue
		}
		for _, p := range promoted {
			if err := paperModel.UpdateZone(ctx, p.Id, cfg.NextZone); err != nil {
				log.Printf("[lifecycle] error promoting paper %d: %v", p.Id, err)
			} else {
				log.Printf("[lifecycle] ✓ promoted paper %d '%s' %s → %s (score=%.4f, ratings=%d, auth=%.4f)",
					p.Id, truncate(p.Title, 40), zone, cfg.NextZone, p.ShitScore, p.RatingCount, p.ReviewerAuthority)
			}
		}

		// Demotion with flag-based criteria
		if cfg.DemoteScore > 0 || cfg.DemoteFlags > 0 {
			stale, err := paperModel.GetStalePapersV2(ctx, zone, cfg.MaxStaleAge, cfg.DemoteScore, cfg.DemoteFlags)
			if err != nil {
				log.Printf("[lifecycle] error finding stale papers in %s: %v", zone, err)
				continue
			}
			prevZone := getPrevZone(zone)
			for _, p := range stale {
				if err := paperModel.UpdateZone(ctx, p.Id, prevZone); err != nil {
					log.Printf("[lifecycle] error demoting paper %d: %v", p.Id, err)
				} else {
					log.Printf("[lifecycle] ↓ demoted paper %d '%s' %s → %s (score=%.4f, flags=%d)",
						p.Id, truncate(p.Title, 40), zone, prevZone, p.ShitScore, p.FlagCount)
				}
			}
		}
	}
	log.Println("[lifecycle] Zone transition check done.")
}

// ==================== Contribution Decay ====================

func runContributionDecay(ctx context.Context, userModel *model.UserModel) {
	log.Println("[contribution] Running decay for inactive users...")
	// Decay 5% of score for users inactive > 30 days
	affected, err := userModel.BatchDecayContribution(ctx, 30, 0.05)
	if err != nil {
		log.Printf("[contribution] error during batch decay: %v", err)
		return
	}
	log.Printf("[contribution] decayed %d users", affected)
}

// ==================== Role Audit ====================

func runRoleAudit(ctx context.Context, userModel *model.UserModel, calc *contribution.Calculator) {
	log.Println("[role-audit] Running automatic role assignment...")
	users, err := userModel.GetAllActiveUsers(ctx)
	if err != nil {
		log.Printf("[role-audit] error fetching users: %v", err)
		return
	}

	updated := 0
	for _, u := range users {
		score, err := calc.CalcForUser(ctx, u.Id)
		if err != nil {
			log.Printf("[role-audit] error calculating score for user %d: %v", u.Id, err)
			continue
		}
		if err := userModel.UpdateContributionScore(ctx, u.Id, score); err != nil {
			log.Printf("[role-audit] error updating contribution for user %d: %v", u.Id, err)
			continue
		}

		expectedRole := contribution.RoleForScore(score)
		if expectedRole != u.Role {
			if err := userModel.AutoAssignRole(ctx, u.Id, expectedRole); err != nil {
				log.Printf("[role-audit] error updating role for user %d: %v", u.Id, err)
			} else {
				log.Printf("[role-audit] ↕ user %d '%s' role %d → %d (score=%.2f)",
					u.Id, u.Username, u.Role, expectedRole, score)
				updated++
			}
		}
	}
	log.Printf("[role-audit] audit complete, %d users updated", updated)
}

// ==================== Degradation Sweep ====================

func runDegradationSweep(ctx context.Context, engine *degradation.Engine, flagModel *model.FlagModel) {
	log.Println("[degradation] Running pending flag sweep...")
	flags, _, err := flagModel.ListPending(ctx, 1, 500)
	if err != nil {
		log.Printf("[degradation] error fetching pending flags: %v", err)
		return
	}

	processed := 0
	for _, f := range flags {
		level, err := engine.EvaluateDegradation(ctx, f.TargetType, f.TargetId)
		if err != nil {
			log.Printf("[degradation] error evaluating %s/%d: %v", f.TargetType, f.TargetId, err)
			continue
		}
		if level > 0 {
			processed++
			log.Printf("[degradation] ⚠ %s/%d → level %d", f.TargetType, f.TargetId, level)
		}
	}
	log.Printf("[degradation] sweep complete, %d targets degraded", processed)
}

// ==================== Helpers ====================

func getPrevZone(zone string) string {
	switch zone {
	case "stone":
		return "septic_tank"
	case "septic_tank":
		return "latrine"
	default:
		return "latrine"
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return fmt.Sprintf("%s...", s[:maxLen-3])
}

// ==================== Cold Data Sweep ====================

// runColdDataSweep 归档冷数据：>90天未访问且zone=sediment的论文，或status=0的论文
// 每次最多处理 200 篇，避免长事务
func runColdDataSweep(ctx context.Context, paperModel *model.PaperModel) {
	const coldDays = 90
	const batchSize = 200

	log.Println("[cold-data] Scanning for cold papers...")
	papers, err := paperModel.GetColdPapers(ctx, coldDays, batchSize)
	if err != nil {
		log.Printf("[cold-data] error scanning cold papers: %v", err)
		return
	}

	if len(papers) == 0 {
		log.Println("[cold-data] no cold papers found")
		return
	}

	archived := 0
	for _, p := range papers {
		if err := paperModel.ArchiveColdPaper(ctx, p.Id); err != nil {
			log.Printf("[cold-data] error archiving paper %d: %v", p.Id, err)
		} else {
			log.Printf("[cold-data] ❄ archived paper %d '%s' (zone=%s, last_access=%v)",
				p.Id, truncate(p.Title, 40), p.Zone, p.LastAccessedAt.Time)
			archived++
		}
	}
	log.Printf("[cold-data] sweep complete, %d/%d papers archived", archived, len(papers))
}
