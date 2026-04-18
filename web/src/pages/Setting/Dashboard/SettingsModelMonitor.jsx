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
import {
  Button,
  Form,
  Typography,
  Switch,
  InputNumber,
  Select,
  Divider,
  Space,
} from '@douyinfe/semi-ui';
import { Activity, Save } from 'lucide-react';
import { API, showError, showSuccess } from '../../../helpers';
import { useTranslation } from 'react-i18next';

const { Text } = Typography;

const SORT_BY_OPTIONS = [
  { value: 'request_count', label: '请求数' },
  { value: 'success_rate', label: '成功率' },
  { value: 'avg_latency', label: '平均延迟' },
  { value: 'quota', label: '消耗额度' },
  { value: 'total_tokens', label: '总Tokens' },
];

const SORT_ORDER_OPTIONS = [
  { value: 'desc', label: '降序' },
  { value: 'asc', label: '升序' },
];

const WINDOW_OPTIONS = [
  { value: '1h', label: '最近1小时' },
  { value: '24h', label: '最近24小时' },
  { value: '7d', label: '最近7天' },
];

const DEFAULT_CONFIG = {
  enabled: true,
  models: [],
  sort_by: 'request_count',
  sort_order: 'desc',
  default_window: '24h',
  limit: 20,
  refresh_sec: 30,
};

const SettingsModelMonitor = () => {
  const { t } = useTranslation();
  const [config, setConfig] = useState(DEFAULT_CONFIG);
  const [allModels, setAllModels] = useState([]);
  const [loading, setLoading] = useState(false);

  const loadConfig = async () => {
    try {
      const res = await API.get('/api/model_monitor/config');
      const { success, data, message } = res.data || {};
      if (success && data) {
        setConfig({ ...DEFAULT_CONFIG, ...data, models: data.models || [] });
      } else if (!success) {
        showError(message);
      }
    } catch (e) {
      showError(e.message);
    }
  };

  const loadModels = async () => {
    try {
      const res = await API.get('/api/channel/models_enabled');
      const { success, data } = res.data || {};
      if (success && Array.isArray(data)) {
        setAllModels(data);
      }
    } catch (e) {
      // silent
    }
  };

  useEffect(() => {
    loadConfig();
    loadModels();
  }, []);

  const save = async () => {
    setLoading(true);
    try {
      const res = await API.put('/api/model_monitor/config', config);
      const { success, message } = res.data || {};
      if (success) {
        showSuccess(t('模型监控配置已保存'));
      } else {
        showError(message || t('保存失败'));
      }
    } catch (e) {
      showError(e.message);
    } finally {
      setLoading(false);
    }
  };

  const updateField = (key, value) => {
    setConfig((c) => ({ ...c, [key]: value }));
  };

  return (
    <Form.Section
      text={
        <div className='flex flex-col w-full'>
          <div className='mb-2'>
            <div className='flex items-center text-blue-500'>
              <Activity size={16} className='mr-2' />
              <Text>{t('模型监控页面设置（仅 Root 可修改）')}</Text>
            </div>
          </div>
          <Divider margin='12px' />
        </div>
      }
    >
      <div className='flex flex-col gap-4'>
        <div className='flex items-center gap-2'>
          <Switch
            checked={config.enabled}
            onChange={(v) => updateField('enabled', v)}
          />
          <Text>{config.enabled ? t('已启用') : t('已禁用')}</Text>
        </div>

        <div className='grid grid-cols-1 md:grid-cols-2 gap-4'>
          <div>
            <Text strong>{t('展示模型（空表示全部）')}</Text>
            <Select
              multiple
              filter
              value={config.models}
              onChange={(v) => updateField('models', v || [])}
              optionList={allModels.map((m) => ({ value: m, label: m }))}
              placeholder={t('选择要展示的模型')}
              style={{ width: '100%', marginTop: 4 }}
            />
          </div>
          <div>
            <Text strong>{t('排序字段')}</Text>
            <Select
              value={config.sort_by}
              onChange={(v) => updateField('sort_by', v)}
              optionList={SORT_BY_OPTIONS.map((o) => ({
                value: o.value,
                label: t(o.label),
              }))}
              style={{ width: '100%', marginTop: 4 }}
            />
          </div>
          <div>
            <Text strong>{t('排序方向')}</Text>
            <Select
              value={config.sort_order}
              onChange={(v) => updateField('sort_order', v)}
              optionList={SORT_ORDER_OPTIONS.map((o) => ({
                value: o.value,
                label: t(o.label),
              }))}
              style={{ width: '100%', marginTop: 4 }}
            />
          </div>
          <div>
            <Text strong>{t('默认时间窗口')}</Text>
            <Select
              value={config.default_window}
              onChange={(v) => updateField('default_window', v)}
              optionList={WINDOW_OPTIONS.map((o) => ({
                value: o.value,
                label: t(o.label),
              }))}
              style={{ width: '100%', marginTop: 4 }}
            />
          </div>
          <div>
            <Text strong>{t('展示数量上限（0 表示不限）')}</Text>
            <InputNumber
              min={0}
              max={200}
              value={config.limit}
              onChange={(v) => updateField('limit', Number(v) || 0)}
              style={{ width: '100%', marginTop: 4 }}
            />
          </div>
          <div>
            <Text strong>{t('自动刷新间隔（秒）')}</Text>
            <InputNumber
              min={5}
              max={3600}
              value={config.refresh_sec}
              onChange={(v) => updateField('refresh_sec', Number(v) || 30)}
              style={{ width: '100%', marginTop: 4 }}
            />
          </div>
        </div>

        <Space>
          <Button
            icon={<Save size={14} />}
            type='primary'
            theme='solid'
            loading={loading}
            onClick={save}
          >
            {t('保存设置')}
          </Button>
        </Space>
      </div>
    </Form.Section>
  );
};

export default SettingsModelMonitor;
