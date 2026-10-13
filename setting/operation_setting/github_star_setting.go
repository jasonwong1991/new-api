package operation_setting

import (
	"strings"

	"github.com/QuantumNous/new-api/setting/config"
)

// GitHubStarSetting Star GitHub 仓库领取额度奖励配置
type GitHubStarSetting struct {
	Enabled      bool     `json:"enabled"`        // 是否启用 Star 奖励
	Repos        []string `json:"repos"`          // 参与活动的仓库，格式 owner/repo，每个仓库可单独领取一次
	QuotaPerRepo int      `json:"quota_per_repo"` // 每个仓库的奖励额度
	ApiKey       string   `json:"api_key"`        // 可选 GitHub Token，将 GitHub API 限额从 60 次/小时提升到 5000 次/小时
}

// 默认配置
var gitHubStarSetting = GitHubStarSetting{
	Enabled:      false,
	Repos:        []string{},
	QuotaPerRepo: 50000000, // 默认 QuotaPerUnit 下约 100 USD
	ApiKey:       "",
}

func init() {
	config.GlobalConfig.Register("github_star_setting", &gitHubStarSetting)
}

// GetGitHubStarSetting 获取 Star 奖励配置
func GetGitHubStarSetting() *GitHubStarSetting {
	return &gitHubStarSetting
}

// FindRepo 按不区分大小写的方式查找活动仓库，返回配置中的仓库名
func (s *GitHubStarSetting) FindRepo(repo string) (string, bool) {
	repo = strings.TrimSpace(repo)
	for _, r := range s.Repos {
		if r = strings.TrimSpace(r); r != "" && strings.EqualFold(r, repo) {
			return r, true
		}
	}
	return "", false
}
