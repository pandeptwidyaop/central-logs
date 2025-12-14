package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(url string) (*RedisClient, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisClient{client: client}, nil
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}

func (r *RedisClient) Client() *redis.Client {
	return r.client
}

// Pub/Sub for realtime log streaming

func (r *RedisClient) PublishLog(ctx context.Context, projectID string, data interface{}) error {
	channel := fmt.Sprintf("logs:%s", projectID)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return r.client.Publish(ctx, channel, jsonData).Err()
}

func (r *RedisClient) SubscribeLogs(ctx context.Context, projectIDs ...string) *redis.PubSub {
	channels := make([]string, len(projectIDs))
	for i, id := range projectIDs {
		channels[i] = fmt.Sprintf("logs:%s", id)
	}
	return r.client.Subscribe(ctx, channels...)
}

func (r *RedisClient) PSubscribeLogs(ctx context.Context) *redis.PubSub {
	return r.client.PSubscribe(ctx, "logs:*")
}

// Rate Limiting

type RateLimiter struct {
	client *redis.Client
}

func NewRateLimiter(client *redis.Client) *RateLimiter {
	return &RateLimiter{client: client}
}

// Allow checks if the action is allowed within the rate limit
// Returns (allowed, remaining, resetTime)
func (rl *RateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, int, time.Time, error) {
	now := time.Now()
	windowKey := fmt.Sprintf("%s:%d", key, now.Unix()/int64(window.Seconds()))

	pipe := rl.client.Pipeline()
	incr := pipe.Incr(ctx, windowKey)
	pipe.Expire(ctx, windowKey, window)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, time.Time{}, err
	}

	count := int(incr.Val())
	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}

	resetTime := now.Truncate(window).Add(window)

	return count <= limit, remaining, resetTime, nil
}

// AllowChannel checks rate limit for notification channels
func (rl *RateLimiter) AllowChannel(ctx context.Context, channelID string, limit int) (bool, error) {
	key := fmt.Sprintf("ratelimit:channel:%s", channelID)
	allowed, _, _, err := rl.Allow(ctx, key, limit, time.Minute)
	return allowed, err
}

// AllowAPI checks rate limit for API requests per project
func (rl *RateLimiter) AllowAPI(ctx context.Context, projectID string, limit int) (bool, int, time.Time, error) {
	key := fmt.Sprintf("ratelimit:api:%s", projectID)
	return rl.Allow(ctx, key, limit, time.Minute)
}

// Notification Queue

type NotificationJob struct {
	LogID     string `json:"log_id"`
	ChannelID string `json:"channel_id"`
	ProjectID string `json:"project_id"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Source    string `json:"source"`
	Timestamp string `json:"timestamp"`
}

const notificationQueue = "notifications:queue"

func (r *RedisClient) EnqueueNotification(ctx context.Context, job *NotificationJob) error {
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return r.client.LPush(ctx, notificationQueue, data).Err()
}

func (r *RedisClient) DequeueNotification(ctx context.Context, timeout time.Duration) (*NotificationJob, error) {
	result, err := r.client.BRPop(ctx, timeout, notificationQueue).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var job NotificationJob
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		return nil, err
	}

	return &job, nil
}

func (r *RedisClient) GetQueueLength(ctx context.Context) (int64, error) {
	return r.client.LLen(ctx, notificationQueue).Result()
}
