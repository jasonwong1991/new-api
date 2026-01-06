import React, { useEffect, useState, useContext } from 'react';
import {
  Table,
  Typography,
  Tag,
  Button,
  Empty,
  Spin,
  Avatar,
  Space,
} from '@douyinfe/semi-ui';
import { IconUser } from '@douyinfe/semi-icons';
import { useTranslation } from 'react-i18next';
import { API, showError } from '../../../../helpers';
import { UserContext } from '../../../../context/User';
import AppealModal from './AppealModal';

const { Text } = Typography;

const BanListTab = () => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [bannedUsers, setBannedUsers] = useState([]);
  const [userState] = useContext(UserContext);
  const [appealModalVisible, setAppealModalVisible] = useState(false);
  const [selectedUser, setSelectedUser] = useState(null);

  const fetchBanList = async () => {
    setLoading(true);
    try {
      const res = await API.get('/api/ban-list');
      const { success, data, message } = res.data;
      if (success) {
        setBannedUsers(data || []);
      } else {
        showError(message);
      }
    } catch (error) {
      showError(t('获取封禁名单失败'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchBanList();
  }, []);

  const handleAppealClick = (user) => {
    setSelectedUser(user);
    setAppealModalVisible(true);
  };

  const isCurrentUserBanned = (userId) => {
    return userState?.user?.id === userId && userState?.user?.status === 2;
  };

  const columns = [
    {
      title: t('LinuxDO 用户'),
      dataIndex: 'linux_do_username',
      key: 'linux_do_user',
      render: (text, record) => (
        <Space>
          {record.linux_do_avatar ? (
            <Avatar size='small' src={record.linux_do_avatar} />
          ) : (
            <Avatar size='small' icon={<IconUser />} />
          )}
          <Text strong>{text || '-'}</Text>
        </Space>
      ),
    },
    {
      title: t('站内用户名'),
      dataIndex: 'display_name',
      key: 'display_name',
      render: (text) => <Text>{text || '-'}</Text>,
    },
    {
      title: t('封禁理由'),
      dataIndex: 'ban_reason',
      key: 'ban_reason',
      render: (text) => (
        <Text type='danger'>{text || t('未说明')}</Text>
      ),
    },
    {
      title: t('申诉状态'),
      dataIndex: 'has_pending_appeal',
      key: 'appeal_status',
      width: 120,
      render: (hasPending) =>
        hasPending ? (
          <Tag color='orange'>{t('申诉中')}</Tag>
        ) : (
          <Tag color='grey'>{t('未申诉')}</Tag>
        ),
    },
    {
      title: t('操作'),
      key: 'action',
      width: 100,
      render: (_, record) => {
        const canAppeal =
          isCurrentUserBanned(record.id) && !record.has_pending_appeal;
        return canAppeal ? (
          <Button
            size='small'
            theme='solid'
            type='warning'
            onClick={() => handleAppealClick(record)}
          >
            {t('申诉')}
          </Button>
        ) : null;
      },
    },
  ];

  return (
    <div className='p-4'>
      <Spin spinning={loading}>
        {bannedUsers.length > 0 ? (
          <Table
            columns={columns}
            dataSource={bannedUsers}
            rowKey='id'
            pagination={false}
          />
        ) : (
          <Empty
            image={<IconUser style={{ fontSize: 48, color: 'var(--semi-color-text-2)' }} />}
            description={t('暂无封禁用户')}
          />
        )}
      </Spin>

      <AppealModal
        visible={appealModalVisible}
        onClose={() => {
          setAppealModalVisible(false);
          setSelectedUser(null);
        }}
        onSuccess={() => {
          setAppealModalVisible(false);
          setSelectedUser(null);
          fetchBanList();
        }}
        user={selectedUser}
      />
    </div>
  );
};

export default BanListTab;
