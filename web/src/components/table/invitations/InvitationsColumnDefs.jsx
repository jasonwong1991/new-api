import React from 'react';
import { Tag, Button, Space, Popover, Dropdown } from '@douyinfe/semi-ui';
import { IconMore } from '@douyinfe/semi-icons';
import { timestamp2string } from '../../../helpers';

export const INVITATION_STATUS = {
  ENABLED: 1,
  DISABLED: 2,
};

export const INVITATION_STATUS_MAP = {
  [INVITATION_STATUS.ENABLED]: { text: '已启用', color: 'green' },
  [INVITATION_STATUS.DISABLED]: { text: '已禁用', color: 'red' },
};

export const isExpired = (record) => {
  return (
    record.expired_time !== -1 &&
    record.expired_time < Math.floor(Date.now() / 1000)
  );
};

export const isUsedUp = (record) => {
  return record.max_uses !== -1 && record.used_count >= record.max_uses;
};

const renderTimestamp = (timestamp) => {
  if (timestamp === -1) return null;
  return <>{timestamp2string(timestamp)}</>;
};

const renderStatus = (status, record, t) => {
  if (isExpired(record)) {
    return (
      <Tag color='orange' shape='circle'>
        {t('已过期')}
      </Tag>
    );
  }

  if (isUsedUp(record)) {
    return (
      <Tag color='grey' shape='circle'>
        {t('已用完')}
      </Tag>
    );
  }

  const statusConfig = INVITATION_STATUS_MAP[status];
  if (statusConfig) {
    return (
      <Tag color={statusConfig.color} shape='circle'>
        {t(statusConfig.text)}
      </Tag>
    );
  }

  return (
    <Tag color='black' shape='circle'>
      {t('未知状态')}
    </Tag>
  );
};

export const getInvitationsColumns = ({
  t,
  manageInvitation,
  copyText,
  setEditingInvitation,
  setShowEdit,
  showDeleteInvitationModal,
}) => {
  return [
    {
      title: t('ID'),
      dataIndex: 'id',
      width: 60,
    },
    {
      title: t('名称'),
      dataIndex: 'name',
      width: 120,
      render: (text) => text || '-',
    },
    {
      title: t('邀请码'),
      dataIndex: 'code',
      width: 180,
      render: (text) => (
        <code style={{ fontSize: '12px', background: 'var(--semi-color-fill-0)', padding: '2px 6px', borderRadius: '4px' }}>
          {text}
        </code>
      ),
    },
    {
      title: t('状态'),
      dataIndex: 'status',
      width: 90,
      render: (text, record) => renderStatus(text, record, t),
    },
    {
      title: t('使用次数'),
      dataIndex: 'used_count',
      width: 100,
      render: (text, record) => {
        const usedCount = record.used_count || 0;
        const maxUses = record.max_uses;
        if (maxUses === -1) {
          return (
            <Tag color='blue' shape='circle'>
              {usedCount} / {t('无限')}
            </Tag>
          );
        }
        return (
          <Tag color={usedCount >= maxUses ? 'red' : 'green'} shape='circle'>
            {usedCount} / {maxUses}
          </Tag>
        );
      },
    },
    {
      title: t('创建时间'),
      dataIndex: 'created_time',
      width: 160,
      render: (text) => renderTimestamp(text),
    },
    {
      title: t('过期时间'),
      dataIndex: 'expired_time',
      width: 160,
      render: (text) => (text === -1 ? t('永不过期') : renderTimestamp(text)),
    },
    {
      title: '',
      dataIndex: 'operate',
      fixed: 'right',
      width: 180,
      render: (text, record) => {
        const moreMenuItems = [
          {
            node: 'item',
            name: t('删除'),
            type: 'danger',
            onClick: () => showDeleteInvitationModal(record),
          },
        ];

        const canToggle = !isExpired(record) && !isUsedUp(record);
        if (canToggle) {
          if (record.status === INVITATION_STATUS.ENABLED) {
            moreMenuItems.push({
              node: 'item',
              name: t('禁用'),
              type: 'warning',
              onClick: () => manageInvitation(record.id, 'disable', record),
            });
          } else {
            moreMenuItems.push({
              node: 'item',
              name: t('启用'),
              type: 'secondary',
              onClick: () => manageInvitation(record.id, 'enable', record),
            });
          }
        }

        return (
          <Space>
            <Popover content={record.code} style={{ padding: 20 }} position='top'>
              <Button type='tertiary' size='small'>
                {t('查看')}
              </Button>
            </Popover>
            <Button size='small' onClick={() => copyText(record.code)}>
              {t('复制')}
            </Button>
            <Button
              type='tertiary'
              size='small'
              onClick={() => {
                setEditingInvitation(record);
                setShowEdit(true);
              }}
            >
              {t('编辑')}
            </Button>
            <Dropdown trigger='click' position='bottomRight' menu={moreMenuItems}>
              <Button type='tertiary' size='small' icon={<IconMore />} />
            </Dropdown>
          </Space>
        );
      },
    },
  ];
};
