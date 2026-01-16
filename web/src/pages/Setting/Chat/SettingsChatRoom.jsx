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

import React, { useEffect, useState, useRef } from 'react';
import {
  Button,
  Form,
  Space,
  Spin,
  InputNumber,
  Switch,
  TagInput,
  TextArea,
} from '@douyinfe/semi-ui';
import { IconSaveStroked } from '@douyinfe/semi-icons';
import { API, showError, showSuccess } from '../../../helpers';
import { useTranslation } from 'react-i18next';

export default function SettingsChatRoom() {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [inputs, setInputs] = useState({
    enabled: true,
    message_limit: 1000,
    max_message_length: 8000,
    announcement: '',
    image_enabled: true,
    image_max_bytes: 10485760,
    image_cache_max_bytes: 1073741824,
    anti_hotlink_enabled: true,
    allowed_referers: [],
  });
  const refForm = useRef();

  const fetchSettings = async () => {
    setLoading(true);
    try {
      const res = await API.get('/api/chat/setting');
      if (res.data.success) {
        const data = res.data.data;
        setInputs({
          enabled: data.enabled ?? true,
          message_limit: data.message_limit ?? 1000,
          max_message_length: data.max_message_length ?? 8000,
          announcement: data.announcement ?? '',
          image_enabled: data.image_enabled ?? true,
          image_max_bytes: data.image_max_bytes ?? 10485760,
          image_cache_max_bytes: data.image_cache_max_bytes ?? 1073741824,
          anti_hotlink_enabled: data.anti_hotlink_enabled ?? true,
          allowed_referers: data.allowed_referers ?? [],
        });
        if (refForm.current) {
          refForm.current.setValues({
            enabled: data.enabled ?? true,
            message_limit: data.message_limit ?? 1000,
            max_message_length: data.max_message_length ?? 8000,
            announcement: data.announcement ?? '',
            image_enabled: data.image_enabled ?? true,
            image_max_bytes: data.image_max_bytes ?? 10485760,
            image_cache_max_bytes: data.image_cache_max_bytes ?? 1073741824,
            anti_hotlink_enabled: data.anti_hotlink_enabled ?? true,
            allowed_referers: data.allowed_referers ?? [],
          });
        }
      } else {
        showError(res.data.message || t('获取聊天室设置失败'));
      }
    } catch (error) {
      showError(t('获取聊天室设置失败'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchSettings();
  }, []);

  const handleSubmit = async () => {
    setSaving(true);
    try {
      const res = await API.put('/api/chat/setting', inputs);
      if (res.data.success) {
        showSuccess(t('聊天室设置已保存'));
      } else {
        showError(res.data.message || t('保存失败'));
      }
    } catch (error) {
      showError(t('保存失败'));
    } finally {
      setSaving(false);
    }
  };

  const handleInputChange = (field, value) => {
    setInputs((prev) => ({ ...prev, [field]: value }));
  };

  const formatBytes = (bytes) => {
    if (bytes >= 1073741824) return `${(bytes / 1073741824).toFixed(0)} GB`;
    if (bytes >= 1048576) return `${(bytes / 1048576).toFixed(0)} MB`;
    return `${bytes} B`;
  };

  return (
    <Spin spinning={loading}>
      <Space vertical style={{ width: '100%' }} align="start">
        <Form.Section text={t('聊天室设置')}>
          <Form
            getFormApi={(formAPI) => (refForm.current = formAPI)}
            style={{ width: '100%' }}
          >
            <Form.Slot label={t('启用聊天室')}>
              <Switch
                checked={inputs.enabled}
                onChange={(v) => handleInputChange('enabled', v)}
              />
            </Form.Slot>

            <Form.Slot label={t('消息数量限制')}>
              <InputNumber
                value={inputs.message_limit}
                onChange={(v) => handleInputChange('message_limit', v)}
                min={100}
                max={5000}
                step={100}
                suffix={t('条')}
                style={{ width: 200 }}
              />
              <span className="text-gray-500 text-sm ml-2">
                {t('聊天室最多保留的消息数量')}
              </span>
            </Form.Slot>

            <Form.Slot label={t('单条消息最大长度')}>
              <InputNumber
                value={inputs.max_message_length}
                onChange={(v) => handleInputChange('max_message_length', v)}
                min={100}
                max={50000}
                step={100}
                suffix={t('字符')}
                style={{ width: 200 }}
              />
            </Form.Slot>

            <Form.Slot label={t('聊天室公告')}>
              <TextArea
                placeholder={t('在此输入聊天室置顶公告内容，留空则不显示')}
                value={inputs.announcement}
                onChange={(v) => handleInputChange('announcement', v)}
                autosize={{ minRows: 4, maxRows: 12 }}
                style={{ fontFamily: 'JetBrains Mono, Consolas', width: '100%' }}
              />
              <div className="text-gray-500 text-sm mt-1">
                {t('支持 Markdown 和 HTML 格式')}
              </div>
            </Form.Slot>

            <Form.Slot label={t('启用图片功能')}>
              <Switch
                checked={inputs.image_enabled}
                onChange={(v) => handleInputChange('image_enabled', v)}
              />
            </Form.Slot>

            <Form.Slot label={t('单张图片最大大小')}>
              <InputNumber
                value={inputs.image_max_bytes / 1048576}
                onChange={(v) => handleInputChange('image_max_bytes', v * 1048576)}
                min={1}
                max={50}
                step={1}
                suffix="MB"
                style={{ width: 200 }}
              />
            </Form.Slot>

            <Form.Slot label={t('图片缓存最大容量')}>
              <InputNumber
                value={inputs.image_cache_max_bytes / 1073741824}
                onChange={(v) => handleInputChange('image_cache_max_bytes', v * 1073741824)}
                min={0.1}
                max={100}
                step={0.1}
                suffix="GB"
                style={{ width: 200 }}
              />
              <span className="text-gray-500 text-sm ml-2">
                {t('超出后自动清理旧图片，设为0禁用自动清理')}
              </span>
            </Form.Slot>

            <Form.Slot label={t('启用防盗链')}>
              <Switch
                checked={inputs.anti_hotlink_enabled}
                onChange={(v) => handleInputChange('anti_hotlink_enabled', v)}
              />
            </Form.Slot>

            <Form.Slot label={t('允许的来源域名')}>
              <TagInput
                value={inputs.allowed_referers}
                onChange={(v) => handleInputChange('allowed_referers', v)}
                placeholder={t('输入域名后按回车添加')}
                style={{ width: '100%', maxWidth: 500 }}
              />
              <div className="text-gray-500 text-sm mt-1">
                {t('除本站外允许访问图片的域名，如 example.com')}
              </div>
            </Form.Slot>
          </Form>
        </Form.Section>

        <Space style={{ marginTop: 16 }}>
          <Button
            type="primary"
            theme="solid"
            icon={<IconSaveStroked />}
            onClick={handleSubmit}
            loading={saving}
          >
            {t('保存聊天室设置')}
          </Button>
        </Space>
      </Space>
    </Spin>
  );
}
