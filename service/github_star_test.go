package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func withGitHubAPI(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	server := httptest.NewServer(handler)
	original := gitHubAPIBaseURL
	gitHubAPIBaseURL = server.URL
	t.Cleanup(func() {
		gitHubAPIBaseURL = original
		server.Close()
	})
}

func TestResolveGitHubAccountByNumericIdAndLegacyLogin(t *testing.T) {
	var gotAuth string
	withGitHubAPI(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		switch r.URL.Path {
		case "/user/42", "/users/octocat":
			fmt.Fprint(w, `{"id":42,"login":"octocat"}`)
		default:
			http.NotFound(w, r)
		}
	})

	account, err := ResolveGitHubAccount(context.Background(), "42", "secret")
	require.NoError(t, err)
	require.Equal(t, &GitHubAccount{Id: 42, Login: "octocat"}, account)
	require.Equal(t, "Bearer secret", gotAuth)

	account, err = ResolveGitHubAccount(context.Background(), "octocat", "")
	require.NoError(t, err)
	require.Equal(t, int64(42), account.Id)
	require.Empty(t, gotAuth)

	_, err = ResolveGitHubAccount(context.Background(), "7", "")
	require.ErrorIs(t, err, ErrGitHubAccountNotFound)
}

func TestHasStarredRepoPagesAndMatchesCaseInsensitively(t *testing.T) {
	withGitHubAPI(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/users/octocat/starred", r.URL.Path)
		if r.URL.Query().Get("page") == "1" {
			items := make([]string, gitHubStarredPerPage)
			for i := range items {
				items[i] = fmt.Sprintf(`{"full_name":"someone/repo%d"}`, i)
			}
			fmt.Fprint(w, "["+strings.Join(items, ",")+"]")
			return
		}
		fmt.Fprint(w, `[{"full_name":"JasonWong1991/WonderBox"}]`)
	})

	starred, err := HasStarredRepo(context.Background(), "octocat", "jasonwong1991/wonderbox", "")
	require.NoError(t, err)
	require.True(t, starred)

	starred, err = HasStarredRepo(context.Background(), "octocat", "jasonwong1991/easy_proxies", "")
	require.NoError(t, err)
	require.False(t, starred)
}

func TestHasStarredRepoReportsRateLimit(t *testing.T) {
	withGitHubAPI(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})

	_, err := HasStarredRepo(context.Background(), "octocat", "owner/repo", "")
	require.ErrorIs(t, err, ErrGitHubRateLimited)
}
