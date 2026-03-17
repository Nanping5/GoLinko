import { useEffect, useRef, useState, useCallback } from 'react';
import { useAuthStore } from '../store/useAuthStore';
import type { ChatMessage } from '../types';

interface UseWebSocketOptions {
  onMessage?: (msg: ChatMessage) => void;
  onOpen?: () => void;
  onClose?: () => void;
  autoReconnect?: boolean;
  reconnectDelaysMs?: number[]; // e.g., [1000, 2000, 5000]
}

export const useWebSocket = (url: string, options: UseWebSocketOptions = {}) => {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [status, setStatus] = useState<'connecting' | 'open' | 'closed'>('closed');
  const ws = useRef<WebSocket | null>(null);
  const connectionSeqRef = useRef(0);
  const token = useAuthStore((state) => state.token);
  const reconnectTimer = useRef<number | null>(null);
  const attemptRef = useRef(0);
  const reconnectDelays = options.reconnectDelaysMs || [1000, 2000, 5000, 10000];
  const callbacksRef = useRef<Pick<UseWebSocketOptions, 'onMessage' | 'onOpen' | 'onClose'>>({
    onMessage: options.onMessage,
    onOpen: options.onOpen,
    onClose: options.onClose,
  });

  // 始终保持回调为最新，避免闭包读取到旧的 activeChat
  useEffect(() => {
    callbacksRef.current = {
      onMessage: options.onMessage,
      onOpen: options.onOpen,
      onClose: options.onClose,
    };
  }, [options.onMessage, options.onOpen, options.onClose]);

  const clearTimer = () => {
    if (reconnectTimer.current) {
      window.clearTimeout(reconnectTimer.current);
      reconnectTimer.current = null;
    }
  };

  useEffect(() => {
    if (!token || !url) return;

    const connect = () => {
      clearTimer();
      const seq = ++connectionSeqRef.current;
      const fullUrl = url.includes('token=') ? url : `${url}?token=${token}`;
      const socket = new WebSocket(fullUrl);
      ws.current = socket;
      setStatus('connecting');

      socket.onopen = () => {
        if (seq !== connectionSeqRef.current) return;
        attemptRef.current = 0;
        setStatus('open');
        callbacksRef.current.onOpen?.();
      };

      socket.onmessage = (event) => {
        if (seq !== connectionSeqRef.current) return;
        try {
          const data = JSON.parse(event.data) as ChatMessage;
          // 兼容后端字段
          const normalized: ChatMessage = {
            sessionId: data.sessionId || (data as any).session_id,
            sendId: data.sendId || (data as any).send_id,
            sendName: data.sendName || (data as any).send_name,
            sendAvatar: data.sendAvatar || (data as any).send_avatar,
            receiveId: data.receiveId || (data as any).receive_id,
            content: data.content,
            url: data.url,
            type: typeof data.type === 'number' ? data.type : (data as any).type ?? 0,
            fileType: data.fileType || (data as any).file_type,
            fileName: data.fileName || (data as any).file_name,
            fileSize: data.fileSize || (data as any).file_size,
            avData: data.avData || (data as any).av_data,
            createdAt: data.createdAt || (data as any).created_at,
          };
          setMessages((prev) => [...prev, normalized]);
          callbacksRef.current.onMessage?.(normalized);
        } catch (e) {
          const raw = event.data;
          // 后端首条欢迎语是纯文本，例如 “登录成功,欢迎来到GoLinko”，忽略即可
          if (typeof raw === 'string' && raw.includes('登录成功')) {
            return;
          }
          console.error('WS Message Parse Error:', e);
        }
      };

      socket.onclose = (event) => {
        // 忽略旧连接回调，避免开发模式 StrictMode 下旧连接把新连接状态覆盖成 closed
        if (seq !== connectionSeqRef.current) return;
        setStatus('closed');
        callbacksRef.current.onClose?.();
        // 4001 = 被新设备挤下线，不自动重连，避免两端互踢死循环
        if (event.code === 4001) {
          console.warn('WS: kicked by another device, stopping reconnect');
          return;
        }
        if (options.autoReconnect !== false) {
          const delay = reconnectDelays[Math.min(attemptRef.current, reconnectDelays.length - 1)];
          attemptRef.current += 1;
          reconnectTimer.current = window.setTimeout(connect, delay);
        }
      };
    };

    connect();

    return () => {
      clearTimer();
      const current = ws.current;
      // 标记当前连接为过期，后续旧连接 onclose/onmessage 不再影响状态
      connectionSeqRef.current += 1;
      if (!current) return;
      // 避免在 CONNECTING 状态直接 close 触发 "WebSocket is closed before the connection is established" 噪音
      if (current.readyState === WebSocket.OPEN || current.readyState === WebSocket.CLOSING) {
        current.close();
      }
    };
  }, [url, token]);

  const sendMessage = useCallback((msg: Partial<ChatMessage>) => {
    if (ws.current?.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify(msg));
    } else {
      console.error('WS not open. Current state:', ws.current?.readyState);
    }
  }, []);

  return { messages, status, sendMessage };
};
