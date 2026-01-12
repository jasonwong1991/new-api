import React, { useState } from 'react';
import { Input, Button, Modal, Typography } from '@douyinfe/semi-ui';
import { IconSearch } from '@douyinfe/semi-icons';
import { API } from '../../../helpers';
import { useTranslation } from 'react-i18next';
import { renderQuota } from '../../../helpers/render';
import { timestamp2string } from '../../../helpers/utils';

const { Text, Title } = Typography;

const ArchivedUserCheck = () => {
  const { t } = useTranslation();
  const [keyword, setKeyword] = useState('');
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [result, setResult] = useState(null);
  const [notFound, setNotFound] = useState(false);

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
    } catch (error) {
      setNotFound(true);
    }
    setLoading(false);
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
          <Button onClick={() => setModalVisible(false)}>{t('我知道了')}</Button>
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
    </div>
  );
};

export default ArchivedUserCheck;
