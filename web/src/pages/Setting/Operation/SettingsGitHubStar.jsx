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
import { Button, Col, Form, Row, Spin, Typography } from '@douyinfe/semi-ui';
import {
  API,
  renderQuota,
  showError,
  showSuccess,
  showWarning,
} from '../../../helpers';
import { useTranslation } from 'react-i18next';

const parseRepos = (value) => {
  try {
    const repos = JSON.parse(value || '[]');
    return Array.isArray(repos) ? repos.join('\n') : '';
  } catch (e) {
    return '';
  }
};

const reposToJson = (text) =>
  JSON.stringify(
    text
      .split('\n')
      .map((r) => r.trim())
      .filter(Boolean),
  );

export default function SettingsGitHubStar(props) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    'github_star_setting.enabled': false,
    'github_star_setting.repos': '',
    'github_star_setting.quota_per_repo': 0,
    'github_star_setting.api_key': '',
  });
  const refForm = useRef();
  const [inputsRow, setInputsRow] = useState(inputs);

  function handleFieldChange(fieldName) {
    return (value) => {
      setInputs((inputs) => ({ ...inputs, [fieldName]: value }));
    };
  }

  function onSubmit() {
    const updates = [];
    if (
      inputs['github_star_setting.enabled'] !==
      inputsRow['github_star_setting.enabled']
    ) {
      updates.push({
        key: 'github_star_setting.enabled',
        value: String(inputs['github_star_setting.enabled']),
      });
    }
    if (
      inputs['github_star_setting.repos'] !==
      inputsRow['github_star_setting.repos']
    ) {
      updates.push({
        key: 'github_star_setting.repos',
        value: reposToJson(inputs['github_star_setting.repos']),
      });
    }
    if (
      inputs['github_star_setting.quota_per_repo'] !==
      inputsRow['github_star_setting.quota_per_repo']
    ) {
      updates.push({
        key: 'github_star_setting.quota_per_repo',
        value: String(inputs['github_star_setting.quota_per_repo']),
      });
    }
    // GitHub Token 不会回显，留空表示不修改
    if (inputs['github_star_setting.api_key']) {
      updates.push({
        key: 'github_star_setting.api_key',
        value: inputs['github_star_setting.api_key'],
      });
    }
    if (!updates.length) return showWarning(t('你似乎并没有修改什么'));

    setLoading(true);
    Promise.all(updates.map((item) => API.put('/api/option/', item)))
      .then((res) => {
        if (res.includes(undefined))
          return showError(t('部分保存失败，请重试'));
        showSuccess(t('保存成功'));
        props.refresh();
      })
      .catch(() => {
        showError(t('保存失败，请重试'));
      })
      .finally(() => {
        setLoading(false);
      });
  }

  useEffect(() => {
    const currentInputs = {
      'github_star_setting.enabled':
        props.options['github_star_setting.enabled'] ?? false,
      'github_star_setting.repos': parseRepos(
        props.options['github_star_setting.repos'],
      ),
      'github_star_setting.quota_per_repo':
        props.options['github_star_setting.quota_per_repo'] ?? 0,
      'github_star_setting.api_key': '',
    };
    setInputs(currentInputs);
    setInputsRow(structuredClone(currentInputs));
    refForm.current.setValues(currentInputs);
  }, [props.options]);

  const disabled = !inputs['github_star_setting.enabled'];

  return (
    <Spin spinning={loading}>
      <Form
        values={inputs}
        getFormApi={(formAPI) => (refForm.current = formAPI)}
        style={{ marginBottom: 15 }}
      >
        <Form.Section text={t('GitHub Star 奖励')}>
          <Typography.Text
            type='tertiary'
            style={{ marginBottom: 16, display: 'block' }}
          >
            {t(
              '用户绑定 GitHub 并 Star 指定仓库后可领取额度奖励，每个仓库每人限领一次。开启后即使未开启 GitHub 登录，用户也可在个人设置中绑定 GitHub。',
            )}
          </Typography.Text>
          <Row gutter={16}>
            <Col xs={24} sm={12} md={8} lg={8} xl={8}>
              <Form.Switch
                field={'github_star_setting.enabled'}
                label={t('启用 GitHub Star 奖励')}
                size='default'
                checkedText='｜'
                uncheckedText='〇'
                onChange={handleFieldChange('github_star_setting.enabled')}
              />
            </Col>
            <Col xs={24} sm={12} md={8} lg={8} xl={8}>
              <Form.InputNumber
                field={'github_star_setting.quota_per_repo'}
                label={t('每个仓库奖励额度')}
                extraText={renderQuota(
                  inputs['github_star_setting.quota_per_repo'] || 0,
                )}
                onChange={handleFieldChange(
                  'github_star_setting.quota_per_repo',
                )}
                min={0}
                disabled={disabled}
              />
            </Col>
            <Col xs={24} sm={12} md={8} lg={8} xl={8}>
              <Form.Input
                field={'github_star_setting.api_key'}
                label={t('GitHub Token（可选）')}
                placeholder={t('留空则不修改')}
                extraText={t('配置后 GitHub API 限额从 60 次/小时提升到 5000 次/小时')}
                mode='password'
                onChange={handleFieldChange('github_star_setting.api_key')}
                disabled={disabled}
              />
            </Col>
          </Row>
          <Row>
            <Col span={24}>
              <Form.TextArea
                field={'github_star_setting.repos'}
                label={t('参与活动的仓库')}
                placeholder={'jasonwong1991/WonderBox\njasonwong1991/easy_proxies'}
                extraText={t('每行一个，格式 owner/repo')}
                autosize={{ minRows: 2, maxRows: 6 }}
                onChange={handleFieldChange('github_star_setting.repos')}
                disabled={disabled}
              />
            </Col>
          </Row>
          <Row>
            <Button size='default' onClick={onSubmit}>
              {t('保存 GitHub Star 奖励设置')}
            </Button>
          </Row>
        </Form.Section>
      </Form>
    </Spin>
  );
}
