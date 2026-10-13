package operation_setting

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGitHubStarSettingFindRepo(t *testing.T) {
	s := &GitHubStarSetting{Repos: []string{" JasonWong1991/WonderBox ", "", "jasonwong1991/easy_proxies"}}

	repo, ok := s.FindRepo("jasonwong1991/wonderbox")
	require.True(t, ok)
	require.Equal(t, "JasonWong1991/WonderBox", repo)

	_, ok = s.FindRepo("someone/else")
	require.False(t, ok)

	_, ok = s.FindRepo("")
	require.False(t, ok)
}
