import React from 'react';
import { Typography, Card, Banner } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';

const { Title, Text } = Typography;

const OpenAIGuide = () => {
  const { t } = useTranslation();
  const siteAddress = 'https://wzw.pp.ua';

  const codeExample = `OPENAI_API_KEY=sk-xxxxxx
OPENAI_BASE_URL=${siteAddress}/v1`;

  return (
    <div className='space-y-4'>
      <Banner
        type='warning'
        description={t('注意：仅 default 分组的 API Key 支持 OpenAI 格式调用')}
        className='mb-4'
      />

      <Card title={t('基础配置')}>
        <div className='mb-6'>
          <Title heading={6} style={{ marginBottom: 8 }}>
            {t('站点地址')}
          </Title>
          <Text
            copyable
            className='text-lg bg-gray-50 dark:bg-gray-800 px-3 py-1 rounded'
          >
            {siteAddress}
          </Text>
        </div>

        <div>
          <Title heading={6} style={{ marginBottom: 8 }}>
            {t('配置示例')}
          </Title>
          <pre className='bg-gray-50 dark:bg-gray-800 p-4 rounded border border-gray-100 dark:border-gray-700 overflow-x-auto font-mono text-sm'>
            <code>{codeExample}</code>
          </pre>
        </div>
      </Card>

      <Card title={t('使用说明')}>
        <div className='space-y-3'>
          <div className='flex items-start gap-2'>
            <span className='bg-blue-500 text-white rounded-full w-6 h-6 flex items-center justify-center text-sm flex-shrink-0'>
              1
            </span>
            <Text>{t('登录后进入控制台，创建一个新的 API Key')}</Text>
          </div>
          <div className='flex items-start gap-2'>
            <span className='bg-blue-500 text-white rounded-full w-6 h-6 flex items-center justify-center text-sm flex-shrink-0'>
              2
            </span>
            <Text>{t('确保 API Key 属于 default 分组')}</Text>
          </div>
          <div className='flex items-start gap-2'>
            <span className='bg-blue-500 text-white rounded-full w-6 h-6 flex items-center justify-center text-sm flex-shrink-0'>
              3
            </span>
            <Text>{t('将 Base URL 设置为上述站点地址')}</Text>
          </div>
          <div className='flex items-start gap-2'>
            <span className='bg-blue-500 text-white rounded-full w-6 h-6 flex items-center justify-center text-sm flex-shrink-0'>
              4
            </span>
            <Text>{t('使用创建的 API Key 作为 OPENAI_API_KEY')}</Text>
          </div>
        </div>
      </Card>
    </div>
  );
};

export default OpenAIGuide;
