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
  Col,
  Form,
  InputNumber,
  Row,
  Spin,
  Switch,
  Typography,
} from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
import {
  compareObjects,
  API,
  showError,
  showSuccess,
  showWarning,
} from '../../../helpers';

export default function SettingsQuotaLimit(props) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    QuotaLimitEnabled: false,
    QuotaDailyLimit: 0,
    QuotaWeeklyLimit: 0,
    QuotaLimitWhitelistUsers: '',
    QuotaLimitWhitelistGroups: '',
  });
  const refForm = useRef();
  const [inputsRow, setInputsRow] = useState(inputs);

  function onSubmit() {
    const updateArray = compareObjects(inputs, inputsRow);
    if (!updateArray.length) return showWarning(t('你似乎并没有修改什么'));
    const requestQueue = updateArray.map((item) => {
      let value = inputs[item.key];
      if (typeof value === 'boolean') value = String(value);
      if (typeof value === 'number') value = String(value);
      return API.put('/api/option/', { key: item.key, value });
    });
    setLoading(true);
    Promise.all(requestQueue)
      .then((res) => {
        if (res.includes(undefined)) {
          return showError(
            requestQueue.length > 1
              ? t('部分保存失败，请重试')
              : t('保存失败'),
          );
        }
        for (let i = 0; i < res.length; i++) {
          if (!res[i].data.success) {
            return showError(res[i].data.message);
          }
        }
        showSuccess(t('保存成功'));
        props.refresh();
      })
      .catch(() => showError(t('保存失败，请重试')))
      .finally(() => setLoading(false));
  }

  useEffect(() => {
    const currentInputs = { ...inputs };
    for (let key in props.options) {
      if (Object.prototype.hasOwnProperty.call(currentInputs, key)) {
        const raw = props.options[key];
        if (key === 'QuotaLimitEnabled') {
          currentInputs[key] =
            raw === true || raw === 'true' || raw === 1 || raw === '1';
        } else if (key === 'QuotaDailyLimit' || key === 'QuotaWeeklyLimit') {
          currentInputs[key] = Number(raw) || 0;
        } else {
          currentInputs[key] = raw ?? '';
        }
      }
    }
    setInputs(currentInputs);
    setInputsRow(structuredClone(currentInputs));
    if (refForm.current) refForm.current.setValues(currentInputs);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [props.options]);

  return (
    <Spin spinning={loading}>
      <Form
        getFormApi={(formAPI) => (refForm.current = formAPI)}
        style={{ marginBottom: 15 }}
      >
        <Form.Section text={t('额度天/周限额设置')}>
          <Row gutter={16}>
            <Col span={24}>
              <div style={{ marginBottom: 12 }}>
                <Typography.Text strong>{t('启用额度限流')}</Typography.Text>
                <div style={{ marginTop: 4, marginBottom: 4 }}>
                  <Switch
                    checked={!!inputs.QuotaLimitEnabled}
                    checkedText='|'
                    uncheckedText='O'
                    onChange={(value) =>
                      setInputs({ ...inputs, QuotaLimitEnabled: value })
                    }
                  />
                </div>
                <Typography.Text type='tertiary' size='small'>
                  {t('开启后，非白名单用户超过每日/每周额度上限将被拒绝并返回 429。')}
                </Typography.Text>
              </div>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col xs={24} sm={12} md={8}>
              <div style={{ marginBottom: 12 }}>
                <Typography.Text strong>{t('每日额度上限')}</Typography.Text>
                <div style={{ marginTop: 4, marginBottom: 4 }}>
                  <InputNumber
                    value={inputs.QuotaDailyLimit}
                    min={0}
                    step={1000}
                    suffix={'Quota'}
                    style={{ width: '100%' }}
                    onChange={(value) =>
                      setInputs({ ...inputs, QuotaDailyLimit: Number(value) || 0 })
                    }
                  />
                </div>
                <Typography.Text type='tertiary' size='small'>
                  {t('按自然日（今日 00:00 起）累计，0 表示不限制')}
                </Typography.Text>
              </div>
            </Col>
            <Col xs={24} sm={12} md={8}>
              <div style={{ marginBottom: 12 }}>
                <Typography.Text strong>{t('每周额度上限')}</Typography.Text>
                <div style={{ marginTop: 4, marginBottom: 4 }}>
                  <InputNumber
                    value={inputs.QuotaWeeklyLimit}
                    min={0}
                    step={1000}
                    suffix={'Quota'}
                    style={{ width: '100%' }}
                    onChange={(value) =>
                      setInputs({ ...inputs, QuotaWeeklyLimit: Number(value) || 0 })
                    }
                  />
                </div>
                <Typography.Text type='tertiary' size='small'>
                  {t('按自然周（本周一 00:00 起）累计，0 表示不限制')}
                </Typography.Text>
              </div>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col xs={24} md={12}>
              <Form.TextArea
                label={t('白名单用户')}
                field={'QuotaLimitWhitelistUsers'}
                autosize={{ minRows: 2, maxRows: 6 }}
                extraText={t('逗号分隔的用户ID或用户名，例如：1,admin,alice')}
                onChange={(value) =>
                  setInputs({ ...inputs, QuotaLimitWhitelistUsers: value })
                }
              />
            </Col>
            <Col xs={24} md={12}>
              <Form.TextArea
                label={t('白名单分组')}
                field={'QuotaLimitWhitelistGroups'}
                autosize={{ minRows: 2, maxRows: 6 }}
                extraText={t('逗号分隔的分组名称，例如：vip,internal')}
                onChange={(value) =>
                  setInputs({ ...inputs, QuotaLimitWhitelistGroups: value })
                }
              />
            </Col>
          </Row>

          <Row>
            <Button size='default' onClick={onSubmit}>
              {t('保存额度限额设置')}
            </Button>
          </Row>
        </Form.Section>
      </Form>
    </Spin>
  );
}
