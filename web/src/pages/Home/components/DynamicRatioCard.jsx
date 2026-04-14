import React, { useEffect, useState, useCallback } from 'react';
import { Card, Typography, Tag, Tooltip, Spin } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
import { API } from '../../../helpers';

const DynamicRatioCard = () => {
  const { t } = useTranslation();
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);

  const fetchData = useCallback(async () => {
    try {
      const res = await API.get('/api/dynamic_ratio');
      const { success, data: ratioData } = res.data;
      if (success) {
        setData(ratioData);
      }
    } catch (error) {
      console.error('Failed to fetch dynamic ratio:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 30000);
    return () => clearInterval(interval);
  }, [fetchData]);

  if (loading) {
    return (
      <Card bodyStyle={{ padding: '16px 24px' }}>
        <div className="flex items-center justify-center py-2">
          <Spin size="middle" />
        </div>
      </Card>
    );
  }

  if (!data || !data.enabled) return null;

  const formatTokens = (tokens) => {
    const yi = tokens / 100000000;
    if (yi >= 10000) return (yi / 10000).toFixed(1) + t('万亿');
    if (yi >= 1) return yi.toFixed(1) + t('亿');
    return (tokens / 10000).toFixed(0) + t('万');
  };

  const getRatioColor = (ratio) => {
    if (ratio <= 1.0) return 'green';
    if (ratio <= 2.0) return 'blue';
    if (ratio <= 3.5) return 'orange';
    return 'red';
  };

  const getRatioLevel = (ratio) => {
    if (ratio <= 1.0) return t('空闲');
    if (ratio <= 2.0) return t('正常');
    if (ratio <= 3.5) return t('繁忙');
    return t('高峰');
  };

  const ratioColor = getRatioColor(data.dynamic_ratio);
  const borderColorMap = {
    green: '#00b42a',
    blue: '#0077fa',
    orange: '#ff7d00',
    red: '#f53f3f',
  };

  return (
    <Card
      bodyStyle={{ padding: '16px 24px' }}
      className="border-l-4"
      style={{ borderLeftColor: borderColorMap[ratioColor] }}
    >
      <div className="flex items-center justify-between flex-wrap gap-4">
        <div className="flex items-center gap-6 flex-wrap">
          <Tooltip content={t('过去24小时的总Token用量')}>
            <div className="flex items-center gap-2">
              <Typography.Text type="secondary" size="small">
                {t('24h Tokens')}
              </Typography.Text>
              <Typography.Text strong>
                {formatTokens(data.tokens_24h)}
              </Typography.Text>
            </div>
          </Tooltip>
          <Tooltip content={t('当前每分钟请求数')}>
            <div className="flex items-center gap-2">
              <Typography.Text type="secondary" size="small">
                {t('当前 RPM')}
              </Typography.Text>
              <Typography.Text strong>{data.current_rpm}</Typography.Text>
            </div>
          </Tooltip>
        </div>
        <Tooltip
          content={t(
            '基于24h用量和RPM动态计算的倍率乘数，应用于分组倍率之上'
          )}
        >
          <div className="flex items-center gap-2">
            <Typography.Text type="secondary" size="small">
              {t('动态倍率')}
            </Typography.Text>
            <Tag
              color={ratioColor}
              size="large"
              style={{ borderRadius: 12, fontWeight: 700, fontSize: 16 }}
            >
              x{data.dynamic_ratio.toFixed(1)}
            </Tag>
            <Tag
              color={ratioColor}
              size="small"
              style={{ borderRadius: 12 }}
            >
              {getRatioLevel(data.dynamic_ratio)}
            </Tag>
          </div>
        </Tooltip>
      </div>
    </Card>
  );
};

export default DynamicRatioCard;
