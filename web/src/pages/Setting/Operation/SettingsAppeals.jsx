import React, { useEffect, useState } from 'react';
import {
  Card,
  Table,
  Button,
  Tag,
  Space,
  Typography,
  Modal,
  TextArea,
  Empty,
  Avatar,
  Tabs,
  TabPane,
  Form,
} from '@douyinfe/semi-ui';
import { IconUser, IconTick, IconClose } from '@douyinfe/semi-icons';
import { useTranslation } from 'react-i18next';
import { API, showError, showSuccess } from '../../../helpers';

const { Text } = Typography;

const APPEAL_STATUS = {
  PENDING: 1,
  APPROVED: 2,
  REJECTED: 3,
};

const SettingsAppeals = () => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [appeals, setAppeals] = useState([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [statusFilter, setStatusFilter] = useState('0');
  const [reviewModalVisible, setReviewModalVisible] = useState(false);
  const [selectedAppeal, setSelectedAppeal] = useState(null);
  const [reviewAction, setReviewAction] = useState('');
  const [reviewNote, setReviewNote] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const pageSize = 10;

  const fetchAppeals = async () => {
    setLoading(true);
    try {
      const res = await API.get('/api/appeal/', {
        params: {
          status: statusFilter,
          page,
          page_size: pageSize,
        },
      });
      const { success, data, message } = res.data;
      if (success) {
        setAppeals(data.items || []);
        setTotal(data.total || 0);
      } else {
        showError(message);
      }
    } catch (error) {
      showError(t('获取申诉列表失败'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchAppeals();
  }, [page, statusFilter]);

  const handleReview = (appeal, action) => {
    setSelectedAppeal(appeal);
    setReviewAction(action);
    setReviewNote('');
    setReviewModalVisible(true);
  };

  const submitReview = async () => {
    if (!selectedAppeal) return;

    setSubmitting(true);
    try {
      const endpoint =
        reviewAction === 'approve'
          ? `/api/appeal/${selectedAppeal.id}/approve`
          : `/api/appeal/${selectedAppeal.id}/reject`;

      const res = await API.post(endpoint, { note: reviewNote });
      const { success, message } = res.data;
      if (success) {
        showSuccess(message);
        setReviewModalVisible(false);
        fetchAppeals();
      } else {
        showError(message);
      }
    } catch (error) {
      showError(t('操作失败'));
    } finally {
      setSubmitting(false);
    }
  };

  const getStatusTag = (status) => {
    switch (status) {
      case APPEAL_STATUS.PENDING:
        return <Tag color='orange'>{t('待处理')}</Tag>;
      case APPEAL_STATUS.APPROVED:
        return <Tag color='green'>{t('已通过')}</Tag>;
      case APPEAL_STATUS.REJECTED:
        return <Tag color='red'>{t('已拒绝')}</Tag>;
      default:
        return <Tag>{t('未知')}</Tag>;
    }
  };

  const formatTime = (timestamp) => {
    if (!timestamp) return '-';
    return new Date(timestamp * 1000).toLocaleString();
  };

  const columns = [
    {
      title: t('用户'),
      dataIndex: 'linux_do_username',
      key: 'user',
      width: 180,
      render: (text, record) => (
        <Space>
          {record.linux_do_avatar ? (
            <Avatar size='small' src={record.linux_do_avatar} />
          ) : (
            <Avatar size='small' icon={<IconUser />} />
          )}
          <div>
            <Text strong>{text || record.display_name || record.username}</Text>
          </div>
        </Space>
      ),
    },
    {
      title: t('封禁理由'),
      dataIndex: 'ban_reason',
      key: 'ban_reason',
      width: 200,
      render: (text) => (
        <Text type='danger' ellipsis={{ showTooltip: true }} style={{ maxWidth: 180 }}>
          {text || t('未说明')}
        </Text>
      ),
    },
    {
      title: t('申诉理由'),
      dataIndex: 'reason',
      key: 'reason',
      render: (text) => (
        <Text ellipsis={{ showTooltip: true }} style={{ maxWidth: 250 }}>
          {text}
        </Text>
      ),
    },
    {
      title: t('状态'),
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status) => getStatusTag(status),
    },
    {
      title: t('提交时间'),
      dataIndex: 'created_at',
      key: 'created_at',
      width: 170,
      render: (time) => formatTime(time),
    },
    {
      title: t('操作'),
      key: 'action',
      width: 150,
      render: (_, record) => {
        if (record.status !== APPEAL_STATUS.PENDING) {
          return (
            <Text type='tertiary' size='small'>
              {record.admin_note || '-'}
            </Text>
          );
        }
        return (
          <Space>
            <Button
              size='small'
              type='primary'
              theme='solid'
              icon={<IconTick />}
              onClick={() => handleReview(record, 'approve')}
            >
              {t('通过')}
            </Button>
            <Button
              size='small'
              type='danger'
              icon={<IconClose />}
              onClick={() => handleReview(record, 'reject')}
            >
              {t('拒绝')}
            </Button>
          </Space>
        );
      },
    },
  ];

  return (
    <Card>
      <Form.Section text={t('申诉管理')} extraText={t('处理用户的封禁申诉请求')}>
        <Tabs
          activeKey={statusFilter}
          onChange={(key) => {
            setStatusFilter(key);
            setPage(1);
          }}
          type='button'
          style={{ marginBottom: 16 }}
        >
          <TabPane tab={t('全部')} itemKey='0' />
          <TabPane tab={t('待处理')} itemKey='1' />
          <TabPane tab={t('已通过')} itemKey='2' />
          <TabPane tab={t('已拒绝')} itemKey='3' />
        </Tabs>

        {appeals.length > 0 ? (
          <Table
            columns={columns}
            dataSource={appeals}
            rowKey='id'
            loading={loading}
            pagination={{
              currentPage: page,
              pageSize,
              total,
              onPageChange: setPage,
            }}
          />
        ) : (
          <Empty description={t('暂无申诉记录')} />
        )}
      </Form.Section>

      <Modal
        title={reviewAction === 'approve' ? t('通过申诉') : t('拒绝申诉')}
        visible={reviewModalVisible}
        onCancel={() => setReviewModalVisible(false)}
        footer={null}
      >
        {selectedAppeal && (
          <div className='mb-4'>
            <div className='mb-2'>
              <Text type='secondary'>{t('用户：')}</Text>
              <Text strong>
                {selectedAppeal.linux_do_username ||
                  selectedAppeal.display_name ||
                  selectedAppeal.username}
              </Text>
            </div>
            <div className='mb-2'>
              <Text type='secondary'>{t('封禁理由：')}</Text>
              <Text type='danger'>{selectedAppeal.ban_reason || t('未说明')}</Text>
            </div>
            <div className='mb-4'>
              <Text type='secondary'>{t('申诉理由：')}</Text>
              <Text>{selectedAppeal.reason}</Text>
            </div>
          </div>
        )}

        <Form.Slot label={t('处理备注（可选）')}>
          <TextArea
            value={reviewNote}
            onChange={setReviewNote}
            placeholder={t('添加处理备注...')}
            rows={3}
          />
        </Form.Slot>

        <div className='flex justify-end gap-2 mt-4'>
          <Button onClick={() => setReviewModalVisible(false)} disabled={submitting}>
            {t('取消')}
          </Button>
          <Button
            theme='solid'
            type={reviewAction === 'approve' ? 'primary' : 'danger'}
            onClick={submitReview}
            loading={submitting}
          >
            {reviewAction === 'approve' ? t('确认通过') : t('确认拒绝')}
          </Button>
        </div>
      </Modal>
    </Card>
  );
};

export default SettingsAppeals;
