package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"journal/model"
)

// Paper lifecycle zone promotion/demotion thresholds
var zoneConfig = map[string]struct {
	NextZone       string
	MinRatingCount int
	MinShitScore   float64
	MaxStaleAge    time.Duration
	DemoteScore    float64
}{
	"latrine": {
		NextZone:       "septic_tank",
		MinRatingCount: 3,
		MinShitScore:   0.3,
		MaxStaleAge:    30 * 24 * time.Hour,
		DemoteScore:    0.0,
	},
	"septic_tank": {
		NextZone:       "stone",
		MinRatingCount: 10,
		MinShitScore:   0.5,
		MaxStaleAge:    60 * 24 * time.Hour,
		DemoteScore:    0.2,
	},
	"stone": {
		NextZone:       "sediment",
		MinRatingCount: 25,
		MinShitScore:   0.7,
		MaxStaleAge:    90 * 24 * time.Hour,
		DemoteScore:    0.4,
	},
}

func main() {
	dsn := "journal:banishmentB022.@tcp(127.0.0.1:13306)/journal_biz?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("failed to connect mysql:", err)
	}
	defer db.Close()

	paperModel := model.NewPaperModelFromDB(db)
	ctx := context.Background()

	log.Println("=== S.H.I.T Paper Lifecycle Cron Started ===")
	runLifecycle(ctx, paperModel)

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		runLifecycle(ctx, paperModel)
	}
}

func runLifecycle(ctx context.Context, paperModel *model.PaperModel) {
	log.Println("[lifecycle] Running zone transition check...")

	for _, zone := range []string{"latrine", "septic_tank", "stone"} {
		cfg, ok := zoneConfig[zone]
		if !ok {
			continue
		}

		promoted, err := paperModel.GetPapersForPromotion(ctx, zone, cfg.MinRatingCount, cfg.MinShitScore)
		if err != nil {
			log.Printf("[lifecycle] error finding promotion candidates in %s: %v", zone, err)
			continue
		}
		for _, p := range promoted {
			if err := paperModel.UpdateZone(ctx, p.Id, cfg.NextZone); err != nil {
				log.Printf("[lifecycle] error promoting paper %d: %v", p.Id, err)
			} else {
				log.Printf("[lifecycle] ✓ promoted paper %d '%s' %s → %s (score=%.4f, ratings=%d)",
					p.Id, truncate(p.Title, 40), zone, cfg.NextZone, p.ShitScore, p.RatingCount)
			}
		}

		if cfg.DemoteScore > 0 {
			stale, err := paperModel.GetStalePapers(ctx, zone, cfg.MaxStaleAge, cfg.DemoteScore)
			if err != nil {
				log.Printf("[lifecycle] error finding stale papers in %s: %v", zone, err)
				continue
			}
			prevZone := getPrevZone(zone)
			for _, p := range stale {
				if err := paperModel.UpdateZone(ctx, p.Id, prevZone); err != nil {
					log.Printf("[lifecycle] error demoting paper %d: %v", p.Id, err)
				} else {
					log.Printf("[lifecycle] ↓ demoted paper %d '%s' %s → %s (score=%.4f)",
						p.Id, truncate(p.Title, 40), zone, prevZone, p.ShitScore)
				}
			}
		}
	}
	log.Println("[lifecycle] Zone transition check done.")
}

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
