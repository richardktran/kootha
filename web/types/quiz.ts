export interface Question {
  id: string;
  question: string;
  options: string[];
  correctAnswer: number;
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

export interface Answer {
  participantId: string;
  questionId: string;
  selectedOption: number;
  timeToAnswer: number;
}

export type WebSocketEvent =
  | { type: "JOIN_ROOM"; payload: { roomId: string; userId: string } }
  | { type: "ROOM_JOINED"; payload: Room }
  | { type: "START_QUIZ"; payload: { roomId: string } }
  | { type: "NEXT_QUESTION"; payload: { roomId: string } }
  | { type: "QUESTION_TIMEOUT"; payload: { roomId: string } }
  | { type: "SUBMIT_ANSWER"; payload: Answer }
  | { type: "LEADERBOARD_UPDATE"; payload: { participants: Participant[] } }
  | { type: "QUIZ_ENDED"; payload: { participants: Participant[] } }
  | { type: "PARTICIPANT_JOINED"; payload: { participant: Participant } }
  | { type: "PARTICIPANT_LEFT"; payload: { participantId: string } }
  | { type: "QUESTION_STARTED"; payload: { questionIndex: number } };
