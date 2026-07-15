"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { motion, AnimatePresence } from "framer-motion";

import { useWebSocket } from "./contexts/WebSocketContext";
import { useUser } from "./contexts/UserContext";
import {
  ArenaShell,
  BrandMark,
  GameButton,
  GameInput,
} from "@/components/game";

import { createUser, createRoom } from "@/services/api";

type Mode = "hub" | "join" | "create";

export default function Home() {
  const [playerName, setPlayerName] = useState("");
  const [roomId, setRoomId] = useState("");
  const [roomName, setRoomName] = useState("");
  const [mode, setMode] = useState<Mode>("hub");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const router = useRouter();
  const { connect, sendMessage } = useWebSocket();
  const { user, setUser } = useUser();
  const showNameForm = !user;

  const handleNameSubmit = async () => {
    if (!playerName.trim()) {
      setError("Pick a name to enter the arena");
      return;
    }

    setIsLoading(true);
    setError(null);

    const response = await createUser(playerName);

    if (response.error || !response.data) {
      setError(response.error || "Failed to create user");
      setIsLoading(false);
      return;
    }

    setUser(response.data);
    setIsLoading(false);
  };

  const handleJoinRoom = async () => {
    if (!user || !roomId.trim()) {
      setError("Enter a room code");
      return;
    }
    setIsLoading(true);
    setError(null);
    try {
      await connect();
      sendMessage({
        type: "JOIN_ROOM",
        payload: {
          roomId: roomId.trim(),
          userId: user.id,
          name: user.name,
        },
      });
      router.push(`/quiz/${roomId.trim()}`);
    } catch (err) {
      console.error("Failed to join room:", err);
      setError("Couldn’t connect. Is the server running?");
      setIsLoading(false);
    }
  };

  const handleCreateRoom = async () => {
    if (!user || !roomName.trim()) {
      setError("Give your room a name");
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const response = await createRoom(roomName.trim(), user.id);

      if (response.error || !response.data) {
        setError(response.error || "Failed to create room");
        setIsLoading(false);
        return;
      }

      await connect();
      sendMessage({
        type: "JOIN_ROOM",
        payload: {
          roomId: response.data.roomId,
          userId: user.id,
          name: user.name,
        },
      });
      router.push(`/quiz/${response.data.roomId}`);
    } catch (err) {
      console.error("Failed to create room:", err);
      setError("Couldn’t connect. Is the server running?");
      setIsLoading(false);
    }
  };

  return (
    <ArenaShell>
      <main className="mx-auto flex min-h-dvh w-full max-w-lg flex-col justify-center px-4 py-10 sm:px-6">
        <div className="mb-8 text-center sm:mb-10">
          <div className="mb-4 flex justify-center">
            <BrandMark size="hero" />
          </div>
          <motion.p
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.2 }}
            className="mx-auto max-w-sm text-base text-[var(--muted)] sm:text-lg"
          >
            Create a room, drop the code, and battle friends in a live quiz.
          </motion.p>
        </div>

        <AnimatePresence mode="wait">
          {showNameForm ? (
            <motion.div
              key="name"
              initial={{ opacity: 0, y: 18 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -12 }}
              className="panel-solid space-y-5 p-6 sm:p-8"
            >
              <div>
                <h1 className="text-display text-2xl text-[var(--ink)]">
                  Enter the arena
                </h1>
                <p className="mt-1 text-sm text-black/55">
                  Choose a display name — no account needed.
                </p>
              </div>

              <GameInput
                light
                id="playerName"
                label="Your name"
                placeholder="e.g. QuizChamp"
                disabled={isLoading}
                value={playerName}
                maxLength={24}
                onChange={(e) => setPlayerName(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === "Enter") handleNameSubmit();
                }}
              />

              {error ? (
                <p className="text-sm font-semibold text-[var(--pulse-coral)]">
                  {error}
                </p>
              ) : null}

              <GameButton
                fullWidth
                disabled={isLoading}
                onClick={handleNameSubmit}
              >
                {isLoading ? "Joining…" : "Let’s play"}
              </GameButton>
            </motion.div>
          ) : (
            <motion.div
              key="hub"
              initial={{ opacity: 0, y: 18 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -12 }}
              className="space-y-4"
            >
              <div className="panel flex items-center justify-between gap-3 px-4 py-3">
                <div>
                  <div className="text-xs font-bold uppercase tracking-wider text-[var(--muted)]">
                    Playing as
                  </div>
                  <div className="text-display text-xl text-[var(--pulse-lime)]">
                    {user?.name}
                  </div>
                </div>
                <div className="h-12 w-12 rounded-2xl bg-[var(--pulse-coral)] text-center text-display text-xl leading-[3rem] text-white shadow-pop">
                  {user?.name?.slice(0, 1).toUpperCase()}
                </div>
              </div>

              {mode === "hub" ? (
                <div className="grid gap-3">
                  <GameButton
                    fullWidth
                    variant="primary"
                    onClick={() => {
                      setMode("join");
                      setError(null);
                    }}
                  >
                    Join a room
                  </GameButton>
                  <GameButton
                    fullWidth
                    variant="secondary"
                    onClick={() => {
                      setMode("create");
                      setError(null);
                    }}
                  >
                    Host a room
                  </GameButton>
                </div>
              ) : null}

              <AnimatePresence mode="wait">
                {mode === "join" ? (
                  <motion.div
                    key="join"
                    initial={{ opacity: 0, y: 10 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0 }}
                    className="panel-solid space-y-4 p-6"
                  >
                    <div className="flex items-start justify-between gap-2">
                      <div>
                        <h2 className="text-display text-xl text-[var(--ink)]">
                          Join room
                        </h2>
                        <p className="text-sm text-black/55">
                          Paste the code from your host.
                        </p>
                      </div>
                      <button
                        type="button"
                        className="text-sm font-bold text-black/45 hover:text-black"
                        onClick={() => setMode("hub")}
                      >
                        Back
                      </button>
                    </div>
                    <GameInput
                      light
                      id="roomId"
                      label="Room code"
                      placeholder="Room ID"
                      disabled={isLoading}
                      value={roomId}
                      onChange={(e) => setRoomId(e.target.value)}
                      onKeyDown={(e) => {
                        if (e.key === "Enter") handleJoinRoom();
                      }}
                    />
                    {error ? (
                      <p className="text-sm font-semibold text-[var(--pulse-coral)]">
                        {error}
                      </p>
                    ) : null}
                    <GameButton
                      fullWidth
                      disabled={isLoading}
                      onClick={handleJoinRoom}
                    >
                      {isLoading ? "Connecting…" : "Jump in"}
                    </GameButton>
                  </motion.div>
                ) : null}

                {mode === "create" ? (
                  <motion.div
                    key="create"
                    initial={{ opacity: 0, y: 10 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0 }}
                    className="panel-solid space-y-4 p-6"
                  >
                    <div className="flex items-start justify-between gap-2">
                      <div>
                        <h2 className="text-display text-xl text-[var(--ink)]">
                          Host a room
                        </h2>
                        <p className="text-sm text-black/55">
                          Name it, then share the code.
                        </p>
                      </div>
                      <button
                        type="button"
                        className="text-sm font-bold text-black/45 hover:text-black"
                        onClick={() => setMode("hub")}
                      >
                        Back
                      </button>
                    </div>
                    <GameInput
                      light
                      id="roomName"
                      label="Room name"
                      placeholder="Friday Night Trivia"
                      disabled={isLoading}
                      value={roomName}
                      maxLength={40}
                      onChange={(e) => setRoomName(e.target.value)}
                      onKeyDown={(e) => {
                        if (e.key === "Enter") handleCreateRoom();
                      }}
                    />
                    {error ? (
                      <p className="text-sm font-semibold text-[var(--pulse-coral)]">
                        {error}
                      </p>
                    ) : null}
                    <GameButton
                      fullWidth
                      variant="secondary"
                      disabled={isLoading}
                      onClick={handleCreateRoom}
                    >
                      {isLoading ? "Creating…" : "Create & enter"}
                    </GameButton>
                  </motion.div>
                ) : null}
              </AnimatePresence>
            </motion.div>
          )}
        </AnimatePresence>
      </main>
    </ArenaShell>
  );
}
