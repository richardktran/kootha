import React, { createContext, useContext, useEffect, useRef, useState } from 'react';
import { Room, WebSocketEvent } from '@/types/quiz';

interface WebSocketContextType {
  connect: () => void;
  disconnect: () => void;
  sendMessage: (event: WebSocketEvent) => void;
  isConnected: boolean;
  currentRoom: Room | null;
}

const WebSocketContext = createContext<WebSocketContextType | null>(null);

export const useWebSocket = () => {
  const context = useContext(WebSocketContext);
  if (!context) {
    throw new Error('useWebSocket must be used within a WebSocketProvider');
  }
  return context;
};

export const WebSocketProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [isConnected, setIsConnected] = useState(false);
  const [currentRoom, setCurrentRoom] = useState<Room | null>(null);
  const ws = useRef<WebSocket | null>(null);

  const connect = () => {
    // In production, this would be your actual WebSocket server URL
    ws.current = new WebSocket('ws://localhost:8080');

    ws.current.onopen = () => {
      setIsConnected(true);
      console.log('Connected to WebSocket');
    };

    ws.current.onclose = () => {
      setIsConnected(false);
      console.log('Disconnected from WebSocket');
    };

    ws.current.onmessage = (event) => {
      const data: WebSocketEvent = JSON.parse(event.data);
      handleWebSocketMessage(data);
    };
  };

  const disconnect = () => {
    ws.current?.close();
  };

  const sendMessage = (event: WebSocketEvent) => {
    if (ws.current?.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify(event));
    }
  };

  const handleWebSocketMessage = (event: WebSocketEvent) => {
    switch (event.type) {
      case 'ROOM_JOINED':
        setCurrentRoom(event.payload);
        break;
      case 'LEADERBOARD_UPDATE':
        if (currentRoom) {
          setCurrentRoom({
            ...currentRoom,
            participants: event.payload.participants,
          });
        }
        break;
      case 'QUIZ_ENDED':
        if (currentRoom) {
          setCurrentRoom({
            ...currentRoom,
            status: 'finished',
            participants: event.payload.participants,
          });
        }
        break;
      // Add more cases as needed
    }
  };

  useEffect(() => {
    return () => {
      disconnect();
    };
  }, []);

  return (
    <WebSocketContext.Provider value={{ connect, disconnect, sendMessage, isConnected, currentRoom }}>
      {children}
    </WebSocketContext.Provider>
  );
}; 