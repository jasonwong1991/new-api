import React from 'react';
import { Modal, Typography } from '@douyinfe/semi-ui';

const { Text } = Typography;

const DeleteInvitationModal = ({
  visible,
  onCancel,
  record,
  manageInvitation,
  refresh,
  t,
}) => {
  const handleDelete = async () => {
    await manageInvitation(record.id, 'delete', record);
    onCancel();
    await refresh();
  };

  return (
    <Modal
      title={t('删除邀请码')}
      visible={visible}
      onOk={handleDelete}
      onCancel={onCancel}
      okType='danger'
      okText={t('删除')}
      cancelText={t('取消')}
    >
      <Text>
        {t('确定要删除邀请码')} <Text strong>{record?.code}</Text> {t('吗？此操作不可撤销。')}
      </Text>
    </Modal>
  );
};

export default DeleteInvitationModal;
