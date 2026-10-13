package controller

import (
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/gin-gonic/gin"
)

type gitHubStarRepoStatus struct {
	Repo    string `json:"repo"`
	Claimed bool   `json:"claimed"`
}

type gitHubStarClaimRequest struct {
	Repo string `json:"repo"`
}

func GetGitHubStarStatus(c *gin.Context) {
	cfg := operation_setting.GetGitHubStarSetting()
	userId := c.GetInt("id")
	user, err := model.GetUserById(userId, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	// 与领取奖励使用同一账号标识，否则领取记录无法对应
	githubId := ""
	if user.GitHubId != "" {
		_, githubId, err = service.PrepareGitHubStarAccount(c.Request.Context(), user.GitHubId, cfg.ApiKey)
		if err != nil {
			common.ApiError(c, err)
			return
		}
	}
	claimedRepos, err := model.GetClaimedGitHubStarRepos(userId, githubId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	claimed := make(map[string]bool, len(claimedRepos))
	for _, r := range claimedRepos {
		claimed[r] = true
	}
	repos := make([]gitHubStarRepoStatus, 0, len(cfg.Repos))
	for _, r := range cfg.Repos {
		repos = append(repos, gitHubStarRepoStatus{Repo: r, Claimed: claimed[strings.ToLower(r)]})
	}
	common.ApiSuccess(c, gin.H{
		"enabled":        cfg.Enabled,
		"quota_per_repo": cfg.QuotaPerRepo,
		"github_bound":   user.GitHubId != "",
		"repos":          repos,
	})
}

func ClaimGitHubStar(c *gin.Context) {
	cfg := operation_setting.GetGitHubStarSetting()
	if !cfg.Enabled {
		common.ApiErrorMsg(c, "Star 奖励活动未开启")
		return
	}
	if cfg.QuotaPerRepo <= 0 {
		common.ApiErrorMsg(c, "管理员尚未配置有效的奖励额度")
		return
	}
	var req gitHubStarClaimRequest
	if err := common.DecodeJson(c.Request.Body, &req); err != nil {
		common.ApiErrorMsg(c, "无效的参数")
		return
	}
	repo, ok := cfg.FindRepo(req.Repo)
	if !ok {
		common.ApiErrorMsg(c, "该仓库不在 Star 奖励活动中")
		return
	}

	userId := c.GetInt("id")
	user, err := model.GetUserById(userId, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if user.GitHubId == "" {
		common.ApiErrorMsg(c, "请先绑定 GitHub 账号")
		return
	}

	account, githubId, err := service.PrepareGitHubStarAccount(c.Request.Context(), user.GitHubId, cfg.ApiKey)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	claimed, err := model.IsGitHubStarClaimed(userId, githubId, repo)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if claimed {
		common.ApiErrorMsg(c, "你或该 GitHub 账号已领取过该仓库的奖励")
		return
	}

	starred, err := service.HasStarredRepo(c.Request.Context(), account.Login, repo, cfg.ApiKey)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if !starred {
		common.ApiErrorMsg(c, fmt.Sprintf("未检测到 GitHub 账号 %s Star 了 %s，请 Star 后再试", account.Login, repo))
		return
	}

	if err := model.ClaimGitHubStarReward(userId, githubId, repo, cfg.QuotaPerRepo); err != nil {
		common.ApiError(c, err)
		return
	}
	model.RecordLog(userId, model.LogTypeSystem, fmt.Sprintf("Star GitHub 仓库 %s 获得奖励 %s", repo, logger.LogQuota(cfg.QuotaPerRepo)))
	common.ApiSuccess(c, gin.H{
		"repo":  repo,
		"quota": cfg.QuotaPerRepo,
	})
}
