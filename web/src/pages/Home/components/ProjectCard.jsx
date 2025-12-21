import React, { useEffect, useState } from 'react';
import { Card, Typography, Skeleton } from '@douyinfe/semi-ui';
import { IconGithubLogo, IconStar } from '@douyinfe/semi-icons';
import { useTranslation } from 'react-i18next';

const ProjectCard = () => {
  const { t } = useTranslation();
  const [stars, setStars] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch('https://api.github.com/repos/jasonwong1991/easy_proxies')
      .then((res) => res.json())
      .then((data) => {
        setStars(data.stargazers_count);
        setLoading(false);
      })
      .catch(() => setLoading(false));
  }, []);

  return (
    <Card
      className='cursor-pointer hover:shadow-md transition-shadow duration-300'
      onClick={() =>
        window.open('https://github.com/jasonwong1991/easy_proxies', '_blank')
      }
      bodyStyle={{ padding: '16px 24px' }}
    >
      <div className='flex items-center justify-between'>
        <div className='flex items-center gap-4'>
          <div className='p-2 bg-gray-100 rounded-full dark:bg-gray-700'>
            <IconGithubLogo size='large' />
          </div>
          <div>
            <Typography.Title heading={5} style={{ margin: 0 }}>
              easy_proxies
            </Typography.Title>
            <Typography.Text type='secondary'>
              {t('超简单拥有自己的代理池，只需要有节点即可')}
            </Typography.Text>
          </div>
        </div>
        <div className='flex items-center gap-1 bg-gray-50 dark:bg-gray-800 px-3 py-1 rounded-full border border-gray-200 dark:border-gray-700'>
          <IconStar style={{ color: '#fbbf24' }} />
          <span className='font-medium'>
            {loading ? (
              <Skeleton.Title style={{ width: 30, height: 20 }} />
            ) : (
              stars ?? '-'
            )}
          </span>
        </div>
      </div>
    </Card>
  );
};

export default ProjectCard;
