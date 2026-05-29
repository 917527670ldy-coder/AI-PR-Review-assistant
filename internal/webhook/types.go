package webhook

type PullRequestEvent struct {
    Action       string `json:"action"`
    Number       int    `json:"number"`
    PullRequest  struct {
        Number int `json:"number"`
        Head   struct {
            SHA string `json:"sha"`
        } `json:"head"`
    } `json:"pull_request"`
    Repository struct {
        FullName string `json:"full_name"`
    } `json:"repository"`
}
