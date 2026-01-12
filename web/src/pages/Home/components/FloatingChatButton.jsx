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

import React, { useContext, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { MessageCircle } from 'lucide-react';
import { StatusContext } from '../../../context/Status';
import { API } from '../../../helpers';

const FloatingChatButton = () => {
  const navigate = useNavigate();
  const [statusState] = useContext(StatusContext);
  const [messageCount, setMessageCount] = useState(0);
  const [isHovered, setIsHovered] = useState(false);

  const chatRoomEnabled = statusState?.status?.chat_room_enabled !== false;

  useEffect(() => {
    const fetchMessageCount = async () => {
      try {
        const res = await API.get('/api/chat/count?room=global');
        if (res.data.success && res.data.data) {
          setMessageCount(res.data.data.count || 0);
        }
      } catch (error) {
        // Silently fail
      }
    };

    fetchMessageCount();
    const interval = setInterval(fetchMessageCount, 30000);
    return () => clearInterval(interval);
  }, []);

  if (!chatRoomEnabled) return null;

  const displayCount = messageCount > 999 ? '999+' : messageCount.toString();

  return (
    <button
      onClick={() => navigate('/chat-room')}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
      className="fixed bottom-6 right-6 z-40 flex items-center justify-center w-14 h-14 bg-gradient-to-r from-blue-500 to-cyan-500 text-white rounded-full shadow-lg hover:shadow-2xl transition-all duration-300 group"
      style={{
        animation: 'pulse-glow 2s ease-in-out infinite',
        transform: isHovered ? 'scale(1.1)' : 'scale(1)',
      }}
      aria-label="进入聊天室"
    >
      <MessageCircle
        size={24}
        className="transition-transform duration-300 group-hover:rotate-12"
      />

      {messageCount > 0 && (
        <span
          className="absolute -top-1 -right-1 min-w-[20px] h-5 px-1.5 flex items-center justify-center bg-red-500 text-white text-xs font-bold rounded-full shadow-md"
          style={{
            animation: 'badge-bounce 1s ease-in-out infinite',
          }}
        >
          {displayCount}
        </span>
      )}

      <span
        className={`absolute right-full mr-3 px-3 py-1.5 bg-gray-800 text-white text-sm rounded-lg whitespace-nowrap shadow-lg transition-all duration-300 ${
          isHovered ? 'opacity-100 translate-x-0' : 'opacity-0 translate-x-2 pointer-events-none'
        }`}
      >
        聊天室
        <span className="absolute right-0 top-1/2 -translate-y-1/2 translate-x-full border-8 border-transparent border-l-gray-800" />
      </span>

      <style>{`
        @keyframes pulse-glow {
          0%, 100% {
            box-shadow: 0 4px 15px rgba(59, 130, 246, 0.4);
          }
          50% {
            box-shadow: 0 4px 25px rgba(59, 130, 246, 0.6);
          }
        }
        @keyframes badge-bounce {
          0%, 100% {
            transform: scale(1);
          }
          50% {
            transform: scale(1.1);
          }
        }
      `}</style>
    </button>
  );
};

export default FloatingChatButton;
