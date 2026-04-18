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

import React, { useEffect, useMemo, useRef, useState } from 'react';
import { Tooltip } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';

const STATUS_CLASS = {
  up: 'bg-emerald-500',
  degraded: 'bg-amber-400',
  down: 'bg-rose-500',
  no_data: 'bg-zinc-700/40',
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
        const cls = STATUS_CLASS[b.status] || STATUS_CLASS.no_data;
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
              className={`flex-1 min-w-[4px] rounded-[2px] transition-all duration-150 cursor-pointer ${cls} ${
                isHover ? 'opacity-80 scale-y-110' : ''
              }`}
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
