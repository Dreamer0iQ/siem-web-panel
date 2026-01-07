import { useEffect, useRef, useCallback } from 'react';

export type WSMessage = {
    type: 'event' | 'stats' | 'alert';
    data: any;
    timestamp: string;
};

export type WSEventHandler = (message: WSMessage) => void;

export const useWebSocket = (url: string, onMessage: WSEventHandler) => {
    const wsRef = useRef<WebSocket | null>(null);
    const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
    const reconnectAttemptsRef = useRef(0);
    const maxReconnectAttempts = 10;
    const baseReconnectDelay = 1000;

    const connect = useCallback(() => {
        try {
            const ws = new WebSocket(url);

            ws.onopen = () => {
                console.log('WebSocket connected');
                reconnectAttemptsRef.current = 0;
            };

            ws.onmessage = (event) => {
                try {
                    const message: WSMessage = JSON.parse(event.data);
                    onMessage(message);
                } catch (error) {
                    console.error('Failed to parse WebSocket message:', error);
                }
            };

            ws.onerror = (error) => {
                console.error('WebSocket error:', error);
            };

            ws.onclose = () => {
                console.log('WebSocket disconnected');
                wsRef.current = null;

                // Exponential backoff для переподключения
                if (reconnectAttemptsRef.current < maxReconnectAttempts) {
                    const delay = Math.min(
                        baseReconnectDelay * Math.pow(2, reconnectAttemptsRef.current),
                        30000
                    );

                    console.log(`Reconnecting in ${delay}ms (attempt ${reconnectAttemptsRef.current + 1}/${maxReconnectAttempts})`);

                    reconnectTimeoutRef.current = setTimeout(() => {
                        reconnectAttemptsRef.current++;
                        connect();
                    }, delay);
                } else {
                    console.error('Max reconnection attempts reached');
                }
            };

            wsRef.current = ws;
        } catch (error) {
            console.error('Failed to create WebSocket connection:', error);
        }
    }, [url, onMessage]);

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
    }, [connect]);

    const sendMessage = useCallback((message: any) => {
        if (wsRef.current?.readyState === WebSocket.OPEN) {
            wsRef.current.send(JSON.stringify(message));
        } else {
            console.warn('WebSocket is not connected');
        }
    }, []);

    return {
        sendMessage,
        isConnected: wsRef.current?.readyState === WebSocket.OPEN,
    };
};
