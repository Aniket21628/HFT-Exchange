import { useEffect, useRef, useState } from 'react';
import type { WSMessage } from '../types';

// Derive WebSocket URL from API URL (http -> ws, https -> wss)
const getWsUrl = () => {
  const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';
  const wsUrl = apiUrl.replace(/^http/, 'ws') + '/ws';
  return wsUrl;
};

const WS_URL = getWsUrl();

export function useWebSocket(onMessage: (message: WSMessage) => void) {
  const [isConnected, setIsConnected] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<number | undefined>(undefined);

  const connect = () => {
    try {
      const ws = new WebSocket(WS_URL);

      ws.onopen = () => {
        console.log('WebSocket connected');
        setIsConnected(true);
      };

      ws.onmessage = (event) => {
        try {
          // Handle multiple JSON messages separated by newlines
          const messages = event.data.toString().trim().split('\n');
          messages.forEach((msgStr: string) => {
            if (msgStr.trim()) {
              try {
                const message: WSMessage = JSON.parse(msgStr);
                onMessage(message);
              } catch (e) {
                console.error('Failed to parse message:', msgStr, e);
              }
            }
          });
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error);
        }
      };

      ws.onerror = (error) => {
        console.error('WebSocket error:', error);
      };

      ws.onclose = () => {
        console.log('WebSocket disconnected');
        setIsConnected(false);
        
        // Attempt to reconnect after 3 seconds
        reconnectTimeoutRef.current = window.setTimeout(() => {
          console.log('Attempting to reconnect...');
          connect();
        }, 3000);
      };

      wsRef.current = ws;
    } catch (error) {
      console.error('Failed to create WebSocket connection:', error);
    }
  };

  useEffect(() => {
    connect();

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return { isConnected };
}
