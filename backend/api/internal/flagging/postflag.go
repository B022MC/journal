package flagging

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"journal/common/cachekeys"
	"journal/common/consts"
	"journal/common/degradation"
	"journal/model"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	postFlagQueueKey       = "events:flag:postsubmit:v1"
	postFlagDeadLetterKey  = postFlagQueueKey + ":dead"
	postFlagHandlerTimeout = 30 * time.Second
	postFlagMaxAttempts    = 3
)

type PostFlagEvent struct {
	FlagId     int64  `json:"flag_id,omitempty"`
	TargetType string `json:"target_type"`
	TargetId   int64  `json:"target_id"`
	ReporterId int64  `json:"reporter_id,omitempty"`
	Attempt    int    `json:"attempt,omitempty"`
	OccurredAt int64  `json:"occurred_at"`
}

func (e PostFlagEvent) Validate() error {
	if e.TargetId <= 0 {
		return errors.New("target_id must be positive")
	}

	switch e.TargetType {
	case consts.FlagTargetPaper, consts.FlagTargetRating:
		return nil
	default:
		return errors.New("target_type must be paper or rating")
	}
}

func decodePostFlagEvent(payload string) (PostFlagEvent, error) {
	var event PostFlagEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		return PostFlagEvent{}, err
	}
	if err := event.Validate(); err != nil {
		return PostFlagEvent{}, err
	}
	return event, nil
}

type PostFlagQueue struct {
	store *redis.Redis
}

func NewPostFlagQueue(store *redis.Redis) *PostFlagQueue {
	return &PostFlagQueue{store: store}
}

func (q *PostFlagQueue) Enqueue(ctx context.Context, event PostFlagEvent) error {
	if q == nil || q.store == nil {
		return errors.New("post-flag queue unavailable")
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
	_, err = q.store.RpushCtx(ctx, postFlagQueueKey, string(body))
	return err
}

func (q *PostFlagQueue) deadLetter(ctx context.Context, event PostFlagEvent) error {
	if q == nil || q.store == nil {
		return nil
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = q.store.RpushCtx(ctx, postFlagDeadLetterKey, string(body))
	return err
}

func (q *PostFlagQueue) requeue(ctx context.Context, event PostFlagEvent) error {
	event.Attempt++
	return q.Enqueue(ctx, event)
}

type PostFlagHandler interface {
	HandlePostFlagEvent(ctx context.Context, event PostFlagEvent) error
}

type PostFlagConsumer struct {
	store   *redis.Redis
	queue   *PostFlagQueue
	handler PostFlagHandler
}

func NewPostFlagConsumer(store *redis.Redis, queue *PostFlagQueue, handler PostFlagHandler) *PostFlagConsumer {
	return &PostFlagConsumer{
		store:   store,
		queue:   queue,
		handler: handler,
	}
}

func (c *PostFlagConsumer) Start(ctx context.Context) {
	if c == nil || c.store == nil || c.queue == nil || c.handler == nil {
		return
	}

	node, err := redis.CreateBlockingNode(c.store)
	if err != nil {
		logx.WithContext(ctx).Errorf("failed to create flag event blocking redis node: %v", err)
		return
	}
	defer node.Close()

	logger := logx.WithContext(ctx)
	for {
		if ctx.Err() != nil {
			return
		}

		payload, ok, err := c.store.BlpopExCtx(ctx, node, postFlagQueueKey)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			logger.Errorf("flag event dequeue failed: %v", err)
			time.Sleep(time.Second)
			continue
		}
		if !ok {
			continue
		}

		event, err := decodePostFlagEvent(payload)
		if err != nil {
			logger.Errorf("invalid flag event payload dropped: %v", err)
			continue
		}

		handleCtx, cancel := context.WithTimeout(ctx, postFlagHandlerTimeout)
		err = c.handler.HandlePostFlagEvent(handleCtx, event)
		cancel()
		if err == nil {
			continue
		}

		logger.Errorf("flag event handle failed for target=%s/%d attempt=%d: %v", event.TargetType, event.TargetId, event.Attempt, err)
		if event.Attempt+1 >= postFlagMaxAttempts {
			if deadErr := c.queue.deadLetter(context.Background(), event); deadErr != nil {
				logger.Errorf("flag event dead-letter failed for target=%s/%d: %v", event.TargetType, event.TargetId, deadErr)
			}
			continue
		}
		if retryErr := c.queue.requeue(context.Background(), event); retryErr != nil {
			logger.Errorf("flag event requeue failed for target=%s/%d: %v", event.TargetType, event.TargetId, retryErr)
		}
	}
}

type PostFlagProcessor struct {
	engine     *degradation.Engine
	paperModel *model.PaperModel
	cacheStore *redis.Redis
}

func NewPostFlagProcessor(
	engine *degradation.Engine,
	paperModel *model.PaperModel,
	cacheStore *redis.Redis,
) *PostFlagProcessor {
	return &PostFlagProcessor{
		engine:     engine,
		paperModel: paperModel,
		cacheStore: cacheStore,
	}
}

func (p *PostFlagProcessor) HandlePostFlagEvent(ctx context.Context, event PostFlagEvent) error {
	if err := event.Validate(); err != nil {
		return err
	}

	if _, err := p.engine.EvaluateDegradation(ctx, event.TargetType, event.TargetId); err != nil {
		return fmt.Errorf("evaluate target degradation: %w", err)
	}

	p.invalidateCaches(ctx, event.TargetType, event.TargetId)
	return nil
}

func (p *PostFlagProcessor) invalidateCaches(ctx context.Context, targetType string, targetId int64) {
	if p == nil || p.cacheStore == nil {
		return
	}

	keys := []string{
		cachekeys.FlagStatus(targetType, targetId),
	}

	if targetType == consts.FlagTargetPaper {
		keys = append(keys, cachekeys.PaperModeration(targetId), cachekeys.HotPapers(""))
		if p.paperModel != nil {
			paper, err := p.paperModel.FindByIdPrimary(ctx, targetId)
			if err == nil && paper.Zone != "" {
				keys = append(keys, cachekeys.HotPapers(paper.Zone))
			} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
				logx.WithContext(ctx).Errorf("load paper for post-flag cache invalidation failed for paper=%d: %v", targetId, err)
			}
		}
	}

	if _, err := p.cacheStore.DelCtx(ctx, keys...); err != nil {
		logx.WithContext(ctx).Errorf("invalidate post-flag caches failed for target=%s/%d: %v", targetType, targetId, err)
	}
}
