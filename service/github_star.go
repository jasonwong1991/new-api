package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
)

var gitHubAPIBaseURL = "https://api.github.com"

const (
	gitHubStarredPerPage  = 100
	gitHubStarredMaxPages = 10 // Star 列表按时间倒序，最多检查最近 1000 个 Star
)

var (
	ErrGitHubAccountNotFound = errors.New("绑定的 GitHub 账号不存在")
	ErrGitHubRateLimited     = errors.New("GitHub 接口请求过于频繁，请稍后再试")
)

type GitHubAccount struct {
	Id    int64  `json:"id"`
	Login string `json:"login"`
}

func gitHubGet(ctx context.Context, path string, apiKey string, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, gitHubAPIBaseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	client := http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("连接 GitHub 失败: %w", err)
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
		return common.DecodeJson(res.Body, v)
	case http.StatusNotFound:
		return ErrGitHubAccountNotFound
	case http.StatusForbidden, http.StatusTooManyRequests:
		return ErrGitHubRateLimited
	default:
		return fmt.Errorf("GitHub 接口返回异常状态码 %d", res.StatusCode)
	}
}

// ResolveGitHubAccount 将用户绑定的 github_id（数字 ID，或旧版本保存的用户名）解析为当前的 GitHub 账号
func ResolveGitHubAccount(ctx context.Context, githubId string, apiKey string) (*GitHubAccount, error) {
	path := "/users/" + url.PathEscape(githubId)
	if _, err := strconv.ParseInt(githubId, 10, 64); err == nil {
		path = "/user/" + githubId
	}
	var account GitHubAccount
	if err := gitHubGet(ctx, path, apiKey, &account); err != nil {
		return nil, err
	}
	if account.Id == 0 || account.Login == "" {
		return nil, ErrGitHubAccountNotFound
	}
	return &account, nil
}

// HasStarredRepo 检查 GitHub 用户是否 Star 了指定仓库（owner/repo，不区分大小写）
func HasStarredRepo(ctx context.Context, login string, repo string, apiKey string) (bool, error) {
	for page := 1; page <= gitHubStarredMaxPages; page++ {
		var starred []struct {
			FullName string `json:"full_name"`
		}
		path := fmt.Sprintf("/users/%s/starred?per_page=%d&page=%d", url.PathEscape(login), gitHubStarredPerPage, page)
		if err := gitHubGet(ctx, path, apiKey, &starred); err != nil {
			return false, err
		}
		for _, r := range starred {
			if strings.EqualFold(r.FullName, repo) {
				return true, nil
			}
		}
		if len(starred) < gitHubStarredPerPage {
			return false, nil
		}
	}
	return false, nil
}

// PrepareGitHubStarAccount 解析用户绑定的 GitHub 账号并返回数字 ID
// 领取奖励和查询领取状态都必须使用同一个数字 ID，否则两边的记录无法对应
// （旧版绑定保存的是登录名，数字 ID 才是稳定标识）
func PrepareGitHubStarAccount(ctx context.Context, githubId string, apiKey string) (*GitHubAccount, string, error) {
	account, err := ResolveGitHubAccount(ctx, githubId, apiKey)
	if err != nil {
		return nil, "", err
	}
	return account, strconv.FormatInt(account.Id, 10), nil
}
