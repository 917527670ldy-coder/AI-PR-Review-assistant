# AI PR Review Assistant

一个基于 Go 的 AI PR 自动评审工具。  
当 GitHub PR 事件触发后，系统会自动拉取 PR 信息和代码变更，调用 AI 分析并将评审结果评论回 PR。

## 功能概览

- 接收 GitHub `pull_request` Webhook
- 校验 Webhook 签名（HMAC SHA256）
- 将任务写入 Redis 队列
- Worker 异步消费任务并执行评审流程
- 调用 GitHub API 获取 PR 信息和 Diff
- 调用 AI 模型生成评审结论、风险和建议
- 自动发布评审评论到 PR

## 技术架构

- Web 框架: Gin
- 队列: Redis List
- 代码托管集成: GitHub API
- AI 分析: 自定义 Analyzer（通过 `CLAUDE_API_KEY` 和 `AI_MODEL` 配置）

核心流程:

1. GitHub 发送 `pull_request` 事件到 `/webhook/github`
2. 服务验签并创建评审任务
3. 任务入 Redis 队列
4. Worker 出队处理
5. 拉取 PR + Diff，调用 AI 分析
6. 将评审结果评论回 PR

## 目录结构

```text
cmd/api            # 主程序入口（API + Worker）
internal/server    # HTTP 服务与路由
internal/webhook   # Webhook 解析与验签
internal/queue     # Redis 队列实现
internal/worker    # 任务消费与评审编排
internal/github    # GitHub API 客户端
internal/ai        # AI 分析器
```

## 环境要求

- Go 1.22+（建议）
- Docker（用于快速启动 Redis/Postgres）
- GitHub Token（可访问目标仓库并可评论 PR）
- 可用的 AI API Key

## 配置说明

复制配置模板:

```powershell
Copy-Item .env.example .env
```

`.env` 关键项:

```env
SERVER_PORT=8081
GITHUB_TOKEN=your_github_token
GITHUB_WEBHOOK_SECRET=your_webhook_secret
CLAUDE_API_KEY=your_ai_api_key
AI_MODEL=qwen-plus
REDIS_URL=localhost:6379
DATABASE_URL=postgres://postgres:postgres@localhost:5432/xengineer?sslmode=disable
```

注意:

- `REDIS_URL` 请使用 `host:port`（如 `localhost:6379`），不要写成 `redis://...`
- `GITHUB_WEBHOOK_SECRET` 需与 GitHub Webhook 的 Secret 完全一致

## 快速启动

1. 启动依赖服务（至少 Redis）:

```powershell
docker compose up -d redis
```

2. 启动项目:

```powershell
go run .\cmd\api
```

或:

```powershell
make run
```

3. 健康检查:

```powershell
curl http://localhost:8081/health
```

返回:

```json
{"status":"ok"}
```

## 本地 Webhook 自测（不依赖 ngrok）

在新终端执行（将 `yourname/yourrepo`、`123` 改为你的仓库和 PR 编号）:

```powershell
$secret = "your_webhook_secret"
$body = '{"action":"opened","number":123,"pull_request":{"number":123,"head":{"sha":"abc123"}},"repository":{"full_name":"yourname/yourrepo"}}'
$hmac = New-Object System.Security.Cryptography.HMACSHA256
$hmac.Key = [Text.Encoding]::UTF8.GetBytes($secret)
$hash = ($hmac.ComputeHash([Text.Encoding]::UTF8.GetBytes($body)) | ForEach-Object { $_.ToString("x2") }) -join ""
$sig = "sha256=$hash"

Invoke-RestMethod -Method Post `
  -Uri "http://localhost:8081/webhook/github" `
  -Headers @{
    "X-GitHub-Event"="pull_request"
    "X-Hub-Signature-256"=$sig
    "X-GitHub-Delivery"=[guid]::NewGuid().ToString()
    "Content-Type"="application/json"
  } `
  -Body $body
```

成功返回示例:

```text
status   : accepted
task_id  : pr_review-xxx
repo     : yourname/yourrepo
pr_number: 123
action   : opened
```

## GitHub 联调（ngrok）

1. 启动本地服务（`go run .\cmd\api`）
2. 启动隧道:

```powershell
ngrok http 8081
```

3. 复制 ngrok `Forwarding` 的 HTTPS 地址，例如:

```text
https://xxxx.ngrok-free.dev
```

4. 在 GitHub 仓库配置 Webhook:

- 路径: `Settings -> Webhooks -> Add webhook`
- `Payload URL`: `https://xxxx.ngrok-free.dev/webhook/github`
- `Content type`: `application/json`
- `Secret`: 与 `.env` 中 `GITHUB_WEBHOOK_SECRET` 一致
- Events: 选择 `Pull requests`

5. 对 PR 推送一次新 commit，观察:

- GitHub `Recent Deliveries` 是否为 `200/202`
- 本地日志是否完成“拉 PR -> 拉 Diff -> AI 分析 -> 发布评论”

## 常见问题排查

- ngrok 认证失败 `ERR_NGROK_4018` / `ERR_NGROK_105`  
  原因: authtoken 不正确或未配置。  
  处理: 使用 `ngrok config add-authtoken <your_authtoken>`，注意 API Key 不能替代 authtoken。

- 服务启动报端口占用  
  报错: `listen tcp :8081: bind...`  
  处理: 释放占用进程或改 `SERVER_PORT`。

- Webhook 返回 `accepted` 但最终失败  
  常见是发布评论时 GitHub 返回 `403`（例如目标 PR 评论受限）。  
  处理: 换你自己仓库中的普通 PR 进行联调。

- Redis 连接失败  
  检查 Redis 是否启动，`REDIS_URL` 是否为 `localhost:6379` 这种格式。



