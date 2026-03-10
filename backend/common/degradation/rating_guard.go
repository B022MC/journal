package degradation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"journal/common/consts"
	"journal/model"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

const (
	ratingBurstWindowMinutes   = 10
	ratingBurstThreshold       = 8
	bimodalityMinSamples       = 8
	bimodalityThreshold        = 0.55
	ratingIdentityWindowHours  = 24
	ratingIPClusterThreshold   = 3
	ratingFingerprintThreshold = 2
	maxRatingUserAgentLength   = 512
)

type RatingAnomalyGuard struct {
	ratingModel *model.RatingModel
	flagModel   *model.FlagModel
	engine      *Engine
}

func NewRatingAnomalyGuard(ratingModel *model.RatingModel, flagModel *model.FlagModel, engine *Engine) *RatingAnomalyGuard {
	return &RatingAnomalyGuard{
		ratingModel: ratingModel,
		flagModel:   flagModel,
		engine:      engine,
	}
}

func (g *RatingAnomalyGuard) InspectAfterRating(ctx context.Context, paperId, userId int64) error {
	if err := g.detectUserBurst(ctx, userId); err != nil {
		return err
	}
	if err := g.detectPaperBimodality(ctx, paperId); err != nil {
		return err
	}
	return nil
}

func (g *RatingAnomalyGuard) InspectIdentityCluster(ctx context.Context, paperId int64, sourceIP, userAgent string) error {
	sourceIP = strings.TrimSpace(sourceIP)
	userAgent = NormalizeRatingUserAgent(userAgent)

	if sourceIP == "" {
		return nil
	}

	deviceFingerprint := BuildRatingDeviceFingerprint(sourceIP, userAgent)
	since := time.Now().Add(-ratingIdentityWindowHours * time.Hour)

	ipUsers, err := g.ratingModel.CountDistinctUsersByPaperIPSince(ctx, paperId, sourceIP, since)
	if err != nil {
		return err
	}

	var fingerprintUsers int32
	if deviceFingerprint != "" {
		fingerprintUsers, err = g.ratingModel.CountDistinctUsersByPaperFingerprintSince(ctx, paperId, deviceFingerprint, since)
		if err != nil {
			return err
		}
	}

	if ipUsers < ratingIPClusterThreshold && fingerprintUsers < ratingFingerprintThreshold {
		return nil
	}

	detailParts := []string{
		fmt.Sprintf("system identity cluster detection: source_ip=%s distinct_users=%d/%dh", sourceIP, ipUsers, ratingIdentityWindowHours),
	}
	if fingerprintUsers >= ratingFingerprintThreshold && deviceFingerprint != "" {
		detailParts = append(detailParts, fmt.Sprintf("fingerprint=%s distinct_users=%d/%dh", shortFingerprint(deviceFingerprint), fingerprintUsers, ratingIdentityWindowHours))
	}

	return g.ensureSystemFlag(
		ctx,
		consts.FlagTargetPaper,
		paperId,
		consts.SystemReporterRatingFingerprint,
		consts.FlagReasonManipulation,
		strings.Join(detailParts, "; "),
	)
}

func BimodalityCoefficient(buckets [10]int32) float64 {
	var total float64
	var mean float64
	for idx, count := range buckets {
		score := float64(idx + 1)
		weight := float64(count)
		total += weight
		mean += score * weight
	}
	if total < bimodalityMinSamples {
		return 0
	}

	mean /= total

	var m2, m3, m4 float64
	for idx, count := range buckets {
		score := float64(idx + 1)
		weight := float64(count)
		delta := score - mean
		m2 += weight * math.Pow(delta, 2)
		m3 += weight * math.Pow(delta, 3)
		m4 += weight * math.Pow(delta, 4)
	}

	m2 /= total
	m3 /= total
	m4 /= total

	if m2 == 0 {
		return 0
	}

	skewness := m3 / math.Pow(m2, 1.5)
	kurtosis := m4 / math.Pow(m2, 2)
	if kurtosis == 0 {
		return 0
	}

	return (math.Pow(skewness, 2) + 1) / kurtosis
}

func (g *RatingAnomalyGuard) detectUserBurst(ctx context.Context, userId int64) error {
	since := time.Now().Add(-ratingBurstWindowMinutes * time.Minute)
	count, err := g.ratingModel.CountByUserSince(ctx, userId, since)
	if err != nil {
		return err
	}
	if count < ratingBurstThreshold {
		return nil
	}

	detail := fmt.Sprintf("system burst detection: user submitted %d ratings within %d minutes", count, ratingBurstWindowMinutes)
	return g.ensureSystemFlag(ctx, consts.FlagTargetUser, userId, consts.SystemReporterRatingBurst, consts.FlagReasonManipulation, detail)
}

func (g *RatingAnomalyGuard) detectPaperBimodality(ctx context.Context, paperId int64) error {
	histogram, err := g.ratingModel.GetPaperScoreHistogram(ctx, paperId)
	if err != nil {
		return err
	}
	if histogram.Total < bimodalityMinSamples {
		return nil
	}

	coefficient := BimodalityCoefficient(histogram.Buckets)
	if coefficient <= bimodalityThreshold {
		return nil
	}

	detail := fmt.Sprintf(
		"system bimodality detection: coefficient=%.4f sample=%d histogram=%s",
		coefficient,
		histogram.Total,
		formatHistogram(histogram.Buckets),
	)
	return g.ensureSystemFlag(ctx, consts.FlagTargetPaper, paperId, consts.SystemReporterRatingBimodality, consts.FlagReasonManipulation, detail)
}

func (g *RatingAnomalyGuard) ensureSystemFlag(ctx context.Context, targetType string, targetId, reporterId int64, reason, detail string) error {
	flagged, err := g.flagModel.HasFlagged(ctx, targetType, targetId, reporterId)
	if err != nil {
		return err
	}
	if flagged {
		return nil
	}

	flag := &model.Flag{
		TargetType:           targetType,
		TargetId:             targetId,
		ReporterId:           reporterId,
		Reason:               reason,
		Detail:               detail,
		ReporterContribution: 0,
	}
	_, _, err = g.engine.ProcessFlag(ctx, flag)
	if err == nil {
		return nil
	}

	var mysqlErr *mysqlDriver.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return nil
	}
	return err
}

func formatHistogram(buckets [10]int32) string {
	parts := make([]string, 0, len(buckets))
	for idx, count := range buckets {
		if count == 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%d:%d", idx+1, count))
	}
	if len(parts) == 0 {
		return "empty"
	}
	return strings.Join(parts, ",")
}

func NormalizeRatingUserAgent(userAgent string) string {
	normalized := strings.Join(strings.Fields(strings.TrimSpace(userAgent)), " ")
	if len(normalized) > maxRatingUserAgentLength {
		return normalized[:maxRatingUserAgentLength]
	}
	return normalized
}

func BuildRatingDeviceFingerprint(sourceIP, userAgent string) string {
	sourceIP = strings.TrimSpace(sourceIP)
	userAgent = NormalizeRatingUserAgent(userAgent)
	if sourceIP == "" || userAgent == "" {
		return ""
	}

	sum := sha256.Sum256([]byte(strings.ToLower(sourceIP + "|" + userAgent)))
	return hex.EncodeToString(sum[:])
}

func shortFingerprint(deviceFingerprint string) string {
	if len(deviceFingerprint) <= 12 {
		return deviceFingerprint
	}
	return deviceFingerprint[:12]
}
