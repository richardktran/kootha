export interface Question {
  id: string;
  question: string;
  options: string[];
  correctAnswer?: number;
  timeLimit: number;
}

export interface Room {
  id: string;
  name: string;
  hostId: string;
  participants: Participant[];
  status: "waiting" | "in-progress" | "finished";
  currentQuestionIndex: number;
  questions: Question[];
}

export interface Participant {
  id: string;
  name: string;
  score: number;
}

export interface AnswerPayload {
  selectedOption: number;
  questionId: string;
  timeToAnswer: number;
}

export type WebSocketEvent =
  | {
      type: "JOIN_ROOM";
      payload: { roomId: string; userId: string; name: string };
    }
  | { type: "START_QUIZ"; payload: { roomId: string; userId: string } }
  | { type: "NEXT_QUESTION"; payload: { roomId: string; userId: string } }
  | {
      type: "SUBMIT_ANSWER";
      payload: { roomId: string; userId?: string; answer: AnswerPayload };
    };
