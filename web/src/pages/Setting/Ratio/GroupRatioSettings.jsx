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
import { Button, Col, Form, InputNumber, Row, Spin, Switch, Typography } from '@douyinfe/semi-ui';
import {
  compareObjects,
  API,
  showError,
  showSuccess,
  showWarning,
  verifyJSON,
  toBoolean,
} from '../../../helpers';
import { useTranslation } from 'react-i18next';

export default function GroupRatioSettings(props) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    GroupRatio: '',
    UserUsableGroups: '',
    GroupGroupRatio: '',
    'group_ratio_setting.group_special_usable_group': '',
    AutoGroups: '',
    DefaultUseAutoGroup: false,
  });
  const refForm = useRef();
  const [inputsRow, setInputsRow] = useState(inputs);

  // 动态倍率独立状态（与主表单解耦，避免 JSON 校验副作用影响保存）
  const [dynamicRatio, setDynamicRatio] = useState({
    DynamicRatioEnabled: false,
    DynamicRatioMax: 5,
    DynamicRatioTokenThresholdYi: 100, // 单位: 亿 tokens (= 1e8 tokens)
  });
  const [dynamicRatioLoading, setDynamicRatioLoading] = useState(false);

  async function onSubmit() {
    try {
      await refForm.current
        .validate()
        .then(() => {
          const updateArray = compareObjects(inputs, inputsRow);
          if (!updateArray.length)
            return showWarning(t('你似乎并没有修改什么'));

          const requestQueue = updateArray.map((item) => {
            const value =
              typeof inputs[item.key] === 'boolean'
                ? String(inputs[item.key])
                : inputs[item.key];
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
            .catch((error) => {
              console.error('Unexpected error:', error);
              showError(t('保存失败，请重试'));
            })
            .finally(() => {
              setLoading(false);
            });
        })
        .catch(() => {
          showError(t('请检查输入'));
        });
    } catch (error) {
      showError(t('请检查输入'));
      console.error(error);
    }
  }

  useEffect(() => {
    const currentInputs = {};
    for (let key in props.options) {
      if (Object.keys(inputs).includes(key)) {
        currentInputs[key] = props.options[key];
      }
    }
    setInputs(currentInputs);
    setInputsRow(structuredClone(currentInputs));
    refForm.current.setValues(currentInputs);

    // 同步动态倍率状态（独立于主表单）
    const rawThreshold = props.options.DynamicRatioTokenThreshold;
    let thresholdYi = 100;
    if (rawThreshold !== undefined && rawThreshold !== null && rawThreshold !== '') {
      const tokens = Number(rawThreshold);
      if (Number.isFinite(tokens) && tokens > 0) {
        thresholdYi = Math.max(1, Math.round(tokens / 1e8));
      }
    }
    const nextDynamic = {
      DynamicRatioEnabled: toBoolean(props.options.DynamicRatioEnabled),
      DynamicRatioMax:
        props.options.DynamicRatioMax !== undefined &&
        props.options.DynamicRatioMax !== null &&
        props.options.DynamicRatioMax !== ''
          ? Number(props.options.DynamicRatioMax) || 5
          : 5,
      DynamicRatioTokenThresholdYi: thresholdYi,
    };
    setDynamicRatio(nextDynamic);
  }, [props.options]);

  // 保存动态倍率（独立保存，跳过表单校验与 compareObjects 脏判定）
  const saveDynamicRatio = async () => {
    const enabled = !!dynamicRatio.DynamicRatioEnabled;
    let max = Number(dynamicRatio.DynamicRatioMax);
    if (!Number.isFinite(max) || max < 1) {
      showError(t('动态倍率上限必须为不小于 1 的数字'));
      return;
    }
    if (max > 100) max = 100;
    // 归一化到一位小数，与 InputNumber precision 一致
    max = Math.round(max * 10) / 10;

    const thresholdYi = Number(dynamicRatio.DynamicRatioTokenThresholdYi);
    if (!Number.isFinite(thresholdYi) || thresholdYi < 1) {
      showError(t('动态倍率起始阈值必须为不小于 1 亿 tokens 的整数'));
      return;
    }
    const thresholdTokens = Math.round(thresholdYi) * 100000000;

    setDynamicRatioLoading(true);
    try {
      const results = await Promise.all([
        API.put('/api/option/', {
          key: 'DynamicRatioEnabled',
          value: String(enabled),
        }),
        API.put('/api/option/', {
          key: 'DynamicRatioMax',
          value: max,
        }),
        API.put('/api/option/', {
          key: 'DynamicRatioTokenThreshold',
          value: String(thresholdTokens),
        }),
      ]);
      for (const r of results) {
        if (!r?.data?.success) {
          showError(r?.data?.message || t('动态倍率保存失败'));
          return;
        }
      }
      showSuccess(t('动态倍率设置已保存'));
      if (props.refresh) props.refresh();
    } catch (err) {
      console.error('saveDynamicRatio failed:', err);
      showError(err?.message || t('动态倍率保存失败'));
    } finally {
      setDynamicRatioLoading(false);
    }
  };

  return (
    <Spin spinning={loading}>
      <Form
        values={inputs}
        getFormApi={(formAPI) => (refForm.current = formAPI)}
        style={{ marginBottom: 15 }}
      >
        <Row gutter={16}>
          <Col xs={24} sm={16}>
            <Form.TextArea
              label={t('分组倍率')}
              placeholder={t('为一个 JSON 文本，键为分组名称，值为倍率')}
              extraText={t(
                '分组倍率设置，可以在此处新增分组或修改现有分组的倍率，格式为 JSON 字符串，例如：{"vip": 0.5, "test": 1}，表示 vip 分组的倍率为 0.5，test 分组的倍率为 1',
              )}
              field={'GroupRatio'}
              autosize={{ minRows: 6, maxRows: 12 }}
              trigger='blur'
              stopValidateWithError
              rules={[
                {
                  validator: (rule, value) => verifyJSON(value),
                  message: t('不是合法的 JSON 字符串'),
                },
              ]}
              onChange={(value) => setInputs({ ...inputs, GroupRatio: value })}
            />
          </Col>
        </Row>
        <Row gutter={16}>
          <Col xs={24} sm={16}>
            <Form.TextArea
              label={t('用户可选分组')}
              placeholder={t('为一个 JSON 文本，键为分组名称，值为分组描述')}
              extraText={t(
                '用户新建令牌时可选的分组，格式为 JSON 字符串，例如：{"vip": "VIP 用户", "test": "测试"}，表示用户可以选择 vip 分组和 test 分组',
              )}
              field={'UserUsableGroups'}
              autosize={{ minRows: 6, maxRows: 12 }}
              trigger='blur'
              stopValidateWithError
              rules={[
                {
                  validator: (rule, value) => verifyJSON(value),
                  message: t('不是合法的 JSON 字符串'),
                },
              ]}
              onChange={(value) =>
                setInputs({ ...inputs, UserUsableGroups: value })
              }
            />
          </Col>
        </Row>
        <Row gutter={16}>
          <Col xs={24} sm={16}>
            <Form.TextArea
              label={t('分组特殊倍率')}
              placeholder={t('为一个 JSON 文本')}
              extraText={t(
                '键为分组名称，值为另一个 JSON 对象，键为分组名称，值为该分组的用户的特殊分组倍率，例如：{"vip": {"default": 0.5, "test": 1}}，表示 vip 分组的用户在使用default分组的令牌时倍率为0.5，使用test分组时倍率为1',
              )}
              field={'GroupGroupRatio'}
              autosize={{ minRows: 6, maxRows: 12 }}
              trigger='blur'
              stopValidateWithError
              rules={[
                {
                  validator: (rule, value) => verifyJSON(value),
                  message: t('不是合法的 JSON 字符串'),
                },
              ]}
              onChange={(value) =>
                setInputs({ ...inputs, GroupGroupRatio: value })
              }
            />
          </Col>
        </Row>
        <Row gutter={16}>
          <Col xs={24} sm={16}>
            <Form.TextArea
              label={t('分组特殊可用分组')}
              placeholder={t('为一个 JSON 文本')}
              extraText={t(
                '键为用户分组名称，值为操作映射对象。内层键以"+:"开头表示添加指定分组（键值为分组名称，值为描述），以"-:"开头表示移除指定分组（键值为分组名称），不带前缀的键直接添加该分组。例如：{"vip": {"+:premium": "高级分组", "special": "特殊分组", "-:default": "默认分组"}}，表示 vip 分组的用户可以使用 premium 和 special 分组，同时移除 default 分组的访问权限',
              )}
              field={'group_ratio_setting.group_special_usable_group'}
              autosize={{ minRows: 6, maxRows: 12 }}
              trigger='blur'
              stopValidateWithError
              rules={[
                {
                  validator: (rule, value) => verifyJSON(value),
                  message: t('不是合法的 JSON 字符串'),
                },
              ]}
              onChange={(value) =>
                setInputs({ ...inputs, 'group_ratio_setting.group_special_usable_group': value })
              }
            />
          </Col>
        </Row>
        <Row gutter={16}>
          <Col xs={24} sm={16}>
            <Form.TextArea
              label={t('自动分组auto，从第一个开始选择')}
              placeholder={t('为一个 JSON 文本')}
              field={'AutoGroups'}
              autosize={{ minRows: 6, maxRows: 12 }}
              trigger='blur'
              stopValidateWithError
              rules={[
                {
                  validator: (rule, value) => {
                    if (!value || value.trim() === '') {
                      return true; // Allow empty values
                    }

                    // First check if it's valid JSON
                    try {
                      const parsed = JSON.parse(value);

                      // Check if it's an array
                      if (!Array.isArray(parsed)) {
                        return false;
                      }

                      // Check if every element is a string
                      return parsed.every((item) => typeof item === 'string');
                    } catch (error) {
                      return false;
                    }
                  },
                  message: t('必须是有效的 JSON 字符串数组，例如：["g1","g2"]'),
                },
              ]}
              onChange={(value) => setInputs({ ...inputs, AutoGroups: value })}
            />
          </Col>
        </Row>
        <Row gutter={16}>
          <Col span={16}>
            <div style={{ marginBottom: 8 }}>
              <Typography.Text strong>{t('启用动态倍率')}</Typography.Text>
              <div style={{ marginTop: 4, marginBottom: 8 }}>
                <Switch
                  checked={!!dynamicRatio.DynamicRatioEnabled}
                  checkedText='|'
                  uncheckedText='O'
                  onChange={(value) =>
                    setDynamicRatio((prev) => ({
                      ...prev,
                      DynamicRatioEnabled: value,
                    }))
                  }
                />
              </div>
              <Typography.Text type='tertiary' size='small'>
                {t('开启后，分组倍率将根据平台当日(0点起)Token用量和当前RPM动态调整。起始阈值以下倍率为1，超出后按公式计算，最大值与起始阈值可通过右侧配置。')}
              </Typography.Text>
            </div>
          </Col>
          <Col xs={24} sm={8}>
            <div style={{ marginBottom: 8 }}>
              <Typography.Text strong>{t('动态倍率上限')}</Typography.Text>
              <div style={{ marginTop: 4, marginBottom: 8, display: 'flex', gap: 8, alignItems: 'center' }}>
                <InputNumber
                  value={dynamicRatio.DynamicRatioMax}
                  min={1}
                  max={100}
                  step={0.1}
                  precision={1}
                  onChange={(value) =>
                    setDynamicRatio((prev) => ({
                      ...prev,
                      DynamicRatioMax: value,
                    }))
                  }
                />
                <Button
                  onClick={saveDynamicRatio}
                  loading={dynamicRatioLoading}
                  type='primary'
                  theme='solid'
                  size='default'
                >
                  {t('保存动态倍率')}
                </Button>
              </div>
              <Typography.Text type='tertiary' size='small'>
                {t('动态倍率的最大值，默认为5。该配置与分组倍率独立保存。')}
              </Typography.Text>
            </div>
            <div style={{ marginTop: 12, marginBottom: 8 }}>
              <Typography.Text strong>{t('动态倍率起始阈值（亿 tokens）')}</Typography.Text>
              <div style={{ marginTop: 4, marginBottom: 8, display: 'flex', gap: 8, alignItems: 'center' }}>
                <InputNumber
                  value={dynamicRatio.DynamicRatioTokenThresholdYi}
                  min={1}
                  max={100000}
                  step={1}
                  precision={0}
                  onChange={(value) =>
                    setDynamicRatio((prev) => ({
                      ...prev,
                      DynamicRatioTokenThresholdYi: value,
                    }))
                  }
                />
              </div>
              <Typography.Text type='tertiary' size='small'>
                {t('当日 Tokens 超过该阈值后才开始计算动态倍率，默认 100 亿。')}
              </Typography.Text>
            </div>
          </Col>
        </Row>
        <Row gutter={16}>
          <Col span={16}>
            <Form.Switch
              label={t(
                '创建令牌默认选择auto分组，初始令牌也将设为auto（否则留空，为用户默认分组）',
              )}
              field={'DefaultUseAutoGroup'}
              onChange={(value) =>
                setInputs({ ...inputs, DefaultUseAutoGroup: value })
              }
            />
          </Col>
        </Row>
      </Form>
      <Button onClick={onSubmit}>{t('保存分组倍率设置')}</Button>
    </Spin>
  );
}
