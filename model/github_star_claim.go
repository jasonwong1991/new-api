package model

import (
	"errors"
	"strings"
	"time"
)

// GitHubStarClaim Star GitHub 仓库奖励领取记录
// 唯一约束保证：同一用户每个仓库只能领一次，同一 GitHub 账号每个仓库也只能被领一次
type GitHubStarClaim struct {
	Id           int    `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId       int    `json:"user_id" gorm:"not null;uniqueIndex:idx_github_star_user_repo,priority:1"`
	GitHubId     string `json:"github_id" gorm:"column:github_id;type:varchar(32);not null;uniqueIndex:idx_github_star_github_repo,priority:1"`
	Repo         string `json:"repo" gorm:"type:varchar(140);not null;uniqueIndex:idx_github_star_user_repo,priority:2;uniqueIndex:idx_github_star_github_repo,priority:2"` // 小写 owner/repo
	QuotaAwarded int    `json:"quota_awarded" gorm:"not null"`
	CreatedAt    int64  `json:"created_at" gorm:"bigint"`
}

func (GitHubStarClaim) TableName() string {
	return "github_star_claims"
}

// GetClaimedGitHubStarRepos 获取已不可再领取奖励的仓库（小写）
// 与 IsGitHubStarClaimed 一致：用户本人或该 GitHub 账号领取过都算已领取
func GetClaimedGitHubStarRepos(userId int, githubId string) ([]string, error) {
	query := DB.Model(&GitHubStarClaim{}).Where("user_id = ?", userId)
	if githubId != "" {
		query = DB.Model(&GitHubStarClaim{}).Where("user_id = ? OR github_id = ?", userId, githubId)
	}
	var repos []string
	err := query.Pluck("repo", &repos).Error
	return repos, err
}

// IsGitHubStarClaimed 检查该用户或该 GitHub 账号是否已领取过指定仓库的奖励
func IsGitHubStarClaimed(userId int, githubId string, repo string) (bool, error) {
	var count int64
	err := DB.Model(&GitHubStarClaim{}).
		Where("repo = ? AND (user_id = ? OR github_id = ?)", strings.ToLower(repo), userId, githubId).
		Count(&count).Error
	return count > 0, err
}

// ClaimGitHubStarReward 写入领取记录并发放额度
// 唯一约束防止并发重复领取；发放额度失败时回滚领取记录
func ClaimGitHubStarReward(userId int, githubId string, repo string, quota int) error {
	claim := &GitHubStarClaim{
		UserId:       userId,
		GitHubId:     githubId,
		Repo:         strings.ToLower(repo),
		QuotaAwarded: quota,
		CreatedAt:    time.Now().Unix(),
	}
	if err := DB.Create(claim).Error; err != nil {
		return errors.New("领取失败，可能已领取过，请刷新后重试")
	}
	if err := IncreaseUserQuota(userId, quota, true); err != nil {
		DB.Delete(claim)
		return errors.New("领取失败：发放额度出错")
	}
	return nil
}
