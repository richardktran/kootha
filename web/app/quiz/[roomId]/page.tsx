"use client";

import { useEffect, useRef, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { motion, AnimatePresence } from "framer-motion";

import { useWebSocket } from "../../contexts/WebSocketContext";
import { useUser } from "../../contexts/UserContext";

import { Room } from "@/types/quiz";
import {
  ArenaShell,
  BrandMark,
  StatusPill,
  GameButton,
  TimerRing,
  AnswerPad,
  PlayerRoster,
  Leaderboard,
  FeedbackBanner,
} from "@/components/game";

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

function CopyRoomCode({ roomId }: { roomId: string }) {
  const [copied, setCopied] = useState(false);

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(roomId);
      setCopied(true);
      setTimeout(() => setCopied(false), 1600);
    } catch {
      /* ignore */
    }
  };

  return (
    <button
      type="button"
      onClick={copy}
      className="group inline-flex items-center gap-2 rounded-xl border border-[var(--line)] bg-[var(--surface)] px-3 py-2 font-mono text-sm font-bold tracking-wide text-[var(--pulse-lime)] transition hover:bg-[var(--surface-strong)]"
      title="Copy room code"
    >
      <span className="text-[10px] uppercase tracking-wider text-[var(--muted)]">
        Code
      </span>
      {roomId}
      <span className="text-[10px] uppercase text-[var(--muted)] group-hover:text-white">
        {copied ? "Copied" : "Copy"}
      </span>
    </button>
  );
}

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

  const feedbackKind = (): "correct" | "incorrect" | "timeout" | null => {
    if (!questionResult) return null;
    if (selectedAnswer === null || selectedAnswer < 0) return "timeout";
    if (selectedAnswer === questionResult.correctAnswer) return "correct";
    return "incorrect";
  };

  if (!user) {
    return (
      <ArenaShell>
        <div className="flex min-h-dvh flex-col items-center justify-center gap-6 px-4 text-center">
          <BrandMark size="lg" />
          <p className="max-w-sm text-[var(--muted)]">
            Pick a name on the home screen before joining a room.
          </p>
          <GameButton onClick={handleBack}>Go home</GameButton>
        </div>
      </ArenaShell>
    );
  }

  if (isLoading) {
    return (
      <ArenaShell>
        <div className="flex min-h-dvh flex-col items-center justify-center gap-4">
          <div className="h-12 w-12 animate-spin rounded-full border-4 border-[var(--pulse-lime)] border-t-transparent" />
          <p className="text-display text-xl text-[var(--pulse-lime)]">
            Loading arena…
          </p>
        </div>
      </ArenaShell>
    );
  }

  if (!currentRoom) {
    return (
      <ArenaShell>
        <div className="flex min-h-dvh flex-col items-center justify-center gap-6 px-4 text-center">
          <BrandMark size="md" />
          <h1 className="text-display text-3xl">Room not found</h1>
          <p className="text-[var(--muted)]">
            That code doesn’t match an active room.
          </p>
          <GameButton onClick={handleBack}>Back home</GameButton>
        </div>
      </ArenaShell>
    );
  }

  const feedback = feedbackKind();
  const showQuestion =
    currentRoom.status === "in-progress" &&
    currentQuestion &&
    !showLeaderboard;
  const participants = currentRoom.participants ?? [];
  const questionNumber = (currentRoom.currentQuestionIndex ?? 0) + 1;

  return (
    <ArenaShell>
      <main className="mx-auto min-h-dvh w-full max-w-3xl px-4 py-5 sm:px-6 sm:py-8">
        {/* Top bar */}
        <header className="mb-5 flex flex-wrap items-center justify-between gap-3">
          <div className="min-w-0">
            <BrandMark size="sm" />
            <h1 className="mt-1 truncate text-display text-2xl sm:text-3xl">
              {currentRoom.name || "Quiz room"}
            </h1>
            <div className="mt-2 flex flex-wrap items-center gap-2">
              <CopyRoomCode roomId={currentRoom.id} />
              <StatusPill>
                {participants.length} player{participants.length === 1 ? "" : "s"}
              </StatusPill>
              {!isConnected ? (
                <StatusPill tone="amber">Reconnecting…</StatusPill>
              ) : null}
              {isHost ? <StatusPill tone="amber">You’re host</StatusPill> : null}
            </div>
          </div>
          <GameButton variant="ghost" onClick={handleBack}>
            ← Leave
          </GameButton>
        </header>

        <AnimatePresence mode="wait">
          {/* LOBBY */}
          {currentRoom.status === "waiting" ? (
            <motion.div
              key="lobby"
              initial={{ opacity: 0, y: 12 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0 }}
              className="space-y-5"
            >
              <div className="panel overflow-hidden p-6 sm:p-8">
                <div className="mb-6 text-center">
                  <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-[var(--pulse-lime)] text-display text-3xl text-[var(--arena-ink)] shadow-pop animate-float">
                    ?
                  </div>
                  <h2 className="text-display text-3xl text-[var(--pulse-lime)]">
                    Lobby
                  </h2>
                  <p className="mt-2 text-[var(--muted)]">
                    Share the room code. Start when everyone is ready.
                  </p>
                </div>

                <PlayerRoster
                  participants={participants}
                  hostId={currentRoom.hostId}
                  currentUserId={user.id}
                />

                <div className="mt-8 text-center">
                  {isHost ? (
                    <GameButton
                      variant="host"
                      disabled={participants.length < 1}
                      onClick={handleStartQuiz}
                    >
                      Start quiz
                    </GameButton>
                  ) : (
                    <div className="inline-flex items-center gap-3 rounded-2xl border border-[var(--line)] bg-[var(--surface)] px-5 py-4">
                      <span className="h-2.5 w-2.5 animate-pulse rounded-full bg-[var(--pulse-amber)]" />
                      <span className="font-semibold text-[var(--muted)]">
                        Waiting for host to start…
                      </span>
                    </div>
                  )}
                </div>
              </div>
            </motion.div>
          ) : null}

          {/* QUESTION */}
          {showQuestion ? (
            <motion.div
              key={`q-${currentQuestion.id}`}
              initial={{ opacity: 0, y: 16 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, scale: 0.98 }}
              className="space-y-4"
            >
              <div className="flex items-center justify-between gap-3">
                <StatusPill tone="lime">Round {questionNumber}</StatusPill>
                <PlayerRoster
                  compact
                  participants={participants}
                  hostId={currentRoom.hostId}
                  currentUserId={user.id}
                />
              </div>

              <div className="panel p-5 sm:p-7">
                {!isRevealed ? (
                  <div className="mb-5">
                    <TimerRing
                      timeLeft={timeLeft}
                      timeLimit={timeLimitRef.current}
                    />
                  </div>
                ) : null}

                {feedback ? (
                  <div className="mb-5">
                    <FeedbackBanner kind={feedback} />
                  </div>
                ) : null}

                <h2 className="mb-6 text-center text-display text-2xl leading-snug sm:text-3xl">
                  {currentQuestion.question}
                </h2>

                <AnswerPad
                  options={currentQuestion.options}
                  selectedAnswer={selectedAnswer}
                  correctAnswer={questionResult?.correctAnswer}
                  isRevealed={isRevealed}
                  disabled={hasSubmitted || isRevealed}
                  onSelect={handleAnswerSelect}
                />

                <div className="mt-6">
                  {!hasSubmitted && !isRevealed ? (
                    <GameButton
                      fullWidth
                      disabled={selectedAnswer === null}
                      onClick={() => submitAnswer(selectedAnswer)}
                    >
                      Lock in answer
                    </GameButton>
                  ) : null}

                  {hasSubmitted && !isRevealed ? (
                    <div className="rounded-[var(--radius-tile)] border border-[var(--line)] bg-[var(--surface)] py-4 text-center font-semibold text-[var(--muted)]">
                      Locked in — waiting for the reveal…
                    </div>
                  ) : null}
                </div>
              </div>
            </motion.div>
          ) : null}

          {/* LEADERBOARD / FINISHED */}
          {(showLeaderboard || currentRoom.status === "finished") && (
            <motion.div
              key="board"
              initial={{ opacity: 0, y: 16 }}
              animate={{ opacity: 1, y: 0 }}
              className="panel p-5 sm:p-8"
            >
              {feedback && currentRoom.status !== "finished" ? (
                <div className="mb-5">
                  <FeedbackBanner kind={feedback} />
                </div>
              ) : null}

              <Leaderboard
                participants={participants}
                currentUserId={user.id}
                finished={currentRoom.status === "finished"}
              />

              {isHost &&
              currentRoom.status === "in-progress" &&
              isRevealed ? (
                <div className="mt-6">
                  <GameButton fullWidth variant="host" onClick={handleNextQuestion}>
                    Next question →
                  </GameButton>
                </div>
              ) : null}

              {!isHost &&
              currentRoom.status === "in-progress" &&
              isRevealed ? (
                <p className="mt-6 text-center text-sm font-semibold text-[var(--muted)]">
                  Host is advancing to the next round…
                </p>
              ) : null}

              {currentRoom.status === "finished" ? (
                <div className="mt-6">
                  <GameButton fullWidth onClick={handleBack}>
                    Play again
                  </GameButton>
                </div>
              ) : null}
            </motion.div>
          )}
        </AnimatePresence>
      </main>
    </ArenaShell>
  );
}
