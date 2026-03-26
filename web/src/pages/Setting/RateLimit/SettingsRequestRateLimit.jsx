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

import React, { useEffect, useState, useRef, useCallback } from 'react';
import {
  Button,
  Col,
  Form,
  Row,
  Spin,
  Banner,
  Select,
  TagGroup,
  Typography,
} from '@douyinfe/semi-ui';
import {
  compareObjects,
  API,
  showError,
  showSuccess,
  showWarning,
  verifyJSON,
} from '../../../helpers';
import { useTranslation } from 'react-i18next';
import { debounce } from 'lodash';

export default function RequestRateLimit(props) {
  const { t } = useTranslation();

  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    ModelRequestRateLimitEnabled: false,
    ModelRequestRateLimitCount: -1,
    ModelRequestRateLimitSuccessCount: 1000,
    ModelRequestRateLimitDurationMinutes: 1,
    ModelRequestRateLimitGroup: '',
    RateLimitExemptWhitelist: '',
  });
  const refForm = useRef();
  const [inputsRow, setInputsRow] = useState(inputs);

  // Whitelist user search states
  const [searching, setSearching] = useState(false);
  const [searchResults, setSearchResults] = useState([]);
  const [whitelistUsers, setWhitelistUsers] = useState([]);
  const [originalWhitelistUsers, setOriginalWhitelistUsers] = useState([]);

  // Debounced search function
  const searchUsers = useCallback(
    debounce(async (keyword) => {
      if (!keyword || keyword.trim().length < 1) {
        setSearchResults([]);
        return;
      }
      setSearching(true);
      try {
        const res = await API.get(
          `/api/user/search?keyword=${encodeURIComponent(keyword)}&p=1&page_size=10`,
        );
        const { success, data } = res.data;
        if (success && data?.items) {
          const options = data.items
            .filter((user) => !whitelistUsers.some((u) => u.id === user.id))
            .map((user) => ({
              value: user.id,
              label: `${user.username} (ID: ${user.id})${user.display_name ? ` - ${user.display_name}` : ''}`,
              user: user,
            }));
          setSearchResults(options);
        }
      } catch (error) {
        console.error('Search users failed:', error);
      } finally {
        setSearching(false);
      }
    }, 300),
    [whitelistUsers],
  );

  const handleUserSelect = (value, option) => {
    if (value) {
      const selectedOption = searchResults.find(opt => opt.value === value);
      if (selectedOption?.user) {
        const user = selectedOption.user;
        if (!whitelistUsers.some((u) => u.id === user.id)) {
          setWhitelistUsers([...whitelistUsers, { id: user.id, username: user.username }]);
        }
      }
    }
    setSearchResults([]);
  };

  const handleUserRemove = (userId) => {
    setWhitelistUsers(whitelistUsers.filter((u) => u.id !== userId));
  };

  function onSubmit() {
    // Check whitelist changes
    const currentWhitelistIds = whitelistUsers.map((u) => u.id).sort((a, b) => a - b);
    const originalWhitelistIds = originalWhitelistUsers.map((u) => u.id).sort((a, b) => a - b);
    const whitelistChanged = JSON.stringify(currentWhitelistIds) !== JSON.stringify(originalWhitelistIds);

    // Check other input changes (excluding RateLimitExemptWhitelist which we handle separately)
    const inputsWithoutWhitelist = { ...inputs };
    const inputsRowWithoutWhitelist = { ...inputsRow };
    delete inputsWithoutWhitelist.RateLimitExemptWhitelist;
    delete inputsRowWithoutWhitelist.RateLimitExemptWhitelist;
    const updateArray = compareObjects(inputsWithoutWhitelist, inputsRowWithoutWhitelist);

    if (!updateArray.length && !whitelistChanged) {
      return showWarning(t('你似乎并没有修改什么'));
    }

    const requestQueue = updateArray.map((item) => {
      let value = '';
      if (typeof inputs[item.key] === 'boolean') {
        value = String(inputs[item.key]);
      } else {
        value = inputs[item.key];
      }
      return API.put('/api/option/', {
        key: item.key,
        value,
      });
    });

    // Add whitelist update if changed
    if (whitelistChanged) {
      const whitelistValue = currentWhitelistIds.length > 0 ? JSON.stringify(currentWhitelistIds) : '';
      requestQueue.push(
        API.put('/api/option/', {
          key: 'RateLimitExemptWhitelist',
          value: whitelistValue,
        }),
      );
    }
    setLoading(true);
    Promise.all(requestQueue)
      .then((res) => {
        if (requestQueue.length === 1) {
          if (res.includes(undefined)) return;
        } else if (requestQueue.length > 1) {
          if (res.includes(undefined))
            return showError(t('部分保存失败，请重试'));
        }

        for (let i = 0; i < res.length; i++) {
          if (!res[i].data.success) {
            return showError(res[i].data.message);
          }
        }

        showSuccess(t('保存成功'));
        setOriginalWhitelistUsers([...whitelistUsers]);
        props.refresh();
      })
      .catch(() => {
        showError(t('保存失败，请重试'));
      })
      .finally(() => {
        setLoading(false);
      });
  }

  // Parse whitelist JSON and fetch user info
  const initWhitelistUsers = useCallback(async (whitelistJson) => {
    if (!whitelistJson || whitelistJson.trim() === '') {
      setWhitelistUsers([]);
      setOriginalWhitelistUsers([]);
      return;
    }
    try {
      const userIds = JSON.parse(whitelistJson);
      if (!Array.isArray(userIds) || userIds.length === 0) {
        setWhitelistUsers([]);
        setOriginalWhitelistUsers([]);
        return;
      }
      // Fetch user info for each ID
      const users = await Promise.all(
        userIds.map(async (id) => {
          try {
            const res = await API.get(`/api/user/${id}`);
            if (res.data.success && res.data.data) {
              return { id: res.data.data.id, username: res.data.data.username };
            }
          } catch {
            // User not found, just use ID
          }
          return { id, username: `ID: ${id}` };
        }),
      );
      setWhitelistUsers(users);
      setOriginalWhitelistUsers(users);
    } catch {
      setWhitelistUsers([]);
      setOriginalWhitelistUsers([]);
    }
  }, []);

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

    // Initialize whitelist users
    if (props.options?.RateLimitExemptWhitelist) {
      initWhitelistUsers(props.options.RateLimitExemptWhitelist);
    } else {
      setWhitelistUsers([]);
      setOriginalWhitelistUsers([]);
    }
  }, [props.options]);

  return (
    <>
      <Spin spinning={loading}>
        <Form
          values={inputs}
          getFormApi={(formAPI) => (refForm.current = formAPI)}
          style={{ marginBottom: 15 }}
        >
          <Form.Section text={t('速率限制豁免白名单')}>
            <Row gutter={16}>
              <Col span={24}>
                <Banner
                  type='warning'
                  icon={null}
                  description={t(
                    '白名单用户将完全绕过所有速率限制（最高优先级）。请谨慎添加用户。',
                  )}
                  style={{ marginBottom: 16 }}
                />
              </Col>
            </Row>
            <Row gutter={16}>
              <Col xs={24} sm={16}>
                <Typography.Text strong style={{ display: 'block', marginBottom: 8 }}>
                  {t('豁免用户列表')}
                </Typography.Text>
                <Select
                  style={{ width: '100%', marginBottom: 12 }}
                  placeholder={t('输入用户名、邮箱或ID搜索用户')}
                  filter
                  remote
                  onSearch={searchUsers}
                  loading={searching}
                  optionList={searchResults}
                  onChange={handleUserSelect}
                  value={null}
                  emptyContent={
                    searching ? t('搜索中...') : t('输入关键词搜索用户')
                  }
                  showClear
                />
                {whitelistUsers.length > 0 && (
                  <TagGroup
                    maxTagCount={20}
                    style={{
                      display: 'flex',
                      flexWrap: 'wrap',
                      gap: 8,
                    }}
                    tagList={whitelistUsers.map((user) => ({
                      tagKey: String(user.id),
                      children: `${user.username} (ID: ${user.id})`,
                      closable: true,
                      onClose: () => handleUserRemove(user.id),
                    }))}
                  />
                )}
                {whitelistUsers.length === 0 && (
                  <Typography.Text type='tertiary'>
                    {t('暂无豁免用户')}
                  </Typography.Text>
                )}
                <div
                  style={{ marginTop: 8, fontSize: 12, color: 'var(--semi-color-text-2)' }}
                >
                  <p>{t('说明：')}</p>
                  <ul style={{ paddingLeft: 20, margin: 0 }}>
                    <li>{t('支持通过用户名、显示名称、邮箱或ID搜索用户')}</li>
                    <li>{t('点击标签上的 × 可删除该用户')}</li>
                    <li>{t('白名单用户优先级最高，将完全绕过所有速率限制')}</li>
                    <li>{t('建议仅添加管理员或测试账号')}</li>
                  </ul>
                </div>
              </Col>
            </Row>
            <Row style={{ marginTop: 16 }}>
              <Button size='default' onClick={onSubmit}>
                {t('保存白名单配置')}
              </Button>
            </Row>
          </Form.Section>
          <Form.Section text={t('模型请求速率限制')}>
            <Row gutter={16}>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.Switch
                  field={'ModelRequestRateLimitEnabled'}
                  label={t('启用用户模型请求速率限制（可能会影响高并发性能）')}
                  size='default'
                  checkedText='｜'
                  uncheckedText='〇'
                  onChange={(value) => {
                    setInputs({
                      ...inputs,
                      ModelRequestRateLimitEnabled: value,
                    });
                  }}
                />
              </Col>
            </Row>
            <Row>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.InputNumber
                  label={t('限制周期')}
                  step={1}
                  min={0}
                  suffix={t('分钟')}
                  extraText={t('频率限制的周期（分钟）')}
                  field={'ModelRequestRateLimitDurationMinutes'}
                  onChange={(value) =>
                    setInputs({
                      ...inputs,
                      ModelRequestRateLimitDurationMinutes: String(value),
                    })
                  }
                />
              </Col>
            </Row>
            <Row>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.InputNumber
                  label={t('用户每周期最多请求次数')}
                  step={1}
                  min={0}
                  max={100000000}
                  suffix={t('次')}
                  extraText={t('包括失败请求的次数，0代表不限制')}
                  field={'ModelRequestRateLimitCount'}
                  onChange={(value) =>
                    setInputs({
                      ...inputs,
                      ModelRequestRateLimitCount: String(value),
                    })
                  }
                />
              </Col>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.InputNumber
                  label={t('用户每周期最多请求完成次数')}
                  step={1}
                  min={1}
                  max={100000000}
                  suffix={t('次')}
                  extraText={t('只包括请求成功的次数')}
                  field={'ModelRequestRateLimitSuccessCount'}
                  onChange={(value) =>
                    setInputs({
                      ...inputs,
                      ModelRequestRateLimitSuccessCount: String(value),
                    })
                  }
                />
              </Col>
            </Row>
            <Row>
              <Col xs={24} sm={16}>
                <Form.TextArea
                  label={t('分组速率限制')}
                  placeholder={t(
                    '{\n  "default": [200, 100],\n  "vip": [0, 1000]\n}',
                  )}
                  field={'ModelRequestRateLimitGroup'}
                  autosize={{ minRows: 5, maxRows: 15 }}
                  trigger='blur'
                  stopValidateWithError
                  rules={[
                    {
                      validator: (rule, value) => verifyJSON(value),
                      message: t('不是合法的 JSON 字符串'),
                    },
                  ]}
                  extraText={
                    <div>
                      <p>{t('说明：')}</p>
                      <ul>
                        <li>
                          {t(
                            '使用 JSON 对象格式，格式为：{"组名": [最多请求次数, 最多请求完成次数]}',
                          )}
                        </li>
                        <li>
                          {t(
                            '示例：{"default": [200, 100], "vip": [0, 1000]}。',
                          )}
                        </li>
                        <li>
                          {t(
                            '[最多请求次数]必须大于等于0，[最多请求完成次数]必须大于等于1。',
                          )}
                        </li>
                        <li>
                          {t(
                            '[最多请求次数]和[最多请求完成次数]的最大值为2147483647。',
                          )}
                        </li>
                        <li>{t('分组速率配置优先级高于全局速率限制。')}</li>
                        <li>{t('限制周期统一使用上方配置的“限制周期”值。')}</li>
                      </ul>
                    </div>
                  }
                  onChange={(value) => {
                    setInputs({ ...inputs, ModelRequestRateLimitGroup: value });
                  }}
                />
              </Col>
            </Row>
            <Row>
              <Button size='default' onClick={onSubmit}>
                {t('保存模型速率限制')}
              </Button>
            </Row>
          </Form.Section>
        </Form>
      </Spin>
    </>
  );
}
