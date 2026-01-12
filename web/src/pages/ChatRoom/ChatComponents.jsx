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

import React, { useMemo } from 'react';
import { Avatar, Tag, Typography, Image } from '@douyinfe/semi-ui';
import { IconUser } from '@douyinfe/semi-icons';
import { useTranslation } from 'react-i18next';
import MarkdownRenderer from '../../components/common/markdown/MarkdownRenderer';

const { Text } = Typography;

const DEFAULT_AVATAR = '/avatar.png';

const imageUrlPattern = /\.(jpg|jpeg|png|gif|webp|bmp|svg)(\?.*)?$/i;

function isImageUrl(url) {
  if (!url || typeof url !== 'string') return false;
  try {
    const urlObj = new URL(url, window.location.origin);
    return imageUrlPattern.test(urlObj.pathname);
  } catch {
    return imageUrlPattern.test(url);
  }
}

function extractImageUrls(content) {
  if (!content) return [];
  const urlPattern = /https?:\/\/[^\s<>"{}|\\^`[\]]+/gi;
  const matches = content.match(urlPattern) || [];
  return matches.filter(isImageUrl);
}

function formatQuota(quota) {
  if (quota >= 1000000) {
    return (quota / 1000000).toFixed(1) + 'M';
  }
  if (quota >= 1000) {
    return (quota / 1000).toFixed(1) + 'K';
  }
  return quota.toString();
}

function formatTime(ts) {
  if (!ts) return '';
  const d = new Date(ts * 1000);
  const now = new Date();
  const isToday = d.toDateString() === now.toDateString();
  if (isToday) {
    return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  }
  return d.toLocaleDateString([], { month: 'short', day: 'numeric' }) + ' ' +
         d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

export const UserAvatar = ({ avatar, username, size = 'default', className = '' }) => {
  const avatarSrc = avatar && avatar !== DEFAULT_AVATAR ? avatar : null;

  return (
    <Avatar
      src={avatarSrc}
      size={size}
      color="blue"
      className={`flex-shrink-0 ${className}`}
      style={{ margin: 0 }}
    >
      {!avatarSrc && (username ? username[0].toUpperCase() : <IconUser />)}
    </Avatar>
  );
};

export const ImagePreview = ({ src, onClear }) => {
  if (!src) return null;

  return (
    <div className="relative inline-block mr-2 mb-2">
      <Image
        src={src}
        width={80}
        height={80}
        className="object-cover rounded-lg border border-semi-color-border"
        preview={false}
      />
      <button
        type="button"
        onClick={onClear}
        className="absolute -top-2 -right-2 bg-red-500 text-white rounded-full w-5 h-5 flex items-center justify-center cursor-pointer text-xs shadow hover:bg-red-600 transition-colors border-none p-0"
        aria-label="Remove image"
      >
        ×
      </button>
    </div>
  );
};

export const MessageImages = ({ imageUrls = [], contentImageUrls = [] }) => {
  const allImages = useMemo(() => {
    const set = new Set([...imageUrls, ...contentImageUrls]);
    return Array.from(set);
  }, [imageUrls, contentImageUrls]);

  if (allImages.length === 0) return null;

  return (
    <div className="flex flex-wrap gap-2 mt-2">
      {allImages.map((url) => (
        <Image
          key={url}
          src={url}
          width={200}
          className="rounded-lg max-w-full"
          style={{ maxHeight: 300 }}
        />
      ))}
    </div>
  );
};

export const UserInfoBadges = ({ quota, usedQuota, usageRank, balanceRank, compact = false }) => {
  const { t } = useTranslation();

  if (compact) {
    return (
      <div className="flex flex-wrap gap-1 mt-1">
        {usageRank > 0 && (
          <Tag size="small" color="red" className="text-xs">#{usageRank}</Tag>
        )}
        {balanceRank > 0 && (
          <Tag size="small" color="amber" className="text-xs">#{balanceRank}</Tag>
        )}
      </div>
    );
  }

  return (
    <div className="flex flex-wrap items-center gap-1 text-xs opacity-80">
      {quota !== undefined && quota > 0 && (
        <span className="text-green-600 dark:text-green-400">
          {t('额度')}: {formatQuota(quota)}
        </span>
      )}
      {usedQuota !== undefined && usedQuota > 0 && (
        <span className="text-orange-600 dark:text-orange-400 ml-1">
          {t('已用')}: {formatQuota(usedQuota)}
        </span>
      )}
      {usageRank > 0 && (
        <Tag size="small" color="red" className="ml-1">{t('消耗榜')} #{usageRank}</Tag>
      )}
      {balanceRank > 0 && (
        <Tag size="small" color="amber" className="ml-1">{t('囤囤鼠')} #{balanceRank}</Tag>
      )}
    </div>
  );
};

export const ChatBubble = ({ message, isSelf, showUserInfo = true }) => {
  const { t } = useTranslation();
  const {
    username,
    display_name,
    avatar,
    content,
    image_urls = [],
    created_at,
    quota,
    used_quota,
    usage_rank,
    balance_rank,
  } = message;

  const contentImageUrls = useMemo(() => extractImageUrls(content), [content]);
  const displayContent = useMemo(() => {
    if (!content) return '';
    let result = content;
    contentImageUrls.forEach(url => {
      result = result.replace(url, '');
    });
    return result.trim();
  }, [content, contentImageUrls]);

  const hasContent = displayContent.length > 0;
  const hasImages = (image_urls && image_urls.length > 0) || contentImageUrls.length > 0;

  return (
    <div className={`flex w-full mb-4 gap-2 ${isSelf ? 'flex-row-reverse' : 'flex-row'}`}>
      <UserAvatar
        avatar={avatar}
        username={username}
        size="small"
      />

      <div className={`flex flex-col max-w-[50%] ${isSelf ? 'items-end' : 'items-start'}`}>
        {showUserInfo && !isSelf && (
          <div className={`flex flex-col mb-1 ${isSelf ? 'items-end' : 'items-start'}`}>
            <div className="flex items-center gap-2">
              <Text strong size="small" className="text-sm">
                {display_name || username || t('匿名')}
              </Text>
              <Text type="tertiary" size="small" className="text-xs">
                {formatTime(created_at)}
              </Text>
            </div>
            <UserInfoBadges
              quota={quota}
              usedQuota={used_quota}
              usageRank={usage_rank}
              balanceRank={balance_rank}
            />
          </div>
        )}

        {isSelf && (
          <div className="flex items-center gap-2 mb-1">
            <Text type="tertiary" size="small" className="text-xs">
              {formatTime(created_at)}
            </Text>
          </div>
        )}

        <div
          className={`px-3 py-2 rounded-2xl shadow-sm overflow-hidden ${
            isSelf
              ? 'bg-blue-500 text-white rounded-tr-sm'
              : 'bg-white dark:bg-zinc-800 border border-semi-color-border rounded-tl-sm'
          }`}
        >
          {hasContent && (
            <div className={isSelf ? '[&_*]:!text-white [&_a]:!text-blue-200' : ''}>
              <MarkdownRenderer content={displayContent} />
            </div>
          )}
          {hasImages && (
            <MessageImages
              imageUrls={image_urls || []}
              contentImageUrls={contentImageUrls}
            />
          )}
          {!hasContent && !hasImages && (
            <Text type="tertiary" className={isSelf ? 'text-white/70' : ''}>
              {t('空消息')}
            </Text>
          )}
        </div>
      </div>
    </div>
  );
};

export default ChatBubble;
