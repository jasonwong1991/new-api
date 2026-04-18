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

import React, { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Card,
  Typography,
  Space,
  Tabs,
  TabPane,
  Input,
  Empty,
  Spin,
  Button,
} from '@douyinfe/semi-ui';
import { Activity, Search, RefreshCw } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { API, showError } from '../../helpers';
import ModelStatusRow from './ModelStatusRow';
import {
  MonitorDataProvider,
  useMonitorStatus,
} from './MonitorDataContext';

const { Text } = Typography;

const GRANULARITY_OPTIONS = [
  { key: 'minute', label: '每分钟', desc: '最近 60 分钟' },
  { key: 'hour', label: '每小时', desc: '最近 24 小时' },
  { key: 'day', label: '每天', desc: '最近 30 天' },
];

const LEGEND_ITEMS = [
  { color: 'bg-emerald-500', label: '正常' },
  { color: 'bg-amber-400', label: '降级' },
  { color: 'bg-rose-500', label: '故障' },
  { color: 'bg-zinc-700/40', label: '无数据' },
];

const MonitorStatusBar = () => {
  const { t } = useTranslation();
  const { loading, generatedAt, reload } = useMonitorStatus();
  const generatedAtText = generatedAt
    ? new Date(generatedAt * 1000).toLocaleTimeString()
    : '';
  return (
    <div className='flex items-center gap-3 text-xs text-[var(--semi-color-text-2)]'>
      {loading && <Spin size='small' />}
      {generatedAtText && (
        <span>
          {t('数据时间')}: {generatedAtText}
        </span>
      )}
      <Button
        icon={<RefreshCw size={12} />}
        size='small'
        theme='borderless'
        type='tertiary'
        onClick={reload}
      >
        {t('刷新')}
      </Button>
    </div>
  );
};

const MonitorList = ({ models, granularity }) => {
  const { t } = useTranslation();
  if (!models || models.length === 0) {
    return <Empty description={t('暂无模型数据')} className='py-10' />;
  }
  return (
    <div>
      {models.map((m) => (
        <ModelStatusRow
          key={`${m}-${granularity}`}
          modelName={m}
          granularity={granularity}
        />
      ))}
    </div>
  );
};

const ModelMonitorPage = () => {
  const { t } = useTranslation();
  const [models, setModels] = useState([]);
  const [granularity, setGranularity] = useState('hour');
  const [refreshSec, setRefreshSec] = useState(30);
  const [loadingList, setLoadingList] = useState(false);
  const [keyword, setKeyword] = useState('');

  const loadModels = useCallback(async () => {
    setLoadingList(true);
    try {
      const res = await API.get('/api/model_monitor/models');
      const { success, message, data } = res.data || {};
      if (success) {
        setModels(data?.models || []);
        if (data?.refresh_sec) setRefreshSec(data.refresh_sec);
      } else {
        showError(message || t('加载失败'));
      }
    } catch (e) {
      showError(e.message);
    } finally {
      setLoadingList(false);
    }
  }, [t]);

  useEffect(() => {
    loadModels();
  }, [loadModels]);

  const filteredModels = useMemo(() => {
    const kw = keyword.trim().toLowerCase();
    if (!kw) return models;
    return models.filter((m) => (m || '').toLowerCase().includes(kw));
  }, [models, keyword]);

  return (
    <div className='p-4'>
      <Card
        className='shadow-sm !rounded-2xl'
        title={
          <div className='flex flex-col md:flex-row md:items-center md:justify-between w-full gap-3'>
            <div className='flex items-center gap-2'>
              <Activity size={16} />
              <span>{t('模型监控')}</span>
              <Text type='tertiary' className='ml-2 text-xs'>
                {t('共享数据源 · 服务端缓存')}
              </Text>
            </div>
            <Space>
              <Input
                prefix={<Search size={14} />}
                placeholder={t('搜索模型')}
                value={keyword}
                onChange={setKeyword}
                showClear
                style={{ width: 220 }}
              />
            </Space>
          </div>
        }
      >
        <MonitorDataProvider
          models={filteredModels}
          granularity={granularity}
          refreshSec={refreshSec}
        >
          <div className='flex items-center justify-between flex-wrap gap-2 mb-2'>
            <Tabs
              type='line'
              activeKey={granularity}
              onChange={setGranularity}
              className='flex-1'
            >
              {GRANULARITY_OPTIONS.map((o) => (
                <TabPane
                  key={o.key}
                  itemKey={o.key}
                  tab={
                    <span>
                      {t(o.label)}
                      <Text type='tertiary' className='ml-2 text-xs'>
                        · {t(o.desc)}
                      </Text>
                    </span>
                  }
                />
              ))}
            </Tabs>
            <MonitorStatusBar />
          </div>

          <div className='flex items-center gap-4 mb-3 text-xs text-[var(--semi-color-text-2)]'>
            {LEGEND_ITEMS.map((it) => (
              <div key={it.label} className='flex items-center gap-1'>
                <span
                  className={`inline-block w-3 h-3 rounded-[2px] ${it.color}`}
                />
                <span>{t(it.label)}</span>
              </div>
            ))}
          </div>

          {loadingList ? (
            <div className='flex justify-center py-10'>
              <Spin />
            </div>
          ) : (
            <MonitorList models={filteredModels} granularity={granularity} />
          )}
        </MonitorDataProvider>
      </Card>
    </div>
  );
};

export default ModelMonitorPage;
