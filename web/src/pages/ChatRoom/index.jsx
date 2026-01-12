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

import React, { useCallback, useContext, useEffect, useMemo, useRef, useState } from 'react';
import { Button, Card, Tag, Typography, Toast, TextArea, Spin } from '@douyinfe/semi-ui';
import { IconImage, IconSend, IconRefresh } from '@douyinfe/semi-icons';
import { useTranslation } from 'react-i18next';
import { StatusContext } from '../../context/Status';
import { UserContext } from '../../context/User';
import { useChatRoomSocket } from '../../hooks/chatRoom/useChatRoomSocket';
import { ChatBubble, ImagePreview, UserAvatar } from './ChatComponents';
import { API } from '../../helpers';

const { Title, Text } = Typography;

const CHAT_ROOM_HEIGHT = 'calc(100vh - 120px)';

const ChatRoomPage = () => {
  const { t } = useTranslation();
  const [statusState] = useContext(StatusContext);
  const [userState] = useContext(UserContext);

  const me = userState?.user;
  const isLoggedIn = !!me?.username;
  const chatRoomEnabled = statusState?.status?.chat_room_enabled !== false;
  const enabled = chatRoomEnabled && isLoggedIn;
  const messageLimit = Number(
    statusState?.status?.chat_room_message_limit || 1000,
  );
  const maxMessageLength = Number(
    statusState?.status?.chat_room_max_message_length || 8000,
  );

  const { messages, connectionState, lastError, announcement, sendMessage, reconnect } =
    useChatRoomSocket({
      enabled,
      messageLimit,
      room: 'global',
    });

  const [draft, setDraft] = useState('');
  const [pendingImages, setPendingImages] = useState([]);
  const [uploading, setUploading] = useState(false);
  const listRef = useRef(null);
  const bottomRef = useRef(null);
  const fileInputRef = useRef(null);
  const pendingImagesRef = useRef(pendingImages);
  const [autoScroll, setAutoScroll] = useState(true);

  const canSend = useMemo(() => {
    const content = draft.trim();
    const hasContent = content.length > 0;
    const hasImages = pendingImages.length > 0;
    return (
      enabled &&
      connectionState === 'connected' &&
      (hasContent || hasImages) &&
      content.length <= maxMessageLength &&
      !uploading
    );
  }, [draft, pendingImages, enabled, connectionState, maxMessageLength, uploading]);

  useEffect(() => {
    pendingImagesRef.current = pendingImages;
  }, [pendingImages]);

  useEffect(() => {
    return () => {
      pendingImagesRef.current.forEach((img) => {
        if (img.preview) URL.revokeObjectURL(img.preview);
      });
    };
  }, []);

  useEffect(() => {
    if (lastError) {
      Toast.error(lastError);
    }
  }, [lastError]);

  useEffect(() => {
    if (autoScroll && bottomRef.current) {
      bottomRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [messages.length, autoScroll]);

  const handleScroll = useCallback(() => {
    const el = listRef.current;
    if (!el) return;
    const isNearBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 100;
    if (isNearBottom !== autoScroll) {
      setAutoScroll(isNearBottom);
    }
  }, [autoScroll]);

  const uploadImage = async (file) => {
    const formData = new FormData();
    formData.append('file', file);

    try {
      const res = await API.post('/api/chat/images', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      });
      if (res.data.success) {
        return res.data.data;
      } else {
        Toast.error(res.data.message || t('上传失败'));
        return null;
      }
    } catch (err) {
      Toast.error(err.message || t('上传失败'));
      return null;
    }
  };

  const processImageFile = async (file) => {
    if (!file.type.startsWith('image/')) {
      Toast.error(t('请选择图片文件'));
      return;
    }

    if (file.size > 10 * 1024 * 1024) {
      Toast.error(t('图片大小不能超过 10MB'));
      return;
    }

    setUploading(true);
    const result = await uploadImage(file);
    setUploading(false);

    if (result) {
      setPendingImages(prev => [...prev, {
        url: result.url,
        preview: URL.createObjectURL(file),
      }]);
    }
  };

  const handlePaste = useCallback((e) => {
    const items = e.clipboardData?.items;
    if (!items) return;

    for (const item of items) {
      if (item.type.startsWith('image/')) {
        e.preventDefault();
        const file = item.getAsFile();
        if (file) {
          processImageFile(file);
        }
        break;
      }
    }
  }, []);

  const handleFileSelect = (e) => {
    const files = e.target.files;
    if (files && files.length > 0) {
      Array.from(files).forEach(file => processImageFile(file));
    }
    e.target.value = '';
  };

  const removeImage = (index) => {
    setPendingImages(prev => {
      const newImages = [...prev];
      if (newImages[index]?.preview) {
        URL.revokeObjectURL(newImages[index].preview);
      }
      newImages.splice(index, 1);
      return newImages;
    });
  };

  const onSend = useCallback(() => {
    const content = draft.trim();
    const imageUrls = pendingImages.map(img => img.url);
    const hasContent = content.length > 0;
    const hasImages = imageUrls.length > 0;

    if (!hasContent && !hasImages) return;

    if (content.length > maxMessageLength) {
      Toast.error(t('消息过长'));
      return;
    }
    if (connectionState !== 'connected') {
      Toast.error(t('连接已断开'));
      return;
    }

    sendMessage(content, imageUrls);
    setDraft('');
    pendingImages.forEach(img => {
      if (img.preview) URL.revokeObjectURL(img.preview);
    });
    setPendingImages([]);
    setAutoScroll(true);
  }, [draft, pendingImages, maxMessageLength, connectionState, sendMessage, t]);

  const scrollToBottom = () => {
    if (bottomRef.current) {
      bottomRef.current.scrollIntoView({ behavior: 'smooth' });
      setAutoScroll(true);
    }
  };

  if (!chatRoomEnabled) {
    return (
      <div className="w-full max-w-4xl mx-auto px-4 py-8 mt-16">
        <Card>
          <Title heading={4}>{t('聊天室')}</Title>
          <Text type="tertiary">{t('聊天室已关闭')}</Text>
        </Card>
      </div>
    );
  }

  if (!isLoggedIn) {
    return (
      <div className="w-full max-w-4xl mx-auto px-4 py-8 mt-16">
        <Card>
          <Title heading={4}>{t('聊天室')}</Title>
          <Text type="tertiary">{t('请先登录后再进入聊天室')}</Text>
        </Card>
      </div>
    );
  }

  return (
    <div className="w-full max-w-4xl mx-auto px-4 pt-20 pb-4">
      <Card
        className="flex flex-col overflow-hidden"
        style={{ height: CHAT_ROOM_HEIGHT }}
        bodyStyle={{ padding: 0, display: 'flex', flexDirection: 'column', height: '100%' }}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3 border-b border-semi-color-border bg-semi-color-bg-1">
          <div className="flex items-center gap-3">
            <Title heading={5} style={{ margin: 0 }}>
              {t('聊天室')}
            </Title>
            <Tag
              color={connectionState === 'connected' ? 'green' : 'orange'}
              size="small"
            >
              {connectionState === 'connected' ? t('在线') : t('连接中...')}
            </Tag>
          </div>
          <div className="flex items-center gap-2">
            {me && (
              <div className="flex items-center gap-2">
                <UserAvatar
                  avatar={me.linux_do_avatar}
                  username={me.username}
                  size="extra-small"
                />
                <Text type="tertiary" size="small">
                  {me.display_name || me.username}
                </Text>
              </div>
            )}
            <Button
              icon={<IconRefresh />}
              size="small"
              theme="borderless"
              onClick={() => reconnect()}
              disabled={connectionState === 'connected'}
            />
          </div>
        </div>

        {/* Announcement */}
        {announcement && (
          <div className="px-4 py-2 border-b border-semi-color-border bg-amber-50 dark:bg-amber-900/20">
            <div className="flex items-start gap-2">
              <Text type="warning" strong size="small">📢</Text>
              <Text size="small" className="flex-1">{announcement}</Text>
            </div>
          </div>
        )}

        {/* Messages Area */}
        <div
          ref={listRef}
          onScroll={handleScroll}
          className="flex-1 overflow-y-auto px-4 py-4 bg-gray-50 dark:bg-zinc-900/50"
          style={{ minHeight: 0 }}
        >
          {messages.length === 0 ? (
            <div className="flex items-center justify-center h-full">
              <Text type="tertiary">{t('暂无消息，发送第一条消息吧')}</Text>
            </div>
          ) : (
            <div className="flex flex-col">
              {messages.map((m) => (
                <ChatBubble
                  key={m.id}
                  message={m}
                  isSelf={me?.username && m.username === me.username}
                />
              ))}
              <div ref={bottomRef} />
            </div>
          )}
        </div>

        {/* Scroll to bottom button */}
        {!autoScroll && messages.length > 0 && (
          <div className="absolute bottom-32 right-8">
            <Button
              theme="solid"
              size="small"
              onClick={scrollToBottom}
              className="shadow-lg"
            >
              {t('回到底部')}
            </Button>
          </div>
        )}

        {/* Image Preview Area */}
        {pendingImages.length > 0 && (
          <div className="px-4 py-2 border-t border-semi-color-border bg-semi-color-bg-1">
            <div className="flex flex-wrap gap-2">
              {pendingImages.map((img, idx) => (
                <ImagePreview
                  key={idx}
                  src={img.preview}
                  onClear={() => removeImage(idx)}
                />
              ))}
            </div>
          </div>
        )}

        {/* Input Area */}
        <div className="px-4 py-3 border-t border-semi-color-border bg-semi-color-bg-1">
          <div className="flex gap-2 items-end">
            <input
              type="file"
              accept="image/*"
              multiple
              ref={fileInputRef}
              className="hidden"
              onChange={handleFileSelect}
            />
            <Button
              icon={uploading ? <Spin size="small" /> : <IconImage />}
              theme="light"
              onClick={() => fileInputRef.current?.click()}
              disabled={uploading}
              aria-label={t('上传图片')}
            />
            <TextArea
              value={draft}
              onChange={(v) => setDraft(v)}
              onPaste={handlePaste}
              placeholder={t('输入消息，Ctrl+V 粘贴图片')}
              aria-label={t('输入消息')}
              autosize={{ minRows: 1, maxRows: 4 }}
              className="flex-1"
              onKeyDown={(e) => {
                if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
                  e.preventDefault();
                  onSend();
                }
              }}
            />
            <Button
              icon={<IconSend />}
              theme="solid"
              type="primary"
              onClick={onSend}
              disabled={!canSend}
            >
              {t('发送')}
            </Button>
          </div>
          <div className="mt-2 flex items-center justify-between">
            <Text type="tertiary" size="small">
              {t('支持 Markdown、图片粘贴/上传')}
            </Text>
            <Text type="tertiary" size="small">
              Ctrl/⌘ + Enter {t('发送')}
            </Text>
          </div>
        </div>
      </Card>
    </div>
  );
};

export default ChatRoomPage;
