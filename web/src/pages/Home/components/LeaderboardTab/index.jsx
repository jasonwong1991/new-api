import React, { useEffect, useState } from 'react';
import { Typography, Avatar, Tag } from '@douyinfe/semi-ui';
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

  const getLevelColor = (level) => {
    const colors = {
      0: 'grey',
      1: 'green',
      2: 'blue',
      3: 'purple',
      4: 'orange',
    };
    return colors[level] || 'grey';
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
      dataIndex: 'linux_do_username',
      key: 'user',
      render: (_, record) => {
        const hasLinuxDO = record.linux_do_username && record.linux_do_avatar;
        const displayName = hasLinuxDO
          ? record.linux_do_username
          : record.display_name || t('匿名用户');
        const avatarSrc = hasLinuxDO ? record.linux_do_avatar : null;

        return (
          <div className='flex items-center gap-2'>
            <div className='relative'>
              {avatarSrc ? (
                <Avatar size='small' src={avatarSrc} />
              ) : (
                <Avatar size='small' color={stringToColor(displayName)}>
                  {displayName?.charAt(0)?.toUpperCase() || '?'}
                </Avatar>
              )}
              {hasLinuxDO && record.linux_do_level > 0 && (
                <Tag
                  color={getLevelColor(record.linux_do_level)}
                  size='small'
                  style={{
                    position: 'absolute',
                    bottom: -4,
                    right: -8,
                    fontSize: '10px',
                    padding: '0 4px',
                    minWidth: 'auto',
                    height: '14px',
                    lineHeight: '14px',
                  }}
                >
                  {record.linux_do_level}
                </Tag>
              )}
            </div>
            <Typography.Text>{displayName}</Typography.Text>
          </div>
        );
      },
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
