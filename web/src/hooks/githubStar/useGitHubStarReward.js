/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import { useCallback, useContext, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { StatusContext } from '../../context/Status';
import { UserContext } from '../../context/User';
import {
  API,
  onGitHubOAuthClicked,
  renderQuota,
  showError,
  showSuccess,
} from '../../helpers';

const sameRepo = (a, b) => a.toLowerCase() === b.toLowerCase();

export function useGitHubStarReward() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [statusState] = useContext(StatusContext);
  const [userState] = useContext(UserContext);
  const status = statusState?.status;
  const enabled = status?.github_star_reward_enabled === true;
  const isLoggedIn = !!userState?.user?.username;
  const [info, setInfo] = useState(null);
  const [claimingRepo, setClaimingRepo] = useState('');

  const refresh = useCallback(async () => {
    if (!enabled || !isLoggedIn) {
      setInfo(null);
      return;
    }
    try {
      const res = await API.get('/api/user/github_star');
      if (res.data.success) setInfo(res.data.data);
    } catch (e) {
      // ignore
    }
  }, [enabled, isLoggedIn]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  // Star 和绑定 GitHub 都在新窗口完成，回到页面时刷新领取状态
  useEffect(() => {
    window.addEventListener('focus', refresh);
    return () => window.removeEventListener('focus', refresh);
  }, [refresh]);

  const claim = async (repo) => {
    setClaimingRepo(repo);
    try {
      const res = await API.post('/api/user/github_star/claim', { repo });
      if (res.data.success) {
        showSuccess(
          t('领取成功，已获得 {{quota}}', {
            quota: renderQuota(res.data.data.quota),
          }),
        );
        await refresh();
      } else {
        showError(res.data.message);
      }
    } finally {
      setClaimingRepo('');
    }
  };

  // 返回仓库的奖励状态；仓库不在活动中时返回 null
  const getRepoReward = (repo) => {
    const repos = status?.github_star_reward_repos || [];
    if (!enabled || !repos.some((r) => sameRepo(r, repo))) return null;

    const reward = {
      quota: status.github_star_reward_quota,
      loading: claimingRepo === repo,
    };
    if (!isLoggedIn) {
      return { ...reward, state: 'login', onAction: () => navigate('/login') };
    }
    if (info?.repos?.some((r) => sameRepo(r.repo, repo) && r.claimed)) {
      return { ...reward, state: 'claimed' };
    }
    if (info && !info.github_bound) {
      return {
        ...reward,
        state: 'bind',
        onAction: () => onGitHubOAuthClicked(status.github_client_id),
      };
    }
    return { ...reward, state: 'claim', onAction: () => claim(repo) };
  };

  return { getRepoReward };
}
