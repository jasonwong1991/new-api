/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React, { useEffect, useState } from 'react';
import { Tabs, TabPane, Card } from '@douyinfe/semi-ui';
import { API } from '../../helpers';
import { useIsMobile } from '../../hooks/common/useIsMobile';
import { useTranslation } from 'react-i18next';
import NoticeModal from '../../components/layout/NoticeModal';
import TutorialTab from './components/TutorialTab';
import LeaderboardTab from './components/LeaderboardTab';
import ProjectCard from './components/ProjectCard';
import FAQTab from './components/FAQTab';
import BanListTab from './components/BanListTab';
import FloatingStoreButton from './components/FloatingStoreButton';
import FloatingChatButton from './components/FloatingChatButton';
import ArchivedUserCheck from './components/ArchivedUserCheck';
import ScrollingNotice from './components/ScrollingNotice';

const Home = () => {
  const { t } = useTranslation();
  const [noticeVisible, setNoticeVisible] = useState(false);
  const isMobile = useIsMobile();

  useEffect(() => {
    const checkNoticeAndShow = async () => {
      const lastCloseDate = localStorage.getItem('notice_close_date');
      const today = new Date().toDateString();
      if (lastCloseDate !== today) {
        try {
          const res = await API.get('/api/notice');
          const { success, data } = res.data;
          if (success && data && data.trim() !== '') {
            setNoticeVisible(true);
          }
        } catch (error) {
          console.error('获取公告失败:', error);
        }
      }
    };

    checkNoticeAndShow();
  }, []);

  return (
    <div className='w-full max-w-7xl mx-auto px-4 py-8 mt-16'>
      <FloatingStoreButton />
      <FloatingChatButton />
      <NoticeModal
        visible={noticeVisible}
        onClose={() => setNoticeVisible(false)}
        isMobile={isMobile}
      />

      <div className='flex flex-col gap-6'>
        <ProjectCard />

        <ScrollingNotice />

        <ArchivedUserCheck />

        <Card>
          <Tabs type='line' defaultActiveKey='tutorial'>
            <TabPane tab={t('教程')} itemKey='tutorial'>
              <TutorialTab />
            </TabPane>
            <TabPane tab={t('常见问题')} itemKey='faq'>
              <FAQTab />
            </TabPane>
            <TabPane tab={t('站点使用排行')} itemKey='leaderboard'>
              <LeaderboardTab />
            </TabPane>
            <TabPane tab={t('封禁名单')} itemKey='banlist'>
              <BanListTab />
            </TabPane>
          </Tabs>
        </Card>
      </div>
    </div>
  );
};

export default Home;
