package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"xengineer/internal/queue"
)

func main() {
	fmt.Println("=== Redis 队列功能测试 ===")
	fmt.Println()

	// 1. 创建 Redis 队列
	fmt.Println("步骤 1: 连接 Redis...")
	q, err := queue.NewRedisQueue("localhost:6379", "test_pr_review_tasks")
	if err != nil {
		log.Fatalf("❌ 连接 Redis 失败: %v", err)
	}
	defer q.Close()
	fmt.Println("✅ Redis 连接成功")
	fmt.Println()

	// 2. 清空队列（确保测试环境干净）
	fmt.Println("步骤 2: 清空测试队列...")
	ctx := context.Background()
	if err := q.Clear(ctx); err != nil {
		log.Printf("⚠️  清空队列失败: %v", err)
	} else {
		fmt.Println("✅ 队列已清空")
	}
	fmt.Println()

	// 3. 测试入队
	fmt.Println("步骤 3: 测试入队...")
	task1 := queue.NewPRReviewTask(
		"gin-gonic", "gin", 123,
		"abc123def456", "opened", "event-001",
	)
	task2 := queue.NewPRReviewTask(
		"mongodb", "mongo-go-driver", 456,
		"def789ghi012", "synchronize", "event-002",
	)

	if err := q.Enqueue(ctx, task1); err != nil {
		log.Fatalf("❌ 入队任务1失败: %v", err)
	}
	fmt.Printf("✅ 任务1已入队: %s\n", task1.String())

	if err := q.Enqueue(ctx, task2); err != nil {
		log.Fatalf("❌ 入队任务2失败: %v", err)
	}
	fmt.Printf("✅ 任务2已入队: %s\n", task2.String())
	fmt.Println()

	// 4. 查看队列大小
	fmt.Println("步骤 4: 查看队列大小...")
	size, err := q.Size(ctx)
	if err != nil {
		log.Fatalf("❌ 获取队列大小失败: %v", err)
	}
	fmt.Printf("✅ 当前队列大小: %d\n", size)
	fmt.Println()

	// 5. 测试出队
	fmt.Println("步骤 5: 测试出队...")
	dequeueCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 取出第一个任务
	retrievedTask1, err := q.Dequeue(dequeueCtx, 5*time.Second)
	if err != nil {
		log.Fatalf("❌ 出队任务1失败: %v", err)
	}
	if retrievedTask1 == nil {
		log.Fatal("❌ 未取到任务1")
	}
	fmt.Printf("✅ 任务1已出队: %s\n", retrievedTask1.String())
	verifyTask(task1, retrievedTask1)

	// 取出第二个任务
	retrievedTask2, err := q.Dequeue(dequeueCtx, 5*time.Second)
	if err != nil {
		log.Fatalf("❌ 出队任务2失败: %v", err)
	}
	if retrievedTask2 == nil {
		log.Fatal("❌ 未取到任务2")
	}
	fmt.Printf("✅ 任务2已出队: %s\n", retrievedTask2.String())
	verifyTask(task2, retrievedTask2)
	fmt.Println()

	// 6. 验证队列为空
	fmt.Println("步骤 6: 验证队列为空...")
	size, err = q.Size(ctx)
	if err != nil {
		log.Fatalf("❌ 获取队列大小失败: %v", err)
	}
	if size == 0 {
		fmt.Println("✅ 队列已空")
	} else {
		log.Fatalf("❌ 队列大小应为0，实际为: %d", size)
	}
	fmt.Println()

	// 7. 测试超时出队
	fmt.Println("步骤 7: 测试空队列超时...")
	start := time.Now()
	emptyTask, err := q.Dequeue(dequeueCtx, 2*time.Second)
	elapsed := time.Since(start)

	if emptyTask != nil {
		log.Fatal("❌ 空队列应该返回 nil")
	}
	if elapsed < 1*time.Second {
		log.Fatalf("❌ 应该等待超时，实际等待: %v", elapsed)
	}
	fmt.Printf("✅ 空队列正确超时，等待时间: %v\n", elapsed)
	fmt.Println()

	// 测试总结
	fmt.Println("================================")
	fmt.Println("✅ 所有测试通过！")
	fmt.Println("================================")
	fmt.Println()
	fmt.Println("测试内容总结:")
	fmt.Println("  1. ✅ Redis 连接正常")
	fmt.Println("  2. ✅ 入队功能正常（LPUSH）")
	fmt.Println("  3. ✅ 出队功能正常（BRPOP）")
	fmt.Println("  4. ✅ 队列大小查询正常")
	fmt.Println("  5. ✅ FIFO 顺序正确")
	fmt.Println("  6. ✅ 任务序列化/反序列化正确")
	fmt.Println("  7. ✅ 空队列超时处理正确")
}

// verifyTask 验证任务数据是否正确
func verifyTask(expected, actual *queue.Task) {
	if expected.ID != actual.ID {
		log.Fatalf("❌ 任务ID不匹配: 期望 %s, 实际 %s", expected.ID, actual.ID)
	}
	if expected.Owner != actual.Owner {
		log.Fatalf("❌ Owner不匹配: 期望 %s, 实际 %s", expected.Owner, actual.Owner)
	}
	if expected.Repo != actual.Repo {
		log.Fatalf("❌ Repo不匹配: 期望 %s, 实际 %s", expected.Repo, actual.Repo)
	}
	if expected.PRNumber != actual.PRNumber {
		log.Fatalf("❌ PRNumber不匹配: 期望 %d, 实际 %d", expected.PRNumber, actual.PRNumber)
	}
	if expected.SHA != actual.SHA {
		log.Fatalf("❌ SHA不匹配: 期望 %s, 实际 %s", expected.SHA, actual.SHA)
	}
	fmt.Println("   └─ 数据验证通过 ✓")
}