import React, { useState } from 'react';
import { Typography, Card, List, Button, Tabs, TabPane, Banner } from '@douyinfe/semi-ui';
import { IconGithubLogo, IconTerminal, IconCopy } from '@douyinfe/semi-icons';
import { useTranslation } from 'react-i18next';
import { copy, showSuccess } from '../../../../helpers';

const { Title, Text } = Typography;

const CodexGuide = () => {
  const { t } = useTranslation();
  const [activeOS, setActiveOS] = useState('macos');
  const siteAddress = 'https://wzw.pp.ua';

  const installCommands = {
    npm: 'npm install -g @openai/codex',
    brew: 'brew install --cask codex',
  };

  const configContent = `# ~/.codex/config.toml
model = "gpt-4o"
model_provider = "custom-proxy"

[model_providers.custom-proxy]
name = "Custom Proxy"
base_url = "${siteAddress}/v1"
env_key = "OPENAI_API_KEY"
wire_api = "chat"`;

  const envCommands = {
    windows: {
      temporary: `set OPENAI_API_KEY=your-api-key`,
      permanent: `setx OPENAI_API_KEY "your-api-key"`,
    },
    macos: {
      temporary: `export OPENAI_API_KEY=your-api-key`,
      permanent: `# 添加到 ~/.zshrc 或 ~/.bash_profile
echo 'export OPENAI_API_KEY=your-api-key' >> ~/.zshrc
source ~/.zshrc`,
    },
    linux: {
      temporary: `export OPENAI_API_KEY=your-api-key`,
      permanent: `# 添加到 ~/.bashrc
echo 'export OPENAI_API_KEY=your-api-key' >> ~/.bashrc
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
      <Banner
        type='info'
        description={t('OpenAI Codex CLI 是 OpenAI 官方推出的本地编程代理工具，支���通过自定义代理使用')}
        className='mb-4'
      />

      <Card
        title={t('安装 Codex CLI')}
        headerExtraContent={<IconTerminal />}
      >
        <div className='space-y-4'>
          <CodeBlock
            code={installCommands.npm}
            label={t('使用 npm 安装')}
          />
          <CodeBlock
            code={installCommands.brew}
            label={t('使用 Homebrew 安装 (macOS)')}
          />
          <div className='flex items-center gap-2 mt-4'>
            <Text type='secondary'>{t('或从 GitHub 下载二进制文件')}</Text>
            <Button
              icon={<IconGithubLogo />}
              theme='light'
              type='tertiary'
              size='small'
              onClick={() =>
                window.open('https://github.com/openai/codex/releases/latest', '_blank')
              }
            >
              {t('GitHub Releases')}
            </Button>
          </div>
        </div>
      </Card>

      <Card title={t('配置代理')}>
        <List className='mb-4'>
          <List.Item>
            <Text>
              {t('1. 请确保在 "codex" 专用分组创建 API Key')}
            </Text>
          </List.Item>
          <List.Item>
            <Text>
              {t('2. 创建或编辑配置文件 ~/.codex/config.toml')}
            </Text>
          </List.Item>
        </List>
        <CodeBlock
          code={configContent}
          label={t('config.toml 配置示例')}
        />
      </Card>

      <Card title={t('设置环境变量')}>
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
            code={envCommands[activeOS].temporary}
            label={t('临时设置（当前终端会话有效）')}
          />
          <CodeBlock
            code={envCommands[activeOS].permanent}
            label={t('永久设置（需要重启终端生效）')}
          />
        </div>
      </Card>

      <Card title={t('启动使用')}>
        <List>
          <List.Item>
            <Text>
              {t('配置完成后，在终端运行 codex 即可启动')}
            </Text>
          </List.Item>
          <List.Item>
            <Text type='secondary'>
              {t('首次运行会提示登录，选择 "Sign in with API key" 使用 API Key 登录')}
            </Text>
          </List.Item>
        </List>
        <div className='mt-4'>
          <CodeBlock
            code='codex'
            label={t('启动 Codex CLI')}
          />
        </div>
      </Card>

      <Card title={t('注意事项')}>
        <List>
          <List.Item>
            <Text type='warning'>
              {t('请将 your-api-key 替换为您在 codex 分组创建的实际 API Key')}
            </Text>
          </List.Item>
          <List.Item>
            <Text>
              {t('Codex CLI 支持多种模型，可在 config.toml 中修改 model 参数')}
            </Text>
          </List.Item>
          <List.Item>
            <Text>
              {t('更多配置选项请参考官方文档')}
            </Text>
            <Button
              theme='borderless'
              type='primary'
              size='small'
              onClick={() =>
                window.open('https://developers.openai.com/codex', '_blank')
              }
            >
              {t('查看文档')}
            </Button>
          </List.Item>
        </List>
      </Card>
    </div>
  );
};

export default CodexGuide;
