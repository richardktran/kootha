import React, {
  createContext,
  useContext,
  useEffect,
  useRef,
  useState,
} from "react";

import { Room, WebSocketEvent, Question } from "@/types/quiz";

const WS_URL =
  process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8086/ws";

export interface QuestionResult {
  questionId: string;
  questionIndex: number;
  correctAnswer: number;
  reason: "all_submitted" | "timeout" | string;
}

interface WebSocketContextType {
  connect: () => Promise<void>;
  disconnect: () => void;
  sendMessage: (event: WebSocketEvent) => void;
  isConnected: boolean;
  currentRoom: Room | null;
  setCurrentRoom: React.Dispatch<React.SetStateAction<Room | null>>;
  currentQuestion: Question | null;
  questionResult: QuestionResult | null;
  setQuestionResult: React.Dispatch<React.SetStateAction<QuestionResult | null>>;
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
  const [questionResult, setQuestionResult] = useState<QuestionResult | null>(
    null,
  );
  const ws = useRef<WebSocket | null>(null);

  const connect = () => {
    return new Promise<void>((resolve, reject) => {
      if (ws.current?.readyState === WebSocket.OPEN) {
        resolve();
        return;
      }

      console.log("Attempting to connect to WebSocket...", WS_URL);
      ws.current = new WebSocket(WS_URL);

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
            case "room_joined": {
              const room = message.payload.data;

              setCurrentRoom({
                ...room,
                participants: room.participants ?? [],
                questions: room.questions ?? [],
              });
              break;
            }
            case "participant_joined":
              setCurrentRoom((prev) => {
                if (!prev) return prev;

                const participants = prev.participants ?? [];
                const incoming = message.payload.data.participant;

                if (participants.some((p: { id: string }) => p.id === incoming.id)) {
                  return prev;
                }

                return {
                  ...prev,
                  participants: [...participants, incoming],
                };
              });
              break;
            case "participant_left":
              setCurrentRoom((prev) => {
                if (!prev) return prev;

                return {
                  ...prev,
                  participants: (prev.participants ?? []).filter(
                    (p) => p.id !== message.payload.data.participantId,
                  ),
                };
              });
              break;
            case "question_started":
              setQuestionResult(null);
              setCurrentQuestion(message.payload.data.question);
              setCurrentRoom((prev) => {
                if (!prev) return prev;
                return {
                  ...prev,
                  status: "in-progress",
                  currentQuestionIndex:
                    message.payload.data.questionIndex ??
                    prev.currentQuestionIndex,
                };
              });
              break;
            case "question_result": {
              const data = message.payload.data;
              setQuestionResult({
                questionId: data.questionId,
                questionIndex: data.questionIndex,
                correctAnswer: data.correctAnswer,
                reason: data.reason,
              });
              setCurrentQuestion((prev) =>
                prev
                  ? { ...prev, correctAnswer: data.correctAnswer }
                  : prev,
              );
              break;
            }
            case "leaderboard_update":
              setCurrentRoom((prev) => {
                if (!prev) return prev;

                const raw = message.payload.data.participants;
                const participants = Array.isArray(raw)
                  ? raw
                  : raw && typeof raw === "object"
                    ? Object.values(raw)
                    : [];

                return {
                  ...prev,
                  participants: participants as Room["participants"],
                };
              });
              break;
            case "quiz_ended":
              setCurrentRoom((prev) => {
                if (!prev) return prev;

                const raw = message.payload.data.participants;
                const participants = Array.isArray(raw)
                  ? raw
                  : raw && typeof raw === "object"
                    ? Object.values(raw)
                    : prev.participants;

                return {
                  ...prev,
                  status: "finished",
                  participants: participants as Room["participants"],
                };
              });
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
        questionResult,
        setQuestionResult,
      }}
    >
      {children}
    </WebSocketContext.Provider>
  );
};
