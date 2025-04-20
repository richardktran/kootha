import React, {
  createContext,
  useContext,
  useEffect,
  useRef,
  useState,
} from "react";

import { Room, WebSocketEvent, Question } from "@/types/quiz";

interface WebSocketContextType {
  connect: () => void;
  disconnect: () => void;
  sendMessage: (event: WebSocketEvent) => void;
  isConnected: boolean;
  currentRoom: Room | null;
  setCurrentRoom: (room: Room | null) => void;
  currentQuestion: Question | null;
}

const WebSocketContext = createContext<WebSocketContextType | null>(null);

export const useWebSocket = () => {
  const context = useContext(WebSocketContext);

  if (!context) {
    throw new Error("useWebSocket must be used within a WebSocketProvider");
  }

  return context;
};

export const WebSocketProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [isConnected, setIsConnected] = useState(false);
  const [currentRoom, setCurrentRoom] = useState<Room | null>(null);
  const [currentQuestion, setCurrentQuestion] = useState<Question | null>(null);
  const ws = useRef<WebSocket | null>(null);

  const connect = () => {
    return new Promise<void>((resolve, reject) => {
      console.log("Attempting to connect to WebSocket...");
      ws.current = new WebSocket("ws://localhost:8080/ws");

      ws.current.onopen = () => {
        setIsConnected(true);
        console.log("Connected to WebSocket");
        resolve();
      };

      ws.current.onerror = (error) => {
        console.error("WebSocket error:", error);
        reject(error);
      };

      ws.current.onclose = () => {
        setIsConnected(false);
        console.log("Disconnected from WebSocket");
      };

      ws.current.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data);
          console.log("Received message:", message);

          switch (message.type) {
            case "room_joined":
              setCurrentRoom(message.payload.data);
              break;
            case "participant_joined":
              if (currentRoom) {
                setCurrentRoom({
                  ...currentRoom,
                  participants: [...currentRoom.participants, message.payload.data.participant],
                });
              }
              break;
            case "participant_left":
              if (currentRoom) {
                setCurrentRoom({
                  ...currentRoom,
                  participants: currentRoom.participants.filter(
                    (p) => p.id !== message.payload.data.participantId
                  ),
                });
              }
              break;
            case "question_started":
              setCurrentQuestion(message.payload.data.question);
              break;
            case "leaderboard_update":
              if (currentRoom) {
                setCurrentRoom({
                  ...currentRoom,
                  participants: message.payload.data.participants,
                });
              }
              break;
            case "quiz_ended":
              if (currentRoom) {
                setCurrentRoom({
                  ...currentRoom,
                  status: "finished",
                  participants: message.payload.data.participants,
                });
              }
              break;
            case "error":
              console.error("Server error:", message.payload.message);
              break;
          }
        } catch (error) {
          console.error("Error parsing message:", error);
        }
      };
    });
  };

  const disconnect = () => {
    if (ws.current) {
      ws.current.close();
    }
  };

  const sendMessage = (event: WebSocketEvent) => {
    if (ws.current?.readyState === WebSocket.OPEN) {
      console.log("Sending message: ", event.type);
      ws.current.send(JSON.stringify(event));
    }
  };

  useEffect(() => {
    return () => {
      disconnect();
    };
  }, []);

  return (
    <WebSocketContext.Provider
      value={{
        connect,
        disconnect,
        sendMessage,
        isConnected,
        currentRoom,
        setCurrentRoom,
        currentQuestion,
      }}
    >
      {children}
    </WebSocketContext.Provider>
  );
};
