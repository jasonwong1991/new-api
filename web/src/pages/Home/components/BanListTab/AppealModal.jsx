import React, { useState } from 'react';
import { Modal, Form, TextArea, Button, Typography } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
import { API, showError, showSuccess } from '../../../../helpers';

const { Text } = Typography;

const AppealModal = ({ visible, onClose, onSuccess, user }) => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [reason, setReason] = useState('');

  const handleSubmit = async () => {
    if (!reason || reason.trim().length < 10) {
      showError(t('申诉理由至少需要10个字符'));
      return;
    }
    if (reason.length > 1000) {
      showError(t('申诉理由不能超过1000个字符'));
      return;
    }

    setLoading(true);
    try {
      const res = await API.post('/api/user/appeal', { reason: reason.trim() });
      const { success, message } = res.data;
      if (success) {
        showSuccess(t('申诉已提交，请等待管理员审核'));
        setReason('');
        onSuccess?.();
      } else {
        showError(message);
      }
    } catch (error) {
      showError(t('提交申诉失败'));
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    setReason('');
    onClose?.();
  };

  return (
    <Modal
      title={t('提交申诉')}
      visible={visible}
      onCancel={handleClose}
      footer={null}
      closeOnEsc={!loading}
      maskClosable={!loading}
    >
      <div className='mb-4'>
        <Text type='secondary'>
          {t('您的账户因以下原因被封禁：')}
        </Text>
        <Text type='danger' strong style={{ display: 'block', marginTop: 8 }}>
          {user?.ban_reason || t('未说明')}
        </Text>
      </div>

      <Form>
        <Form.Slot label={t('申诉理由')}>
          <TextArea
            value={reason}
            onChange={setReason}
            placeholder={t('请详细说明您的申诉理由（10-1000字符）')}
            rows={5}
            maxCount={1000}
            showClear
          />
        </Form.Slot>
      </Form>

      <div className='flex justify-end gap-2 mt-4'>
        <Button onClick={handleClose} disabled={loading}>
          {t('取消')}
        </Button>
        <Button
          theme='solid'
          type='warning'
          onClick={handleSubmit}
          loading={loading}
        >
          {t('提交申诉')}
        </Button>
      </div>
    </Modal>
  );
};

export default AppealModal;
