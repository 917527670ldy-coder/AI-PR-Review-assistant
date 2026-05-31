package queue

import (
	"context"
	"testing"
	"time"
)

// TestRedisQueueConnection 测试 Redis 连接
func TestRedisQueueConnection(t *testing.T) {
	q, err := NewRedisQueue("localhost:6379", "test_queue")
	if err != nil {
		t.Fatalf("连接 Redis 失败: %v", err)
	}
	defer q.Close()

	t.Log("✅ Redis 连接成功")
}

// TestEnqueueAndDequeue 测试入队和出队
func TestEnqueueAndDequeue(t *testing.T) {
	// 创建队列
	q, err := NewRedisQueue("localhost:6379", "test_queue_fifo")
	if err != nil {
		t.Fatalf("连接 Redis 失败: %v", err)
	}
	defer q.Close()

	// 清空队列
	ctx := context.Background()
	if err := q.Clear(ctx); err != nil {
		t.Logf("清空队列失败: %v", err)
	}

	// 创建测试任务
	task1 := NewPRReviewTask(
		"owner1", "repo1", 1,
		"sha1", "opened", "event1",
	)
	task2 := NewPRReviewTask(
		"owner2", "repo2", 2,
		"sha2", "synchronize", "event2",
	)

	// 测试入队
	if err := q.Enqueue(ctx, task1); err != nil {
		t.Fatalf("入队任务1失败: %v", err)
	}
	t.Logf("✅ 任务1已入队: %s", task1.String())

	if err := q.Enqueue(ctx, task2); err != nil {
		t.Fatalf("入队任务2失败: %v", err)
	}
	t.Logf("✅ 任务2已入队: %s", task2.String())

	// 测试队列大小
	size, err := q.Size(ctx)
	if err != nil {
		t.Fatalf("获取队列大小失败: %v", err)
	}
	if size != 2 {
		t.Fatalf("队列大小应为2，实际为: %d", size)
	}
	t.Logf("✅ 队列大小正确: %d", size)

	// 测试出队（FIFO）
	dequeueCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	retrievedTask1, err := q.Dequeue(dequeueCtx, 5*time.Second)
	if err != nil {
		t.Fatalf("出队任务1失败: %v", err)
	}
	if retrievedTask1 == nil {
		t.Fatal("未取到任务1")
	}

	// 验证任务1数据
	if retrievedTask1.ID != task1.ID {
		t.Errorf("任务1 ID不匹配: 期望 %s, 实际 %s", task1.ID, retrievedTask1.ID)
	}
	if retrievedTask1.Owner != task1.Owner {
		t.Errorf("任务1 Owner不匹配: 期望 %s, 实际 %s", task1.Owner, retrievedTask1.Owner)
	}
	t.Logf("✅ 任务1已出队并验证: %s", retrievedTask1.String())

	retrievedTask2, err := q.Dequeue(dequeueCtx, 5*time.Second)
	if err != nil {
		t.Fatalf("出队任务2失败: %v", err)
	}
	if retrievedTask2 == nil {
		t.Fatal("未取到任务2")
	}

	// 验证任务2数据
	if retrievedTask2.ID != task2.ID {
		t.Errorf("任务2 ID不匹配: 期望 %s, 实际 %s", task2.ID, retrievedTask2.ID)
	}
	t.Logf("✅ 任务2已出队并验证: %s", retrievedTask2.String())

	// 验证队列为空
	size, err = q.Size(ctx)
	if err != nil {
		t.Fatalf("获取队列大小失败: %v", err)
	}
	if size != 0 {
		t.Fatalf("队列应为空，实际大小: %d", size)
	}
	t.Log("✅ 队列已空")
}

// TestEmptyQueueTimeout 测试空队列超时
func TestEmptyQueueTimeout(t *testing.T) {
	q, err := NewRedisQueue("localhost:6379", "test_queue_timeout")
	if err != nil {
		t.Fatalf("连接 Redis 失败: %v", err)
	}
	defer q.Close()

	// 清空队列
	ctx := context.Background()
	if err := q.Clear(ctx); err != nil {
		t.Logf("清空队列失败: %v", err)
	}

	// 测试空队列超时
	dequeueCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	start := time.Now()
	emptyTask, err := q.Dequeue(dequeueCtx, 2*time.Second)
	elapsed := time.Since(start)

	if emptyTask != nil {
		t.Fatal("空队列应返回 nil")
	}

	if elapsed < 1*time.Second {
		t.Fatalf("应等待超时，实际等待: %v", elapsed)
	}

	t.Logf("✅ 空队列正确超时，等待时间: %v", elapsed)
}

// TestTaskSerialization 测试任务序列化
func TestTaskSerialization(t *testing.T) {
	task := NewPRReviewTask(
		"gin-gonic", "gin", 123,
		"abc123def456", "opened", "event-001",
	)

	// 验证任务字段
	if task.Owner != "gin-gonic" {
		t.Errorf("Owner不匹配: 期望 gin-gonic, 实际 %s", task.Owner)
	}
	if task.Repo != "gin" {
		t.Errorf("Repo不匹配: 期望 gin, 实际 %s", task.Repo)
	}
	if task.PRNumber != 123 {
		t.Errorf("PRNumber不匹配: 期望 123, 实际 %d", task.PRNumber)
	}
	if task.SHA != "abc123def456" {
		t.Errorf("SHA不匹配: 期望 abc123def456, 实际 %s", task.SHA)
	}

	t.Logf("✅ 任务序列化正确: %s", task.String())
}