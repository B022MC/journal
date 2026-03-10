package eventing

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"journal/common/achievement"
	"journal/common/cachekeys"
	"journal/common/contribution"
	"journal/common/degradation"
	"journal/model"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	postRateQueueKey       = "events:rating:postrate:v1"
	postRateDeadLetterKey  = postRateQueueKey + ":dead"
	postRateHandlerTimeout = 30 * time.Second
	postRateMaxAttempts    = 3
)

type PostRateEvent struct {
	PaperId    int64 `json:"paper_id"`
	ReviewerId int64 `json:"reviewer_id"`
	AuthorId   int64 `json:"author_id"`
	Attempt    int   `json:"attempt,omitempty"`
	OccurredAt int64 `json:"occurred_at"`
}

func (e PostRateEvent) Validate() error {
	if e.PaperId <= 0 {
		return errors.New("paper_id must be positive")
	}
	if e.ReviewerId <= 0 {
		return errors.New("reviewer_id must be positive")
	}
	return nil
}

func decodePostRateEvent(payload string) (PostRateEvent, error) {
	var event PostRateEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		return PostRateEvent{}, err
	}
	if err := event.Validate(); err != nil {
		return PostRateEvent{}, err
	}
	return event, nil
}

type PostRateQueue struct {
	store *redis.Redis
}

func NewPostRateQueue(store *redis.Redis) *PostRateQueue {
	return &PostRateQueue{store: store}
}

