import React from 'react';
import { ShoppingBag } from 'lucide-react';

const FloatingStoreButton = () => {
  return (
    <a
      href='https://store.wzw.pp.ua'
      target='_blank'
      rel='noopener noreferrer'
      className='fixed top-24 right-4 z-40 flex items-center gap-2 px-4 py-2 bg-gradient-to-r from-purple-500 to-pink-500 text-white rounded-full shadow-lg hover:shadow-xl hover:scale-105 transition-all duration-300 animate-bounce-slow'
      style={{
        animation: 'float 3s ease-in-out infinite',
      }}
    >
      <ShoppingBag size={18} />
      <span className='font-semibold text-sm'>LDC商店</span>
      <style>{`
        @keyframes float {
          0%, 100% {
            transform: translateY(0px);
          }
          50% {
            transform: translateY(-8px);
          }
        }
      `}</style>
    </a>
  );
};

export default FloatingStoreButton;
