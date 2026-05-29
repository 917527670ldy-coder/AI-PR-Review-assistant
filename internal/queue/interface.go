package queue

import (
	"context"
	"time"
)

// Queue 任务队列接口
// 定义了任务队列的基本操作

//为什么不直接写一个RedisQueue实现这个接口的目的：
// 为了提高代码的可测试性和可维护性，通过接口抽象可以更容易地替换不同的队列实现；并且可以创建一个内存队列实现用于单元测试，避免在测试中依赖外部服务（如 Redis），从而提高测试的可靠性和速度。

type Queue interface {
	// Enqueue 将任务添加到队列
	// 返回错误如果添加失败（如 Redis 连接问题）
	Enqueue(ctx context.Context, task *Task) error

	// Dequeue 从队列取出一个任务
	// 如果队列为空，会阻塞等待直到有任务或超时
	// timeout 为 0 表示无限等待
	Dequeue(ctx context.Context, timeout time.Duration) (*Task, error)

	// Size 获取队列当前大小（待处理任务数量）
	Size(ctx context.Context) (int64, error)

	// Close 关闭队列连接
	Close() error
}

// QueueStats 队列统计信息
type QueueStats struct {
	TotalTasks     int64  // 总任务数
	PendingTasks   int64  // 待处理任务数
	CompletedTasks int64  // 已完成任务数
	FailedTasks    int64  // 失败任务数
}