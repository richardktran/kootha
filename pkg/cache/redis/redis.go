package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *goredis.Client
}

func New(addr string) (*Client, error) {
	rdb := goredis.NewClient(&goredis.Options{
		Addr: addr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	return &Client{rdb: rdb}, nil
}

func (c *Client) Raw() *goredis.Client {
	return c.rdb
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.rdb.Get(ctx, key).Result()
}

func (c *Client) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.rdb.Del(ctx, keys...).Err()
}

func (c *Client) HSet(ctx context.Context, key string, values ...interface{}) error {
	return c.rdb.HSet(ctx, key, values...).Err()
}

func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.rdb.HGetAll(ctx, key).Result()
}

func (c *Client) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return c.rdb.SAdd(ctx, key, members...).Err()
}

func (c *Client) SCard(ctx context.Context, key string) (int64, error) {
	return c.rdb.SCard(ctx, key).Result()
}

func (c *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	return c.rdb.SMembers(ctx, key).Result()
}

func (c *Client) SRem(ctx context.Context, key string, members ...interface{}) error {
	return c.rdb.SRem(ctx, key, members...).Err()
}

func (c *Client) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return c.rdb.Expire(ctx, key, ttl).Err()
}

func (c *Client) ZAdd(ctx context.Context, key string, members ...goredis.Z) error {
	return c.rdb.ZAdd(ctx, key, members...).Err()
}

func (c *Client) ZIncrBy(ctx context.Context, key string, increment float64, member string) (float64, error) {
	return c.rdb.ZIncrBy(ctx, key, increment, member).Result()
}

func (c *Client) ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) ([]goredis.Z, error) {
	return c.rdb.ZRevRangeWithScores(ctx, key, start, stop).Result()
}

func (c *Client) Publish(ctx context.Context, channel string, message interface{}) error {
	return c.rdb.Publish(ctx, channel, message).Err()
}

func (c *Client) Subscribe(ctx context.Context, channels ...string) *goredis.PubSub {
	return c.rdb.Subscribe(ctx, channels...)
}

func (c *Client) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	return c.rdb.SetNX(ctx, key, value, ttl).Result()
}

func (c *Client) Close() error {
	return c.rdb.Close()
}

func SessionKey(sessionID string) string {
	return fmt.Sprintf("session:%s", sessionID)
}

func SessionConnsKey(sessionID string) string {
	return fmt.Sprintf("session:%s:conns", sessionID)
}

func LeaderboardKey(sessionID string) string {
	return fmt.Sprintf("leaderboard:%s", sessionID)
}

func AnswerDedupKey(sessionID, questionID, userID string) string {
	return fmt.Sprintf("answer:%s:%s:%s", sessionID, questionID, userID)
}

func QuestionAnswersKey(sessionID, questionID string) string {
	return fmt.Sprintf("answers:%s:%s", sessionID, questionID)
}

func QuestionRevealKey(sessionID, questionID string) string {
	return fmt.Sprintf("reveal:%s:%s", sessionID, questionID)
}

func SessionLockKey(sessionID, action string) string {
	return fmt.Sprintf("lock:session:%s:%s", sessionID, action)
}

func BankQuestionsKey() string {
	return "quizbank:questions"
}
