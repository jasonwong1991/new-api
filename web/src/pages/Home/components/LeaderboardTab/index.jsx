import React, { useEffect, useState } from 'react';
import { Typography, Avatar, Tag, Card, Tabs, TabPane } from '@douyinfe/semi-ui';
import { IconUser, IconBox } from '@douyinfe/semi-icons';
import { useTranslation } from 'react-i18next';
import CardTable from '../../../../components/common/ui/CardTable';
import { API, showError } from '../../../../helpers';
import { renderNumber } from '../../../../helpers/render';

const LeaderboardTab = () => {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState('users');
  const [userData, setUserData] = useState([]);
  const [modelData, setModelData] = useState([]);
  const [myRank, setMyRank] = useState(null);
  const [userLoading, setUserLoading] = useState(false);
  const [modelLoading, setModelLoading] = useState(false);

  useEffect(() => {
    fetchUserData();
  }, []);

  useEffect(() => {
    if (activeTab === 'models' && modelData.length === 0) {
      fetchModelData();
    }
  }, [activeTab]);

  const fetchUserData = async () => {
    setUserLoading(true);
    try {
      const res = await API.get('/api/leaderboard');
      const { success, message, data: resData } = res.data;
      if (success) {
        setUserData(resData?.leaderboard || []);
        setMyRank(resData?.my_rank || null);
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message || t('获取排行榜数据失败'));
    } finally {
      setUserLoading(false);
    }
  };

  const fetchModelData = async () => {
    setModelLoading(true);
    try {
      const res = await API.get('/api/leaderboard/models');
      const { success, message, data: resData } = res.data;
      if (success) {
        setModelData(resData || []);
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message || t('获取模型排行榜数据失败'));
    } finally {
      setModelLoading(false);
    }
  };

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

  const renderUserCell = (record) => {
    const hasLinuxDO = record.linux_do_username && record.linux_do_avatar;
    const avatarSrc = hasLinuxDO ? record.linux_do_avatar : null;

    // Format: name（username） for Linux.do users, otherwise just display_name
    let displayText;
    if (hasLinuxDO) {
      const name = record.display_name || record.linux_do_username;
      displayText = `${name}（${record.linux_do_username}）`;
    } else {
      displayText = record.display_name || t('匿名用户');
    }

    const avatarName = hasLinuxDO ? record.linux_do_username : (record.display_name || t('匿名用户'));

    return (
      <div className='flex items-center gap-2'>
        <div className='relative'>
          {avatarSrc ? (
            <Avatar size='small' src={avatarSrc} />
          ) : (
            <Avatar size='small' color={stringToColor(avatarName)}>
              {avatarName?.charAt(0)?.toUpperCase() || '?'}
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
        <Typography.Text>{displayText}</Typography.Text>
      </div>
    );
  };

  const renderRankCell = (text) => {
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
  };

  const userColumns = [
    {
      title: t('排名'),
      dataIndex: 'rank',
      key: 'rank',
      width: 80,
      render: renderRankCell,
    },
    {
      title: t('用户'),
      dataIndex: 'linux_do_username',
      key: 'user',
      render: (_, record) => renderUserCell(record),
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

  const modelColumns = [
    {
      title: t('排名'),
      dataIndex: 'rank',
      key: 'rank',
      width: 80,
      render: renderRankCell,
    },
    {
      title: t('模型'),
      dataIndex: 'model_name',
      key: 'model_name',
      render: (text) => (
        <div className='flex items-center gap-2'>
          <IconBox style={{ color: stringToColor(text) }} />
          <Typography.Text copyable={{ content: text }}>{text}</Typography.Text>
        </div>
      ),
    },
    {
      title: t('调用次数'),
      dataIndex: 'request_count',
      key: 'request_count',
      render: (text) => renderNumber(text || 0),
    },
    {
      title: t('Token 用量'),
      dataIndex: 'total_tokens',
      key: 'total_tokens',
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
      <Tabs activeKey={activeTab} onChange={setActiveTab}>
        <TabPane
          tab={
            <span>
              <IconUser style={{ marginRight: 4 }} />
              {t('用户消耗排行')}
            </span>
          }
          itemKey='users'
        >
          {myRank && (
            <Card
              className='mb-4 mt-4'
              bodyStyle={{ padding: '12px 16px' }}
              style={{
                background:
                  'linear-gradient(135deg, var(--semi-color-primary-light-default) 0%, var(--semi-color-primary-light-hover) 100%)',
                border: '1px solid var(--semi-color-primary)',
              }}
            >
              <div className='flex items-center justify-between flex-wrap gap-3'>
                <div className='flex items-center gap-3'>
                  <IconUser
                    size='large'
                    style={{ color: 'var(--semi-color-primary)' }}
                  />
                  <div>
                    <Typography.Text strong style={{ fontSize: '14px' }}>
                      {t('我的排名')}
                    </Typography.Text>
                    <div className='flex items-center gap-2 mt-1'>
                      {renderUserCell(myRank)}
                    </div>
                  </div>
                </div>
                <div className='flex items-center gap-6 flex-wrap'>
                  <div className='text-center'>
                    <Typography.Text
                      strong
                      style={{
                        fontSize: '24px',
                        color: 'var(--semi-color-primary)',
                      }}
                    >
                      #{myRank.rank}
                    </Typography.Text>
                    <Typography.Text
                      type='tertiary'
                      size='small'
                      style={{ display: 'block' }}
                    >
                      {t('排名')}
                    </Typography.Text>
                  </div>
                  <div className='text-center'>
                    <Typography.Text strong style={{ fontSize: '16px' }}>
                      {renderNumber(myRank.request_count || 0)}
                    </Typography.Text>
                    <Typography.Text
                      type='tertiary'
                      size='small'
                      style={{ display: 'block' }}
                    >
                      {t('请求次数')}
                    </Typography.Text>
                  </div>
                  <div className='text-center'>
                    <Typography.Text strong style={{ fontSize: '16px' }}>
                      {renderNumber(myRank.used_quota || 0)}
                    </Typography.Text>
                    <Typography.Text
                      type='tertiary'
                      size='small'
                      style={{ display: 'block' }}
                    >
                      {t('Token 用量')}
                    </Typography.Text>
                  </div>
                  <div className='text-center'>
                    <Typography.Text strong style={{ fontSize: '16px' }}>
                      ${(myRank.amount_usd || 0).toFixed(2)}
                    </Typography.Text>
                    <Typography.Text
                      type='tertiary'
                      size='small'
                      style={{ display: 'block' }}
                    >
                      {t('消费金额')}
                    </Typography.Text>
                  </div>
                </div>
              </div>
            </Card>
          )}
          <CardTable
            loading={userLoading}
            columns={userColumns}
            dataSource={userData}
            rowKey='rank'
            hidePagination={true}
          />
        </TabPane>
        <TabPane
          tab={
            <span>
              <IconBox style={{ marginRight: 4 }} />
              {t('模型调用排行')}
            </span>
          }
          itemKey='models'
        >
          <div className='mt-4'>
            <CardTable
              loading={modelLoading}
              columns={modelColumns}
              dataSource={modelData}
              rowKey='rank'
              hidePagination={true}
            />
          </div>
        </TabPane>
      </Tabs>
    </div>
  );
};

export default LeaderboardTab;
