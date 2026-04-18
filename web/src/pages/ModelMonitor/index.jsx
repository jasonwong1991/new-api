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

import React, { useCallback, useEffect, useRef, useState } from 'react';
import {
  Card,
  Table,
  Select,
  Button,
  Tag,
  Typography,
  Space,
  Spin,
} from '@douyinfe/semi-ui';
import { Activity, RefreshCw } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { API, showError } from '../../helpers';

const { Text } = Typography;

const WINDOW_OPTIONS = [
  { value: '1h', label: '最近1小时' },
  { value: '24h', label: '最近24小时' },
  { value: '7d', label: '最近7天' },
];

const formatNumber = (n, digits = 0) => {
  if (n === null || n === undefined || Number.isNaN(n)) return '-';
  return Number(n).toFixed(digits);
};

const renderSuccessRate = (rate, t) => {
  if (rate === null || rate === undefined || Number.isNaN(rate)) {
    return <Tag color='grey'>-</Tag>;
  }
  const pct = (rate * 100).toFixed(2) + '%';
  let color = 'red';
  if (rate >= 0.99) color = 'green';
  else if (rate >= 0.95) color = 'orange';
  return <Tag color={color}>{pct}</Tag>;
};

const ModelMonitorPage = () => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [window, setWindow] = useState('24h');
  const [models, setModels] = useState([]);
  const [generatedAt, setGeneratedAt] = useState(null);
  const [refreshSec, setRefreshSec] = useState(30);
  const timerRef = useRef(null);

  const loadConfig = useCallback(async () => {
    try {
      const res = await API.get('/api/model_monitor/config');
      const { success, data } = res.data || {};
      if (success && data) {
        if (data.default_window) setWindow(data.default_window);
        if (data.refresh_sec) setRefreshSec(data.refresh_sec);
      }
    } catch (e) {
      // silent
    }
  }, []);

  const loadMetrics = useCallback(
    async (w) => {
      const q = w || window;
      setLoading(true);
      try {
        const res = await API.get(`/api/model_monitor/metrics?window=${q}`);
        const { success, message, data } = res.data || {};
        if (success) {
          setModels(data?.models || []);
          setGeneratedAt(data?.generated_at || null);
        } else {
          showError(message || t('加载失败'));
        }
      } catch (e) {
        showError(e.message);
      } finally {
        setLoading(false);
      }
    },
    [window, t],
  );

  useEffect(() => {
    (async () => {
      await loadConfig();
    })();
  }, [loadConfig]);

  useEffect(() => {
    loadMetrics(window);
    if (timerRef.current) clearInterval(timerRef.current);
    if (refreshSec && refreshSec >= 5) {
      timerRef.current = setInterval(() => {
        loadMetrics(window);
      }, refreshSec * 1000);
    }
    return () => {
      if (timerRef.current) clearInterval(timerRef.current);
    };
  }, [window, refreshSec, loadMetrics]);

  const columns = [
    {
      title: t('模型'),
      dataIndex: 'model_name',
      render: (v) => <Text strong>{v || '-'}</Text>,
      sorter: (a, b) => (a.model_name || '').localeCompare(b.model_name || ''),
    },
    {
      title: t('请求数'),
      dataIndex: 'request_count',
      align: 'right',
      sorter: (a, b) => a.request_count - b.request_count,
    },
    {
      title: t('错误数'),
      dataIndex: 'error_count',
      align: 'right',
      sorter: (a, b) => a.error_count - b.error_count,
    },
    {
      title: t('成功率'),
      dataIndex: 'success_rate',
      align: 'center',
      render: (v) => renderSuccessRate(v, t),
      sorter: (a, b) => a.success_rate - b.success_rate,
    },
    {
      title: t('平均延迟(ms)'),
      dataIndex: 'avg_latency_ms',
      align: 'right',
      render: (v) => formatNumber(v, 2),
      sorter: (a, b) => a.avg_latency_ms - b.avg_latency_ms,
    },
    {
      title: 'RPM',
      dataIndex: 'rpm',
      align: 'right',
      render: (v) => formatNumber(v, 3),
      sorter: (a, b) => a.rpm - b.rpm,
    },
    {
      title: 'Prompt Tokens',
      dataIndex: 'prompt_tokens',
      align: 'right',
      sorter: (a, b) => a.prompt_tokens - b.prompt_tokens,
    },
    {
      title: 'Completion Tokens',
      dataIndex: 'completion_tokens',
      align: 'right',
      sorter: (a, b) => a.completion_tokens - b.completion_tokens,
    },
    {
      title: t('总Tokens'),
      dataIndex: 'total_tokens',
      align: 'right',
      sorter: (a, b) => a.total_tokens - b.total_tokens,
    },
    {
      title: t('消耗额度'),
      dataIndex: 'quota',
      align: 'right',
      sorter: (a, b) => a.quota - b.quota,
    },
  ];

  const generatedAtText =
    generatedAt && new Date(generatedAt * 1000).toLocaleString();

  return (
    <div className='p-4'>
      <Card
        className='shadow-sm !rounded-2xl'
        title={
          <div className='flex items-center justify-between w-full gap-2'>
            <div className='flex items-center gap-2'>
              <Activity size={16} />
              <span>{t('模型监控')}</span>
            </div>
            <Space>
              <Select
                value={window}
                onChange={setWindow}
                style={{ width: 160 }}
                optionList={WINDOW_OPTIONS.map((o) => ({
                  value: o.value,
                  label: t(o.label),
                }))}
              />
              <Button
                icon={<RefreshCw size={14} />}
                onClick={() => loadMetrics(window)}
                loading={loading}
                theme='borderless'
                type='tertiary'
              >
                {t('刷新')}
              </Button>
            </Space>
          </div>
        }
      >
        <Spin spinning={loading}>
          <Table
            dataSource={models}
            columns={columns}
            rowKey='model_name'
            pagination={{
              pageSize: 20,
              showSizeChanger: true,
              pageSizeOptions: ['10', '20', '50', '100'],
            }}
            size='middle'
            scroll={{ x: 'max-content' }}
          />
        </Spin>
        {generatedAtText && (
          <div className='mt-2 text-xs text-gray-500'>
            {t('数据生成时间')}: {generatedAtText}
            {refreshSec ? ` · ${t('自动刷新')} ${refreshSec}s` : ''}
          </div>
        )}
      </Card>
    </div>
  );
};

export default ModelMonitorPage;
