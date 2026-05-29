package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v62/github"
	"golang.org/x/oauth2"
)

// Client GitHub API 客户端
type Client struct {
	client *github.Client
	ctx    context.Context
}

// NewClient 创建新的 GitHub 客户端
func NewClient(token string) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	return &Client{
		client: client,
		ctx:    ctx,
	}
}

// PRInfo PR 信息
type PRInfo struct {
	Owner    string
	Repo     string
	Number   int
	SHA      string
	Title    string
	Body     string
	User     string
	BaseRef  string
	HeadRef  string
}

// GetPR 获取 PR 基本信息
func (c *Client) GetPR(owner, repo string, number int) (*PRInfo, error) {
	pr, _, err := c.client.PullRequests.Get(c.ctx, owner, repo, number)
	if err != nil {
		return nil, fmt.Errorf("获取 PR 失败: %w", err)
	}

	return &PRInfo{
		Owner:   owner,
		Repo:    repo,
		Number:  number,
		SHA:     pr.GetHead().GetSHA(),
		Title:   pr.GetTitle(),
		Body:    pr.GetBody(),
		User:    pr.GetUser().GetLogin(),
		BaseRef: pr.GetBase().GetRef(),
		HeadRef: pr.GetHead().GetRef(),
	}, nil
}

// FileDiff 文件变更信息
type FileDiff struct {
	Filename string
	Status   string // "added", "modified", "removed", "renamed"
	Additions int
	Deletions int
	Changes   int
	Patch     string // 具体的代码变更（diff 格式）
}

// GetPRDiff 获取 PR 的代码变更
func (c *Client) GetPRDiff(owner, repo string, number int) ([]FileDiff, error) {
	// 获取 PR 的文件列表
	opts := &github.ListOptions{
		PerPage: 100,
	}

	files, _, err := c.client.PullRequests.ListFiles(c.ctx, owner, repo, number, opts)
	if err != nil {
		return nil, fmt.Errorf("获取 PR 文件列表失败: %w", err)
	}

	var diffs []FileDiff
	for _, file := range files {
		diffs = append(diffs, FileDiff{
			Filename:  file.GetFilename(),
			Status:    file.GetStatus(),
			Additions: file.GetAdditions(),
			Deletions: file.GetDeletions(),
			Changes:   file.GetChanges(),
			Patch:     file.GetPatch(),
		})
	}

	return diffs, nil
}

// ReviewComment PR 评审评论
type ReviewComment struct {
	Path     string // 文件路径
	Position int    // 在 diff 中的位置（行号）
	Body     string // 评论内容
}

// CreatePRReview 创建 PR 评审
func (c *Client) CreatePRReview(owner, repo string, number int, event string, body string, comments []ReviewComment) error {
	// 构建评审请求
	reviewRequest := &github.PullRequestReviewRequest{
		Event: &event,
		Body:  &body,
	}

	// 添加行内评论
	if len(comments) > 0 {
		var reviewComments []*github.DraftReviewComment
		for _, comment := range comments {
			reviewComments = append(reviewComments, &github.DraftReviewComment{
				Path:     &comment.Path,
				Position: &comment.Position,
				Body:     &comment.Body,
			})
		}
		reviewRequest.Comments = reviewComments
	}

	_, _, err := c.client.PullRequests.CreateReview(c.ctx, owner, repo, number, reviewRequest)
	if err != nil {
		return fmt.Errorf("创建 PR 评审失败: %w", err)
	}

	return nil
}

// CreatePRComment 在 PR 上创建普通评论
func (c *Client) CreatePRComment(owner, repo string, number int, body string) error {
	comment := &github.IssueComment{
		Body: &body,
	}

	_, _, err := c.client.Issues.CreateComment(c.ctx, owner, repo, number, comment)
	if err != nil {
		return fmt.Errorf("创建 PR 评论失败: %w", err)
	}

	return nil
}

// GetFileContent 获取仓库文件内容
func (c *Client) GetFileContent(owner, repo, path, ref string) (string, error) {
	// 获取文件内容
	opts := &github.RepositoryContentGetOptions{
		Ref: ref,
	}

	content, _, _, err := c.client.Repositories.GetContents(c.ctx, owner, repo, path, opts)
	if err != nil {
		return "", fmt.Errorf("获取文件内容失败: %w", err)
	}

	if content == nil {
		return "", fmt.Errorf("文件不存在或不是文件类型")
	}

	return content.GetContent()
}
