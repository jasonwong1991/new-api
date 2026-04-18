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

import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react';
import { API, showError } from '../../helpers';

const MonitorDataContext = createContext({
  resultsByModel: {},
  generatedAt: null,
  loading: false,
  error: null,
});

const granularityDefaultInterval = (granularity, refreshSec) => {
  const base = refreshSec && refreshSec > 0 ? refreshSec : 60;
  if (granularity === 'minute') return Math.min(base, 30);
  if (granularity === 'day') return Math.max(base, 300);
  return Math.min(base, 60);
};

/**
 * 统一数据源：整个页面共享一个 timer，一次请求带所有模型，
 * 后端有缓存 + 单飞，保证 DB 压力可控。
 */
export const MonitorDataProvider = ({
  models,
  granularity,
  refreshSec,
  children,
}) => {
  const [resultsByModel, setResultsByModel] = useState({});
  const [generatedAt, setGeneratedAt] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const abortRef = useRef(null);
  const mountedRef = useRef(true);

  const modelsKey = useMemo(() => (models || []).join('\u0001'), [models]);

  const load = useCallback(async () => {
    if (!models || models.length === 0) {
      setResultsByModel({});
      return;
    }
    if (abortRef.current) {
      try {
        abortRef.current.abort?.();
      } catch (_) {
        // noop
      }
    }
    const controller =
      typeof AbortController !== 'undefined' ? new AbortController() : null;
    abortRef.current = controller;
    setLoading(true);
    try {
      // 分批：每 50 个模型一个请求，避免 URL 过长
      const batches = [];
      for (let i = 0; i < models.length; i += 50) {
        batches.push(models.slice(i, i + 50));
      }
      const merged = {};
      let latestTs = null;
      for (const batch of batches) {
        const url = `/api/model_monitor/bars?granularity=${granularity}&models=${encodeURIComponent(
          batch.join(','),
        )}`;
        const res = await API.get(url, controller ? { signal: controller.signal } : undefined);
        const { success, data, message } = res.data || {};
        if (!success) {
          throw new Error(message || '加载失败');
        }
        Object.assign(merged, data?.results || {});
        if (data?.generated_at) latestTs = data.generated_at;
      }
      if (!mountedRef.current) return;
      setResultsByModel(merged);
      setGeneratedAt(latestTs);
      setError(null);
    } catch (e) {
      if (e?.name === 'CanceledError' || e?.code === 'ERR_CANCELED') return;
      if (!mountedRef.current) return;
      setError(e?.message || String(e));
      // 首次加载失败也提示，但持续失败不再刷屏
      if (!generatedAt) showError(e?.message || '加载失败');
    } finally {
      if (mountedRef.current) setLoading(false);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [modelsKey, granularity]);

  useEffect(() => {
    mountedRef.current = true;
    load();
    const interval = granularityDefaultInterval(granularity, refreshSec);
    const timer = setInterval(() => {
      if (document.visibilityState === 'visible') load();
    }, interval * 1000);
    // 页面从后台切回前台时立即刷新
    const onVis = () => {
      if (document.visibilityState === 'visible') load();
    };
    document.addEventListener('visibilitychange', onVis);
    return () => {
      mountedRef.current = false;
      clearInterval(timer);
      document.removeEventListener('visibilitychange', onVis);
      if (abortRef.current) {
        try {
          abortRef.current.abort?.();
        } catch (_) {
          // noop
        }
      }
    };
  }, [load, granularity, refreshSec]);

  const value = useMemo(
    () => ({ resultsByModel, generatedAt, loading, error, reload: load }),
    [resultsByModel, generatedAt, loading, error, load],
  );

  return (
    <MonitorDataContext.Provider value={value}>
      {children}
    </MonitorDataContext.Provider>
  );
};

export const useMonitorResult = (modelName) => {
  const ctx = useContext(MonitorDataContext);
  return ctx.resultsByModel?.[modelName] || null;
};

export const useMonitorStatus = () => {
  const ctx = useContext(MonitorDataContext);
  return {
    loading: ctx.loading,
    error: ctx.error,
    generatedAt: ctx.generatedAt,
    reload: ctx.reload,
  };
};