func (q *PostRateQueue) Enqueue(ctx context.Context, event PostRateEvent) error {
	if q == nil || q.store == nil {
		return errors.New("post-rate queue unavailable")
	}
	if event.OccurredAt == 0 {
		event.OccurredAt = time.Now().Unix()
	}
	if err := event.Validate(); err != nil {
		return err
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = q.store.RpushCtx(ctx, postRateQueueKey, string(body))
	return err
}

func (q *PostRateQueue) deadLetter(ctx context.Context, event PostRateEvent) error {
	if q == nil || q.store == nil {
		return nil
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = q.store.RpushCtx(ctx, postRateDeadLetterKey, string(body))
	return err
}

func (q *PostRateQueue) requeue(ctx context.Context, event PostRateEvent) error {
	event.Attempt++
	return q.Enqueue(ctx, event)
}

type PostRateHandler interface {
	HandlePostRateEvent(ctx context.Context, event PostRateEvent) error
}

type PostRateConsumer struct {
	store   *redis.Redis
	queue   *PostRateQueue
	handler PostRateHandler
}

func NewPostRateConsumer(store *redis.Redis, queue *PostRateQueue, handler PostRateHandler) *PostRateConsumer {
	return &PostRateConsumer{
		store:   store,
		queue:   queue,
		handler: handler,
	}
}

func (c *PostRateConsumer) Start(ctx context.Context) {
	if c == nil || c.store == nil || c.queue == nil || c.handler == nil {
		return
	}

	node, err := redis.CreateBlockingNode(c.store)
	if err != nil {
		logx.WithContext(ctx).Errorf("failed to create rating event blocking redis node: %v", err)
		return
	}
	defer node.Close()

	logger := logx.WithContext(ctx)
	for {
		if ctx.Err() != nil {
			return
		}

		payload, ok, err := c.store.BlpopExCtx(ctx, node, postRateQueueKey)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			logger.Errorf("rating event dequeue failed: %v", err)
			time.Sleep(time.Second)
			continue
		}
		if !ok {
			continue
		}

		event, err := decodePostRateEvent(payload)
		if err != nil {
			logger.Errorf("invalid rating event payload dropped: %v", err)
			continue
		}

		handleCtx, cancel := context.WithTimeout(ctx, postRateHandlerTimeout)
		err = c.handler.HandlePostRateEvent(handleCtx, event)
		cancel()
		if err == nil {
			continue
		}

		logger.Errorf("rating event handle failed for paper=%d reviewer=%d attempt=%d: %v", event.PaperId, event.ReviewerId, event.Attempt, err)
		if event.Attempt+1 >= postRateMaxAttempts {
			if deadErr := c.queue.deadLetter(context.Background(), event); deadErr != nil {
				logger.Errorf("rating event dead-letter failed for paper=%d reviewer=%d: %v", event.PaperId, event.ReviewerId, deadErr)
			}
			continue
		}
		if retryErr := c.queue.requeue(context.Background(), event); retryErr != nil {
			logger.Errorf("rating event requeue failed for paper=%d reviewer=%d: %v", event.PaperId, event.ReviewerId, retryErr)
		}
	}
}

type PostRateProcessor struct {
	paperModel          *model.PaperModel
	ratingModel         *model.RatingModel
	contributionManager *contribution.Manager
	ratingGuard         *degradation.RatingAnomalyGuard
	achievementService  *achievement.Service
	cacheStore          *redis.Redis
}

func NewPostRateProcessor(
	paperModel *model.PaperModel,
	ratingModel *model.RatingModel,
	contributionManager *contribution.Manager,
	ratingGuard *degradation.RatingAnomalyGuard,
	achievementService *achievement.Service,
	cacheStore *redis.Redis,
) *PostRateProcessor {
	return &PostRateProcessor{
		paperModel:          paperModel,
		ratingModel:         ratingModel,
		contributionManager: contributionManager,
		ratingGuard:         ratingGuard,
		achievementService:  achievementService,
		cacheStore:          cacheStore,
	}
}

func (p *PostRateProcessor) HandlePostRateEvent(ctx context.Context, event PostRateEvent) error {
	if err := event.Validate(); err != nil {
		return err
	}

	paper, err := p.paperModel.FindByIdPrimary(ctx, event.PaperId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	avgScore, count, controversy, err := p.ratingModel.GetPaperRatingStats(ctx, event.PaperId)
	if err != nil {
		return fmt.Errorf("load paper rating stats: %w", err)
	}
	weightedStats, err := p.ratingModel.GetWeightedRatingStats(ctx, event.PaperId)
	if err != nil {
		return fmt.Errorf("load weighted rating stats: %w", err)
	}

	if err := p.paperModel.UpdateScoresV2(
		ctx,
		event.PaperId,
		avgScore,
		count,
		paper.ViewCount,
		controversy,
		weightedStats.WeightedAvg,
		weightedStats.AvgReviewerAuth,
		paper.CreatedAt,
	); err != nil {
		return fmt.Errorf("update paper scores: %w", err)
	}

	if _, _, err := p.contributionManager.SyncUser(ctx, event.ReviewerId); err != nil {
		return fmt.Errorf("sync reviewer contribution: %w", err)
	}
	if p.achievementService != nil {
		if err := p.achievementService.SyncUser(ctx, event.ReviewerId); err != nil {
			return fmt.Errorf("sync reviewer achievements: %w", err)
		}
	}

	authorId := event.AuthorId
	if authorId <= 0 {
		authorId = paper.AuthorId
	}
	if authorId > 0 && authorId != event.ReviewerId {
		if _, _, err := p.contributionManager.SyncUser(ctx, authorId); err != nil {
			return fmt.Errorf("sync author contribution: %w", err)
		}
	}

	if err := p.ratingGuard.InspectAfterRating(ctx, event.PaperId, event.ReviewerId); err != nil {
		return fmt.Errorf("inspect rating anomalies: %w", err)
	}

	p.invalidateCaches(ctx, paper, event.ReviewerId, authorId)

	return nil
}

func (p *PostRateProcessor) invalidateCaches(ctx context.Context, paper *model.Paper, reviewerId, authorId int64) {
	if p == nil || p.cacheStore == nil || paper == nil {
		return
	}

	keys := []string{
		cachekeys.HotPapers(""),
		cachekeys.HotPapers(paper.Zone),
		cachekeys.FlagStatus("paper", paper.Id),
		cachekeys.PaperModeration(paper.Id),
	}
	if reviewerId > 0 {
		keys = append(keys, cachekeys.Contribution(reviewerId))
	}
	if authorId > 0 && authorId != reviewerId {
		keys = append(keys, cachekeys.Contribution(authorId))
	}

	_, err := p.cacheStore.DelCtx(ctx, keys...)
	if err != nil && err != redis.Nil {
		logx.WithContext(ctx).Errorf("invalidate post-rate caches failed for paper=%d: %v", paper.Id, err)
	}
}
