package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisQueue 基于 Redis 的任务队列实现
// 使用 Redis List 数据结构作为队列
type RedisQueue struct {
	client    *redis.Client
	queueName string // 队列名称（Redis key）
}

// NewRedisQueue 创建 Redis 队列实例
// addr: Redis 地址，如 "localhost:6379"
// queueName: 队列名称，如 "pr_review_tasks"
func NewRedisQueue(addr string, queueName string) (*RedisQueue, error) {
	// 创建 Redis 客户端
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("连接 Redis 失败: %w", err)
	}

	return &RedisQueue{
		client:    client,
		queueName: queueName,
	}, nil
}

// Enqueue 将任务添加到队列
// 使用 LPUSH 将任务推入队列左侧（头部）
func (q *RedisQueue) Enqueue(ctx context.Context, task *Task) error {
	// 将任务序列化为 JSON
	data, err := task.ToJSON()
	if err != nil {
		return fmt.Errorf("序列化任务失败: %w", err)
	}

	// 使用 LPUSH 添加到队列
	// LPUSH 返回队列长度，如果成功应该 > 0
	result, err := q.client.LPush(ctx, q.queueName, data).Result()
	if err != nil {
		return fmt.Errorf("添加任务到队列失败: %w", err)
	}

	if result == 0 {
		return fmt.Errorf("添加任务失败：队列为空")
	}

	return nil
}

// Dequeue 从队列取出一个任务
// 使用 BRPOP 阻塞式弹出队列右侧（尾部）的元素
// timeout 为 0 表示无限等待
func (q *RedisQueue) Dequeue(ctx context.Context, timeout time.Duration) (*Task, error) {
	// BRPOP 会阻塞等待直到有数据或超时
	// 返回格式：[queueName, value]
	var popTimeout time.Duration
	if timeout == 0 {
		popTimeout = 0 // 0 表示无限等待
	} else {
		popTimeout = timeout
	}

	result, err := q.client.BRPop(ctx, popTimeout, q.queueName).Result()
	if err != nil {
		if err == redis.Nil {
			// 超时，队列中没有任务
			return nil, nil
		}
		return nil, fmt.Errorf("从队列取出任务失败: %w", err)
	}

	// result[0] 是队列名，result[1] 是数据
	if len(result) < 2 {
		return nil, fmt.Errorf("队列返回数据格式错误")
	}

	// 反序列化任务
	task, err := FromJSON(result[1])
	if err != nil {
		return nil, fmt.Errorf("反序列化任务失败: %w", err)
	}

	return task, nil
}

// Size 获取队列当前大小（待处理任务数量）
func (q *RedisQueue) Size(ctx context.Context) (int64, error) {
	size, err := q.client.LLen(ctx, q.queueName).Result()
	if err != nil {
		return 0, fmt.Errorf("获取队列大小失败: %w", err)
	}
	return size, nil
}

// Close 关闭 Redis 连接
func (q *RedisQueue) Close() error {
	if err := q.client.Close(); err != nil {
		return fmt.Errorf("关闭 Redis 连接失败: %w", err)
	}
	return nil
}

// Clear 清空队列（用于测试）
func (q *RedisQueue) Clear(ctx context.Context) error {
	return q.client.Del(ctx, q.queueName).Err()
}

// GetClient 获取 Redis 客户端（用于测试）
func (q *RedisQueue) GetClient() *redis.Client {
	return q.client
}
