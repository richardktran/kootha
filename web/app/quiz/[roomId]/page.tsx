"use client";

import { useEffect, useRef, useState } from "react";
import { useParams, useRouter } from "next/navigation";

import { useWebSocket } from "../../contexts/WebSocketContext";
import { useUser } from "../../contexts/UserContext";

import { Room } from "@/types/quiz";

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export default function QuizRoom() {
  const { roomId } = useParams();
  const router = useRouter();
  const {
    currentRoom,
    currentQuestion,
    questionResult,
    sendMessage,
    setCurrentRoom,
    setQuestionResult,
    connect,
    isConnected,
  } = useWebSocket();
  const { user } = useUser();
  const [timeLeft, setTimeLeft] = useState<number>(15);
  const [selectedAnswer, setSelectedAnswer] = useState<number | null>(null);
  const [showLeaderboard, setShowLeaderboard] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [hasSubmitted, setHasSubmitted] = useState(false);
  const joinedRef = useRef(false);
  const timeLimitRef = useRef(15);
  const submittedRef = useRef(false);

  const roomIdStr = roomId as string;
  const isHost = !!user && currentRoom?.hostId === user.id;
  const isRevealed = !!questionResult;

  // Fetch room + rejoin via WS on mount / refresh
  useEffect(() => {
    if (!roomIdStr || !user) {
      setIsLoading(false);
      return;
    }

    let cancelled = false;

    const bootstrap = async () => {
      try {
        const response = await fetch(
          `${API_BASE_URL}/api/rooms?id=${roomIdStr}`,
        );
        if (response.ok) {
          const data = await response.json();
          const room: Room = {
            id: data.id,
            name: data.name ?? "",
            hostId: data.hostId ?? "",
            participants: data.participants ?? [],
            status: data.status ?? "waiting",
            currentQuestionIndex: data.currentQuestionIndex ?? 0,
            questions: data.questions ?? [],
          };
          if (!cancelled) {
            setCurrentRoom((prev) =>
              prev?.id === roomIdStr && (prev.participants?.length ?? 0) > 0
                ? prev
                : room,
            );
          }
        }

        await connect();
        if (!joinedRef.current) {
          sendMessage({
            type: "JOIN_ROOM",
            payload: {
              roomId: roomIdStr,
              userId: user.id,
              name: user.name,
            },
          });
          joinedRef.current = true;
        }
      } catch (error) {
        console.error("Error bootstrapping room:", error);
      } finally {
        if (!cancelled) setIsLoading(false);
      }
    };

    bootstrap();
    return () => {
      cancelled = true;
    };
  }, [roomIdStr, user?.id]);

  // Reset state when a new question arrives
  useEffect(() => {
    if (!currentQuestion) return;
    const limit = currentQuestion.timeLimit || 15;
    timeLimitRef.current = limit;
    setTimeLeft(limit);
    setSelectedAnswer(null);
    setShowLeaderboard(false);
    setHasSubmitted(false);
    submittedRef.current = false;
    setQuestionResult(null);
  }, [currentQuestion?.id]);

  // After reveal, show leaderboard shortly so players can see correct/incorrect first
  useEffect(() => {
    if (!questionResult) return;
    if (!submittedRef.current) {
      submitAnswer(selectedAnswer);
    }
    const t = setTimeout(() => setShowLeaderboard(true), 1800);
    return () => clearTimeout(t);
  }, [questionResult?.questionId, questionResult?.reason]);

  useEffect(() => {
    if (!currentQuestion || showLeaderboard || isRevealed || hasSubmitted) return;
    if (timeLeft <= 0) {
      submitAnswer(selectedAnswer);
      return;
    }

    const timer = setTimeout(() => setTimeLeft((t) => t - 1), 1000);
    return () => clearTimeout(timer);
  }, [
    currentQuestion,
    timeLeft,
    showLeaderboard,
    isRevealed,
    hasSubmitted,
    selectedAnswer,
  ]);

  const submitAnswer = (option: number | null) => {
    if (!user || !currentQuestion || submittedRef.current) return;
    submittedRef.current = true;
    setHasSubmitted(true);

    sendMessage({
      type: "SUBMIT_ANSWER",
      payload: {
        roomId: roomIdStr,
        userId: user.id,
        answer: {
          selectedOption: option ?? -1,
          questionId: currentQuestion.id,
          timeToAnswer: Math.max(0, timeLimitRef.current - timeLeft),
        },
      },
    });
  };

  const handleAnswerSelect = (optionIndex: number) => {
    if (hasSubmitted || isRevealed) return;
    setSelectedAnswer(optionIndex);
  };

  const handleStartQuiz = () => {
    if (!user) return;
    sendMessage({
      type: "START_QUIZ",
      payload: { roomId: roomIdStr, userId: user.id },
    });
  };

  const handleNextQuestion = () => {
    if (!user) return;
    setSelectedAnswer(null);
    setShowLeaderboard(false);
    setHasSubmitted(false);
    submittedRef.current = false;
    setQuestionResult(null);
    sendMessage({
      type: "NEXT_QUESTION",
      payload: { roomId: roomIdStr, userId: user.id },
    });
  };

  const handleBack = () => {
    router.push("/");
  };

  const optionClassName = (index: number) => {
    if (isRevealed && questionResult) {
      const isCorrect = index === questionResult.correctAnswer;
      const isSelected = selectedAnswer === index;
      if (isCorrect) {
        return "bg-green-600 text-white ring-2 ring-green-300";
      }
      if (isSelected && !isCorrect) {
        return "bg-red-500 text-white ring-2 ring-red-300";
      }
      return "bg-gray-100 text-gray-500 opacity-70";
    }

    return selectedAnswer === index
      ? "bg-purple-600 text-white"
      : "bg-gray-100 hover:bg-gray-200 text-gray-800";
  };

  const answerFeedback = () => {
    if (!questionResult) return null;
    if (selectedAnswer === null || selectedAnswer < 0) {
      return {
        text: "Time's up — no answer submitted",
        className: "bg-amber-50 text-amber-800 border-amber-200",
      };
    }
    if (selectedAnswer === questionResult.correctAnswer) {
      return {
        text: "Correct!",
        className: "bg-green-50 text-green-800 border-green-200",
      };
    }
    return {
      text: "Incorrect",
      className: "bg-red-50 text-red-800 border-red-200",
    };
  };

  if (!user) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-2xl">Please enter your name on the home page first.</div>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-2xl">Loading...</div>
      </div>
    );
  }

  if (!currentRoom) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-2xl">Room not found</div>
      </div>
    );
  }

  const feedback = answerFeedback();
  const showQuestion =
    currentRoom.status === "in-progress" &&
    currentQuestion &&
    !showLeaderboard;

  return (
    <main className="min-h-screen bg-gradient-to-br from-purple-500 to-pink-500 p-4">
      <div className="max-w-4xl mx-auto">
        <div className="bg-white rounded-lg shadow-xl p-6 mb-6">
          <div className="flex items-start justify-between gap-4 mb-2">
            <div>
              <h1 className="text-2xl font-bold mb-2">Room: {currentRoom.name}</h1>
              <p className="text-gray-600">Room ID: {currentRoom.id}</p>
              <p className="text-gray-600">
                Players: {(currentRoom.participants ?? []).length}
                {!isConnected && (
                  <span className="ml-2 text-amber-600 text-sm">(reconnecting…)</span>
                )}
              </p>
            </div>
            <button
              type="button"
              onClick={handleBack}
              className="shrink-0 px-4 py-2 rounded-lg border border-gray-300 text-gray-700 font-semibold hover:bg-gray-50 transition"
            >
              ← Back
            </button>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow-xl p-6 mb-6">
          <h2 className="text-xl font-bold mb-4">Participants</h2>
          <div className="space-y-2">
            {(currentRoom.participants ?? []).map((participant) => (
              <div
                key={participant.id}
                className="flex items-center justify-between bg-gray-50 p-3 rounded-lg"
              >
                <div className="flex items-center">
                  <div className="font-semibold">{participant.name}</div>
                  {participant.id === currentRoom.hostId && (
                    <span className="ml-2 text-xs bg-purple-100 text-purple-800 px-2 py-1 rounded">
                      Host
                    </span>
                  )}
                  {participant.id === user.id && (
                    <span className="ml-2 text-xs bg-blue-100 text-blue-800 px-2 py-1 rounded">
                      You
                    </span>
                  )}
                </div>
                <div className="text-gray-600">Score: {participant.score}</div>
              </div>
            ))}
          </div>
        </div>

        {currentRoom.status === "waiting" && (
          <div className="text-center">
            {isHost ? (
              <button
                className="bg-green-600 text-white px-8 py-4 rounded-lg text-xl font-bold hover:bg-green-700 transition"
                onClick={handleStartQuiz}
              >
                Start Quiz
              </button>
            ) : (
              <div className="text-xl text-gray-700">
                Waiting for host to start the quiz...
              </div>
            )}
          </div>
        )}

        {showQuestion && (
          <div className="bg-white rounded-lg shadow-xl p-6">
            {!isRevealed && (
              <div className="text-center mb-6">
                <div className="text-4xl font-bold text-purple-600">
                  {timeLeft}
                </div>
                <div className="text-gray-600">seconds remaining</div>
              </div>
            )}

            {feedback && (
              <div
                className={`mb-6 text-center text-xl font-bold border rounded-lg py-3 ${feedback.className}`}
              >
                {feedback.text}
              </div>
            )}

            <div className="mb-8">
              <h2 className="text-2xl font-bold mb-4">
                {currentQuestion.question}
              </h2>
              <div className="grid grid-cols-2 gap-4">
                {currentQuestion.options.map((option, index) => (
                  <button
                    key={index}
                    disabled={hasSubmitted || isRevealed}
                    className={`p-4 rounded-lg text-lg font-semibold transition disabled:cursor-default ${optionClassName(index)}`}
                    onClick={() => handleAnswerSelect(index)}
                  >
                    {option}
                  </button>
                ))}
              </div>
            </div>

            {!hasSubmitted && !isRevealed && (
              <button
                className="w-full bg-purple-600 text-white px-6 py-3 rounded-lg font-bold hover:bg-purple-700 transition disabled:opacity-50"
                disabled={selectedAnswer === null}
                onClick={() => submitAnswer(selectedAnswer)}
              >
                Submit Answer
              </button>
            )}

            {hasSubmitted && !isRevealed && (
              <div className="text-center text-gray-600 font-medium py-3">
                Answer submitted — waiting for other players or the timer…
              </div>
            )}
          </div>
        )}

        {(showLeaderboard || currentRoom.status === "finished") && (
          <div className="bg-white rounded-lg shadow-xl p-6">
            {feedback && currentRoom.status !== "finished" && (
              <div
                className={`mb-6 text-center text-lg font-bold border rounded-lg py-3 ${feedback.className}`}
              >
                {feedback.text}
              </div>
            )}

            <h2 className="text-2xl font-bold mb-6">
              {currentRoom.status === "finished" ? "Final Results" : "Leaderboard"}
            </h2>
            <div className="space-y-4">
              {[...(currentRoom.participants ?? [])]
                .sort((a, b) => b.score - a.score)
                .map((participant, index) => (
                  <div
                    key={participant.id}
                    className="flex items-center justify-between bg-gray-50 p-4 rounded-lg"
                  >
                    <div className="flex items-center">
                      <div className="text-2xl font-bold text-gray-400 mr-4">
                        #{index + 1}
                      </div>
                      <div className="font-semibold">{participant.name}</div>
                    </div>
                    <div className="text-xl font-bold text-purple-600">
                      {participant.score}
                    </div>
                  </div>
                ))}
            </div>

            {isHost && currentRoom.status === "in-progress" && isRevealed && (
              <button
                className="mt-6 w-full bg-purple-600 text-white px-6 py-3 rounded-lg font-bold hover:bg-purple-700 transition"
                onClick={handleNextQuestion}
              >
                Next Question
              </button>
            )}

            {currentRoom.status === "finished" && (
              <div className="mt-6 space-y-3">
                <p className="text-center text-gray-600">
                  Quiz finished — thanks for playing!
                </p>
                <button
                  type="button"
                  onClick={handleBack}
                  className="w-full bg-purple-600 text-white px-6 py-3 rounded-lg font-bold hover:bg-purple-700 transition"
                >
                  Back to Home
                </button>
              </div>
            )}
          </div>
        )}
      </div>
    </main>
  );
}
