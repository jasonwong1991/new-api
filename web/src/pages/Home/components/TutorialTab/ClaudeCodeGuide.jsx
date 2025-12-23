import React, { useState } from 'react';
import { Typography, Card, List, Button, Tabs, TabPane } from '@douyinfe/semi-ui';
import { IconGithubLogo, IconTerminal, IconCopy } from '@douyinfe/semi-icons';
import { useTranslation } from 'react-i18next';
import { copy, showSuccess } from '../../../../helpers';

const { Title, Text } = Typography;

const ClaudeCodeGuide = () => {
  const { t } = useTranslation();
  const [activeOS, setActiveOS] = useState('macos');
  const siteAddress = 'https://wzw.pp.ua';

  const commands = {
    windows: {
      temporary: `set ANTHROPIC_BASE_URL=${siteAddress}
set ANTHROPIC_AUTH_TOKEN=your-api-key`,
      permanent: `setx ANTHROPIC_BASE_URL "${siteAddress}"
setx ANTHROPIC_AUTH_TOKEN "your-api-key"`,
    },
    macos: {
      temporary: `export ANTHROPIC_BASE_URL=${siteAddress}
export ANTHROPIC_AUTH_TOKEN=your-api-key`,
      permanent: `# 添加到 ~/.zshrc 或 ~/.bash_profile
echo 'export ANTHROPIC_BASE_URL=${siteAddress}' >> ~/.zshrc
echo 'export ANTHROPIC_AUTH_TOKEN=your-api-key' >> ~/.zshrc
source ~/.zshrc`,
    },
    linux: {
      temporary: `export ANTHROPIC_BASE_URL=${siteAddress}
export ANTHROPIC_AUTH_TOKEN=your-api-key`,
      permanent: `# 添加到 ~/.bashrc
echo 'export ANTHROPIC_BASE_URL=${siteAddress}' >> ~/.bashrc
echo 'export ANTHROPIC_AUTH_TOKEN=your-api-key' >> ~/.bashrc
source ~/.bashrc`,
    },
  };

  const handleCopy = async (text) => {
    const ok = await copy(text);
    if (ok) {
      showSuccess(t('已复制到剪切板'));
    }
  };

  const CodeBlock = ({ code, label }) => (
    <div className='relative'>
      <div className='flex items-center justify-between mb-2'>
        <Text type='secondary' size='small'>
          {label}
        </Text>
        <Button
          icon={<IconCopy />}
          size='small'
          theme='borderless'
          onClick={() => handleCopy(code)}
        />
      </div>
      <pre className='bg-gray-50 dark:bg-gray-800 p-4 rounded border border-gray-100 dark:border-gray-700 overflow-x-auto font-mono text-sm whitespace-pre-wrap'>
        <code>{code}</code>
      </pre>
    </div>
  );

  return (
    <div className='space-y-4'>
      <Card
        title={t('使用前准备')}
        headerExtraContent={<IconTerminal />}
      >
        <List>
          <List.Item>
            <Text>
              {t('1. 请确保在 "claude code" 专用分组创建 API Key')}
            </Text>
          </List.Item>
          <List.Item>
            <div className='flex items-center gap-2 flex-wrap'>
              <Text>{t('2. 推荐使用 cc-switch 工具快速切换环境')}</Text>
              <Button
                icon={<IconGithubLogo />}
                theme='light'
                type='tertiary'
                size='small'
                onClick={() =>
                  window.open('https://github.com/farion1231/cc-switch/releases', '_blank')
                }
              >
                {t('下载 cc-switch')}
              </Button>
            </div>
          </List.Item>
          <List.Item>
            <Text type='secondary'>
              {t('cc-switch 是一个图形化工具，可以方便地管理多个 Claude Code 配置')}
            </Text>
          </List.Item>
        </List>
      </Card>

      <Card title={t('终端配置指南')}>
        <Tabs
          activeKey={activeOS}
          onChange={setActiveOS}
          type='button'
          className='mb-4'
        >
          <TabPane tab='macOS' itemKey='macos' />
          <TabPane tab='Linux' itemKey='linux' />
          <TabPane tab='Windows' itemKey='windows' />
        </Tabs>

        <div className='space-y-6'>
          <CodeBlock
            code={commands[activeOS].temporary}
            label={t('临时设置（当前终端会话有效）')}
          />
          <CodeBlock
            code={commands[activeOS].permanent}
            label={t('永久设置（需要重启终端生效）')}
          />
        </div>
      </Card>

      <Card title={t('注意事项')}>
        <List>
          <List.Item>
            <Text type='warning'>
              {t('请将 your-api-key 替换为您在 claude code 分组创建的实际 API Key')}
            </Text>
          </List.Item>
          <List.Item>
            <Text>
              {t('永久设置后需要重新打开终端或执行 source 命令才能生效')}
            </Text>
          </List.Item>
          <List.Item>
            <Text>
              {t('Windows 用户使用 setx 命令后需要重新打开命令提示符')}
            </Text>
          </List.Item>
        </List>
      </Card>
    </div>
  );
};

export default ClaudeCodeGuide;
