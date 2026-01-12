import React, { useState, useEffect } from 'react';
import { Input, Button, Modal, Typography, Toast } from '@douyinfe/semi-ui';
import { IconSearch } from '@douyinfe/semi-icons';
import { API } from '../../../helpers';
import { useTranslation } from 'react-i18next';
import { renderQuota } from '../../../helpers/render';
import { timestamp2string } from '../../../helpers/utils';

const { Text } = Typography;

const ArchivedUserCheck = () => {
  const { t } = useTranslation();
  const [keyword, setKeyword] = useState('');
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [result, setResult] = useState(null);
  const [notFound, setNotFound] = useState(false);
  const [currentUser, setCurrentUser] = useState(null);
  const [recoverLoading, setRecoverLoading] = useState(false);
  const [showRecoverConfirm, setShowRecoverConfirm] = useState(false);

  useEffect(() => {
    const fetchCurrentUser = async () => {
      try {
        const res = await API.get('/api/user/self');
        if (res.data.success) {
          setCurrentUser(res.data.data);
        }
      } catch {
        setCurrentUser(null);
      }
    };
    fetchCurrentUser();
  }, []);

  const handleSearch = async () => {
    if (!keyword.trim()) return;

    setLoading(true);
    setNotFound(false);
    try {
      const res = await API.get(`/api/archived-user/check?keyword=${encodeURIComponent(keyword.trim())}`);
      const { success, found, data } = res.data;
      if (success && found) {
        setResult(data);
        setModalVisible(true);
      } else {
        setNotFound(true);
      }
    } catch {
      setNotFound(true);
    }
    setLoading(false);
  };

  const canRecover = () => {
    if (!currentUser || !result) return false;
    if (!currentUser.linux_do_username || !result.linux_do_username) return false;
    return currentUser.linux_do_username === result.linux_do_username;
  };

  const handleRecoverQuota = () => {
    setShowRecoverConfirm(true);
  };

  const confirmRecoverQuota = async () => {
    setShowRecoverConfirm(false);
    setRecoverLoading(true);

    try {
      const res = await API.post('/api/archived-user/recover-quota', {
        archived_id: result.id,
      });
      const { success, message, data } = res.data;

      if (success) {
        Toast.success(t('额度恢复成功！已恢复 {{quota}}', { quota: renderQuota(data?.recovered_quota || 0, 2) }));
        setModalVisible(false);
        setResult(null);
        setKeyword('');
      } else {
        Toast.error(message || t('恢复失败，请稍后重试'));
      }
    } catch (error) {
      Toast.error(error.response?.data?.message || t('恢复失败，请稍后重试'));
    }

    setRecoverLoading(false);
  };

  return (
    <div className='p-4 rounded-lg border border-dashed border-gray-300 dark:border-gray-600'>
      <Text strong className='block mb-2'>{t('账号清理查询')}</Text>
      <Text type='tertiary' size='small' className='block mb-3'>
        {t('输入用户名、显示名称或用户ID查询账号是否被清理')}
      </Text>
      <div className='flex gap-2'>
        <Input
          value={keyword}
          onChange={setKeyword}
          placeholder={t('用户名/显示名称/用户ID')}
          onEnterPress={handleSearch}
          style={{ flex: 1 }}
        />
        <Button
          icon={<IconSearch />}
          onClick={handleSearch}
          loading={loading}
        >
          {t('查询')}
        </Button>
      </div>
      {notFound && (
        <Text type='success' size='small' className='block mt-2'>
          {t('未找到被清理的账号，您的账号状态正常')}
        </Text>
      )}

      <Modal
        title={t('账号已被清理')}
        visible={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={
          <div className='flex gap-2 justify-end'>
            {canRecover() && (
              <Button
                type='primary'
                theme='solid'
                onClick={handleRecoverQuota}
                loading={recoverLoading}
                disabled={recoverLoading}
              >
                {t('恢复额度')}
              </Button>
            )}
            <Button onClick={() => setModalVisible(false)}>
              {t('我知道了')}
            </Button>
          </div>
        }
      >
        {result && (
          <div className='space-y-3'>
            <div
              className='p-3 rounded'
              style={{ background: 'var(--semi-color-danger-light-default)' }}
            >
              <Text type='danger'>
                {t('由于账号不活跃，已被清理')}
              </Text>
            </div>

            {canRecover() && (
              <div
                className='p-3 rounded'
                style={{ background: 'var(--semi-color-success-light-default)' }}
              >
                <Text type='success'>
                  {t('检测到您的 LinuxDO 账号与此归档账号匹配，可以恢复额度到当前账号')}
                </Text>
              </div>
            )}

            <div className='grid grid-cols-2 gap-2'>
              <div>
                <Text type='tertiary' size='small'>{t('用户名')}</Text>
                <Text className='block'>{result.username}</Text>
              </div>
              <div>
                <Text type='tertiary' size='small'>{t('显示名称')}</Text>
                <Text className='block'>{result.display_name || '-'}</Text>
              </div>
              <div>
                <Text type='tertiary' size='small'>{t('额度')}</Text>
                <Text className='block'>{renderQuota(result.quota, 2)}</Text>
              </div>
              <div>
                <Text type='tertiary' size='small'>{t('已用额度')}</Text>
                <Text className='block'>{renderQuota(result.used_quota, 2)}</Text>
              </div>
              <div>
                <Text type='tertiary' size='small'>{t('请求次数')}</Text>
                <Text className='block'>{result.request_count}</Text>
              </div>
              <div>
                <Text type='tertiary' size='small'>{t('清理时间')}</Text>
                <Text className='block'>{timestamp2string(result.archived_at)}</Text>
              </div>
            </div>
          </div>
        )}
      </Modal>

      <Modal
        title={t('确认恢复额度')}
        visible={showRecoverConfirm}
        onCancel={() => setShowRecoverConfirm(false)}
        footer={
          <div className='flex gap-2 justify-end'>
            <Button onClick={() => setShowRecoverConfirm(false)}>
              {t('取消')}
            </Button>
            <Button
              type='primary'
              theme='solid'
              onClick={confirmRecoverQuota}
            >
              {t('确认恢复')}
            </Button>
          </div>
        }
      >
        <div className='space-y-3'>
          <Text>{t('确认要恢复此账号的额度吗？')}</Text>
          {result && (
            <div className='p-3 rounded bg-gray-100 dark:bg-gray-800'>
              <Text type='tertiary' size='small'>{t('将恢复额度')}</Text>
              <Text className='block text-lg font-semibold'>{renderQuota(result.quota, 2)}</Text>
            </div>
          )}
          <Text type='tertiary' size='small'>
            {t('恢复后，额度将添加到您当前账号，归档记录将被删除。')}
          </Text>
        </div>
      </Modal>
    </div>
  );
};

export default ArchivedUserCheck;
