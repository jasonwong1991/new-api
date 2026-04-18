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

import React, { useMemo } from 'react';
import { Typography, Tag } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
import StatusBars from './StatusBars';
import { useMonitorResult } from './MonitorDataContext';

const { Text } = Typography;

const renderSuccessRateTag = (rate) => {
  if (rate === null || rate === undefined || Number.isNaN(rate)) {
    return <Tag color='grey'>-</Tag>;
  }
  const pct = (rate * 100).toFixed(2) + '%';
  let color = 'red';
  if (rate >= 0.99) color = 'green';
  else if (rate >= 0.95) color = 'orange';
  return <Tag color={color}>{pct}</Tag>;
};

const formatNumber = (n, digits = 0) => {
  if (n === null || n === undefined || Number.isNaN(n)) return '-';
  return Number(n).toFixed(digits);
};

const ModelStatusRow = ({ modelName, granularity }) => {
  const { t } = useTranslation();
  const result = useMonitorResult(modelName);

  const buckets = result?.buckets || [];
  const summary = result?.summary || null;
  const generatedAt = result?.generated_at;

  const successRate = summary?.success_rate;

  const generatedAtText = useMemo(() => {
    if (!generatedAt) return '';
    const d = new Date(generatedAt * 1000);
    return d.toLocaleTimeString();
  }, [generatedAt]);

  return (
    <div className='rounded-xl bg-[var(--semi-color-fill-0)] border border-[var(--semi-color-border)] px-4 py-3 mb-3'>
      <div className='flex items-center justify-between gap-3 mb-2 flex-wrap'>
        <div className='flex items-center gap-2 min-w-0'>
          <Text strong className='truncate' style={{ maxWidth: 320 }}>
            {modelName}
          </Text>
          {renderSuccessRateTag(successRate)}
        </div>
        <div className='flex items-center gap-4 text-xs text-[var(--semi-color-text-2)] flex-wrap'>
          <span>
            {t('请求')}: {formatNumber(summary?.request_count)}
          </span>
          <span>
            {t('错误')}: {formatNumber(summary?.error_count)}
          </span>
          <span>
            {t('延迟')}:{' '}
            {summary?.avg_latency_ms
              ? formatNumber(summary.avg_latency_ms, 0) + ' ms'
              : '-'}
          </span>
          <span>RPM: {formatNumber(summary?.rpm, 2)}</span>
          {generatedAtText && <span>{generatedAtText}</span>}
        </div>
      </div>
      <StatusBars buckets={buckets} granularity={granularity} />
    </div>
  );
};

// 共享 Context 已由 Provider 保证只在数据真正变化时通知，这里 memo 防误刷
export default React.memo(ModelStatusRow, (prev, next) => {
  return (
    prev.modelName === next.modelName && prev.granularity === next.granularity
  );
});
