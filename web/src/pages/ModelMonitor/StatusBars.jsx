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

import React, { useState } from 'react';
import { Tooltip } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';

// 注意：项目 tailwind.config.js 覆盖了默认调色板（只保留 semi-* 变量），
// 所以这里不能用 bg-emerald-500 之类的原生 Tailwind 类，必须内联颜色。
export const STATUS_COLORS = {
  up: '#10b981', // 绿
  degraded: '#f59e0b', // 橙
  down: '#ef4444', // 红
  no_data: 'rgba(113, 113, 122, 0.25)', // 灰
};

const formatBucketTime = (ts, granularity) => {
  const d = new Date(ts * 1000);
  const pad = (n) => String(n).padStart(2, '0');
  if (granularity === 'day') {
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
  }
  if (granularity === 'hour') {
    return `${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:00`;
  }
  return `${pad(d.getHours())}:${pad(d.getMinutes())}`;
};

const StatusBars = ({ buckets, granularity }) => {
  const { t } = useTranslation();
  const [hoveredIdx, setHoveredIdx] = useState(null);
  const items = buckets || [];

  return (
    <div className='flex items-stretch gap-[2px] w-full h-10 select-none'>
      {items.map((b, i) => {
        const color = STATUS_COLORS[b.status] || STATUS_COLORS.no_data;
        const isHover = hoveredIdx === i;
        const tipLines = [
          formatBucketTime(b.ts, granularity),
          `${t('请求')}: ${b.request_count}`,
          `${t('错误')}: ${b.error_count}`,
        ];
        if (b.avg_latency_ms > 0) {
          tipLines.push(`${t('延迟')}: ${b.avg_latency_ms.toFixed(0)} ms`);
        }
        return (
          <Tooltip
            key={`${b.ts}-${i}`}
            content={
              <div className='text-xs'>
                {tipLines.map((line, idx) => (
                  <div key={idx}>{line}</div>
                ))}
              </div>
            }
            position='top'
          >
            <div
              className='flex-1 min-w-[4px] cursor-pointer transition-all duration-150'
              style={{
                backgroundColor: color,
                borderRadius: 2,
                opacity: isHover ? 0.85 : 1,
                transform: isHover ? 'scaleY(1.08)' : 'none',
              }}
              onMouseEnter={() => setHoveredIdx(i)}
              onMouseLeave={() => setHoveredIdx(null)}
            />
          </Tooltip>
        );
      })}
    </div>
  );
};

export default React.memo(StatusBars);
