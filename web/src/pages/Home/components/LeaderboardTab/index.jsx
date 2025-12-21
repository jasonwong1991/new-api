import React, { useEffect, useState } from 'react';
import { Typography, Avatar } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
import CardTable from '../../../../components/common/ui/CardTable';
import { API, showError } from '../../../../helpers';
import { renderNumber } from '../../../../helpers/render';

const LeaderboardTab = () => {
  const { t } = useTranslation();
  const [data, setData] = useState([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const fetchData = async () => {
      setLoading(true);
      try {
        const res = await API.get('/api/leaderboard');
        const { success, message, data: resData } = res.data;
        if (success) {
          setData(resData || []);
        } else {
          showError(message);
        }
      } catch (error) {
        showError(error.message || t('获取排行榜数据失败'));
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, []);

  const stringToColor = (str) => {
    if (!str) return '#999';
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
      hash = str.charCodeAt(i) + ((hash << 5) - hash);
    }
    const c = (hash & 0x00ffffff).toString(16).toUpperCase();
    return '#' + '00000'.substring(0, 6 - c.length) + c;
  };

  const columns = [
    {
      title: t('排名'),
      dataIndex: 'rank',
      key: 'rank',
      width: 80,
      render: (text) => {
        let color = 'var(--semi-color-text-2)';
        let emoji = '';
        if (text === 1) {
          color = '#FFD700';
          emoji = '🥇';
        }
        if (text === 2) {
          color = '#C0C0C0';
          emoji = '🥈';
        }
        if (text === 3) {
          color = '#CD7F32';
          emoji = '🥉';
        }
        return (
          <Typography.Text
            style={{
              color,
              fontWeight: text <= 3 ? 'bold' : 'normal',
              fontSize: text <= 3 ? '1.1em' : '1em',
            }}
          >
            {emoji} #{text}
          </Typography.Text>
        );
      },
    },
    {
      title: t('用户'),
      dataIndex: 'display_name',
      key: 'display_name',
      render: (text) => (
        <div className='flex items-center gap-2'>
          <Avatar size='small' color={stringToColor(text)}>
            {text?.charAt(0)?.toUpperCase() || '?'}
          </Avatar>
          <Typography.Text>{text || t('匿名用户')}</Typography.Text>
        </div>
      ),
    },
    {
      title: t('请求次数'),
      dataIndex: 'request_count',
      key: 'request_count',
      render: (text) => renderNumber(text || 0),
    },
    {
      title: t('Token 用量'),
      dataIndex: 'used_quota',
      key: 'used_quota',
      render: (text) => renderNumber(text || 0),
    },
    {
      title: t('消费金额'),
      dataIndex: 'amount_usd',
      key: 'amount_usd',
      render: (text) => `$${(text || 0).toFixed(2)}`,
    },
  ];

  return (
    <div className='p-4'>
      <CardTable
        loading={loading}
        columns={columns}
        dataSource={data}
        rowKey='rank'
        hidePagination={true}
      />
    </div>
  );
};

export default LeaderboardTab;
