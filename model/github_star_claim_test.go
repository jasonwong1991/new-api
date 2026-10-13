package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func setupGitHubStarClaimTest(t *testing.T) (*User, *User) {
	t.Helper()
	setupMigrationTestDB(t)
	require.NoError(t, migrateDB())

	alice := &User{Username: "alice", Password: "password123", AffCode: "a1", Quota: 100}
	bob := &User{Username: "bob", Password: "password123", AffCode: "b1", Quota: 0}
	require.NoError(t, DB.Create(alice).Error)
	require.NoError(t, DB.Create(bob).Error)
	return alice, bob
}

func userQuota(t *testing.T, id int) int {
	t.Helper()
	var u User
	require.NoError(t, DB.First(&u, id).Error)
	return u.Quota
}

func TestClaimGitHubStarRewardAwardsQuotaOncePerUserAndRepo(t *testing.T) {
	alice, _ := setupGitHubStarClaimTest(t)

	require.NoError(t, ClaimGitHubStarReward(alice.Id, "1001", "Owner/Repo", 500))
	require.Equal(t, 600, userQuota(t, alice.Id))

	claimed, err := IsGitHubStarClaimed(alice.Id, "9999", "owner/repo")
	require.NoError(t, err)
	require.True(t, claimed, "same user must count as claimed regardless of repo case or GitHub account")

	// Re-binding a different GitHub account must not allow a second claim for the same repo.
	require.Error(t, ClaimGitHubStarReward(alice.Id, "2002", "owner/repo", 500))
	require.Equal(t, 600, userQuota(t, alice.Id))

	// A different repo is a separate reward.
	require.NoError(t, ClaimGitHubStarReward(alice.Id, "1001", "owner/other", 500))
	require.Equal(t, 1100, userQuota(t, alice.Id))

	repos, err := GetClaimedGitHubStarRepos(alice.Id, "1001")
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"owner/repo", "owner/other"}, repos)
}

func TestClaimGitHubStarRewardRejectsSameGitHubAccountOnAnotherUser(t *testing.T) {
	alice, bob := setupGitHubStarClaimTest(t)

	require.NoError(t, ClaimGitHubStarReward(alice.Id, "1001", "owner/repo", 500))

	claimed, err := IsGitHubStarClaimed(bob.Id, "1001", "owner/repo")
	require.NoError(t, err)
	require.True(t, claimed, "GitHub account already used by another user must count as claimed")

	require.Error(t, ClaimGitHubStarReward(bob.Id, "1001", "owner/repo", 500))
	require.Equal(t, 0, userQuota(t, bob.Id))
}

func TestGetClaimedGitHubStarReposIncludesClaimsFromSameGitHubAccount(t *testing.T) {
	alice, bob := setupGitHubStarClaimTest(t)

	require.NoError(t, ClaimGitHubStarReward(alice.Id, "1001", "owner/repo", 500))
	repos, err := GetClaimedGitHubStarRepos(bob.Id, "1001")
	require.NoError(t, err)
	require.Equal(t, []string{"owner/repo"}, repos, "another user's claim on the same GitHub account must count as claimed")

	repos, err = GetClaimedGitHubStarRepos(bob.Id, "2002")
	require.NoError(t, err)
	require.Empty(t, repos)

	repos, err = GetClaimedGitHubStarRepos(bob.Id, "")
	require.NoError(t, err)
	require.Empty(t, repos, "unbound users only see their own claims")
}
