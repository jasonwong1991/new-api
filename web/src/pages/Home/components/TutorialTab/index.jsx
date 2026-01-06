import React from 'react';
import { Tabs, TabPane } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
import OpenAIGuide from './OpenAIGuide';
import ClaudeCodeGuide from './ClaudeCodeGuide';
import CodexGuide from './CodexGuide';

const TutorialTab = () => {
  const { t } = useTranslation();

  return (
    <div className='p-4'>
      <Tabs type='line' defaultActiveKey='openai'>
        <TabPane tab={t('OpenAI 配置教程')} itemKey='openai'>
          <div className='pt-4'>
            <OpenAIGuide />
          </div>
        </TabPane>
        <TabPane tab={t('Claude Code 使用教程')} itemKey='claude'>
          <div className='pt-4'>
            <ClaudeCodeGuide />
          </div>
        </TabPane>
        <TabPane tab={t('Codex CLI 使用教程')} itemKey='codex'>
          <div className='pt-4'>
            <CodexGuide />
          </div>
        </TabPane>
      </Tabs>
    </div>
  );
};

export default TutorialTab;
