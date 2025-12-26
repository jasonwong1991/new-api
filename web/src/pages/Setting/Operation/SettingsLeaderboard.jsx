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

import React, { useEffect, useState, useCallback } from 'react';
import {
  Button,
  Col,
  Form,
  Row,
  Spin,
  Select,
  Tag,
  TagGroup,
  Typography,
} from '@douyinfe/semi-ui';
import { API, showError, showSuccess, showWarning } from '../../../helpers';
import { useTranslation } from 'react-i18next';
import { debounce } from 'lodash';

export default function SettingsLeaderboard(props) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [searching, setSearching] = useState(false);
  const [hiddenUsers, setHiddenUsers] = useState([]);
  const [originalHiddenUsers, setOriginalHiddenUsers] = useState([]);
  const [searchResults, setSearchResults] = useState([]);

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
          const options = data.items.map((user) => ({
            value: user.username,
            label: `${user.username}${user.display_name ? ` (${user.display_name})` : ''}`,
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
    [],
  );

  function onSubmit() {
    const currentValue = hiddenUsers.join(',');
    const originalValue = originalHiddenUsers.join(',');
    if (currentValue === originalValue) {
      return showWarning(t('你似乎并没有修改什么'));
    }

    setLoading(true);
    API.put('/api/option/', {
      key: 'LeaderboardHiddenUsers',
      value: currentValue,
    })
      .then((res) => {
        const { success, message } = res.data;
        if (success) {
          showSuccess(t('保存成功'));
          setOriginalHiddenUsers([...hiddenUsers]);
          props.refresh();
        } else {
          showError(message || t('保存失败'));
        }
      })
      .catch(() => {
        showError(t('保存失败，请重试'));
      })
      .finally(() => {
        setLoading(false);
      });
  }

  const handleUserSelect = (value) => {
    if (value && !hiddenUsers.includes(value)) {
      setHiddenUsers([...hiddenUsers, value]);
    }
    setSearchResults([]);
  };

  const handleUserRemove = (username) => {
    setHiddenUsers(hiddenUsers.filter((u) => u !== username));
  };

  useEffect(() => {
    if (props.options?.LeaderboardHiddenUsers) {
      const users = props.options.LeaderboardHiddenUsers.split(',').filter(
        (u) => u.trim(),
      );
      setHiddenUsers(users);
      setOriginalHiddenUsers(users);
    } else {
      setHiddenUsers([]);
      setOriginalHiddenUsers([]);
    }
  }, [props.options?.LeaderboardHiddenUsers]);

  return (
    <Spin spinning={loading}>
      <Form style={{ marginBottom: 15 }}>
        <Form.Section
          text={t('榜单设置')}
          extraText={t('配置榜单中需要隐藏的用户')}
        >
          <Row gutter={16}>
            <Col xs={24} sm={24} md={16} lg={12} xl={12}>
              <Typography.Text strong style={{ display: 'block', marginBottom: 8 }}>
                {t('隐藏用户列表')}
              </Typography.Text>
              <Typography.Text type="tertiary" size="small" style={{ display: 'block', marginBottom: 12 }}>
                {t('在榜单中隐藏的用户不会出现在用户榜单和模型榜单中')}
              </Typography.Text>
              <Select
                style={{ width: '100%', marginBottom: 12 }}
                placeholder={t('输入用户名或ID搜索用户')}
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
              {hiddenUsers.length > 0 && (
                <TagGroup
                  maxTagCount={20}
                  style={{
                    display: 'flex',
                    flexWrap: 'wrap',
                    gap: 8,
                  }}
                  tagList={hiddenUsers.map((username) => ({
                    tagKey: username,
                    children: username,
                    closable: true,
                    onClose: () => handleUserRemove(username),
                  }))}
                />
              )}
              {hiddenUsers.length === 0 && (
                <Typography.Text type="tertiary">
                  {t('暂无隐藏用户')}
                </Typography.Text>
              )}
            </Col>
          </Row>
          <Row style={{ marginTop: 16 }}>
            <Button size="default" onClick={onSubmit}>
              {t('保存榜单设置')}
            </Button>
          </Row>
        </Form.Section>
      </Form>
    </Spin>
  );
}
