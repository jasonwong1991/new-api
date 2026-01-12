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

import { useCallback, useEffect, useRef, useState } from 'react';

function toWsBaseUrl(httpBaseUrl) {
  return httpBaseUrl.replace(/^http/i, 'ws');
}

function getWsUrl(path) {
  const httpBase =
    import.meta.env.VITE_REACT_APP_SERVER_URL || window.location.origin;
  return `${toWsBaseUrl(httpBase)}${path}`;
}

function clampMessages(list, limit) {
  if (!Array.isArray(list)) return [];
  if (!limit || limit <= 0) return list;
  if (list.length <= limit) return list;
  return list.slice(list.length - limit);
}

export function useChatRoomSocket({ enabled, messageLimit, room = 'global' }) {
  const wsRef = useRef(null);
  const reconnectTimerRef = useRef(null);
  const reconnectAttemptRef = useRef(0);

  const [messages, setMessages] = useState([]);
  const [connectionState, setConnectionState] = useState('disconnected');
  const [lastError, setLastError] = useState('');

  const cleanup = useCallback(() => {
    if (reconnectTimerRef.current) {
      clearTimeout(reconnectTimerRef.current);
      reconnectTimerRef.current = null;
    }
    if (wsRef.current) {
      // Prevent event handlers from firing during cleanup
      wsRef.current.onclose = null;
      wsRef.current.onerror = null;
      wsRef.current.onmessage = null;
      wsRef.current.onopen = null;
      try {
        wsRef.current.close();
      } catch (e) {
        // ignore
      }
      wsRef.current = null;
    }
  }, []);

  const scheduleReconnect = useCallback(() => {
    if (!enabled) return;
    if (reconnectTimerRef.current) return;
    const attempt = reconnectAttemptRef.current;
    const delay = Math.min(1000 * Math.pow(1.6, attempt), 10000);
    reconnectTimerRef.current = setTimeout(() => {
      reconnectTimerRef.current = null;
      reconnectAttemptRef.current += 1;
      connect();
    }, delay);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [enabled]);

  const connect = useCallback(() => {
    if (!enabled) return;
    cleanup();
    setLastError('');
    setConnectionState('connecting');

    const wsUrl = getWsUrl(`/api/chat/ws?room=${encodeURIComponent(room)}`);
    const ws = new WebSocket(wsUrl, ['chat']);
    wsRef.current = ws;

    ws.onopen = () => {
      reconnectAttemptRef.current = 0;
      setConnectionState('connected');
    };

    ws.onmessage = (evt) => {
      try {
        const payload = JSON.parse(evt.data);
        if (!payload || typeof payload.type !== 'string') return;
        if (payload.type === 'init') {
          const initMessages = payload?.data?.messages || [];
          setMessages(clampMessages(initMessages, messageLimit));
          return;
        }
        if (payload.type === 'message') {
          const m = payload?.data?.message;
          if (!m) return;
          setMessages((prev) => clampMessages([...prev, m], messageLimit));
          return;
        }
        if (payload.type === 'error') {
          const msg = payload?.data?.message || 'error';
          setLastError(String(msg));
        }
      } catch (e) {
        // ignore
      }
    };

    ws.onerror = () => {
      setLastError('WebSocket error');
    };

    ws.onclose = () => {
      // Only schedule reconnect if the ref still points to this ws instance
      if (wsRef.current === ws) {
        setConnectionState('disconnected');
        scheduleReconnect();
      }
    };
  }, [cleanup, enabled, messageLimit, room, scheduleReconnect]);

  const reconnect = useCallback(() => {
    reconnectAttemptRef.current = 0;
    connect();
  }, [connect]);

  const sendMessage = useCallback(
    (content, imageUrls = []) => {
      const ws = wsRef.current;
      if (!ws || ws.readyState !== WebSocket.OPEN) return false;
      const payload = {
        type: 'send',
        data: { content, room, image_urls: imageUrls },
      };
      ws.send(JSON.stringify(payload));
      return true;
    },
    [room],
  );

  useEffect(() => {
    if (!enabled) {
      cleanup();
      setConnectionState('disconnected');
      return;
    }
    connect();
    return () => cleanup();
  }, [connect, cleanup, enabled]);

  return {
    messages,
    connectionState,
    lastError,
    sendMessage,
    reconnect,
  };
}
