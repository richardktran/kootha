"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

import { useWebSocket } from "./contexts/WebSocketContext";
import { useUser } from "./contexts/UserContext";

import { createUser, createRoom } from "@/services/api";

export default function Home() {
  const [playerName, setPlayerName] = useState("");
  const [roomId, setRoomId] = useState("");
  const [roomName, setRoomName] = useState("");
  const [showJoinRoom, setShowJoinRoom] = useState(false);
  const [showCreateRoom, setShowCreateRoom] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const router = useRouter();
  const { connect, sendMessage } = useWebSocket();
  const { user, setUser } = useUser();
  const showNameForm = !user;

  const handleNameSubmit = async () => {
    if (!playerName.trim()) {
      setError("Please enter your name");
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
    if (!user || !roomId) return;
    try {
      await connect();
      sendMessage({
        type: "JOIN_ROOM",
        payload: {
          roomId,
          userId: user.id,
          name: user.name,
        },
      });
      router.push(`/quiz/${roomId}`);
    } catch (err) {
      console.error("Failed to join room:", err);
      setError("Failed to connect to the server. Please try again.");
    }
  };

  const handleCreateRoom = async () => {
    if (!user || !roomName) return;

    setIsLoading(true);
    setError(null);

    try {
      const response = await createRoom(roomName, user.id);

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
      setError("Failed to connect to the server. Please try again.");
      setIsLoading(false);
    }
  };

  return (
    <main className="min-h-screen flex items-center justify-center bg-gradient-to-br from-purple-500 to-pink-500 p-4">
      <div className="w-full max-w-md">
        {showNameForm ? (
          <div className="bg-white rounded-lg shadow-xl p-8 space-y-6">
            <h1 className="text-3xl font-bold text-center text-gray-800">
              Welcome to QuizTime!
            </h1>
            <div className="space-y-2">
              <label
                className="block text-sm font-medium text-gray-700"
                htmlFor="playerName"
              >
                Enter your name
              </label>
              <input
                className="w-full px-4 py-2 border rounded-md focus:ring-2 focus:ring-purple-500"
                disabled={isLoading}
                id="playerName"
                placeholder="Your name"
                type="text"
                value={playerName}
                onChange={(e) => setPlayerName(e.target.value)}
              />
            </div>
            {error && <p className="text-red-500 text-sm">{error}</p>}
            <button
              className="w-full bg-purple-600 text-white px-4 py-2 rounded-md hover:bg-purple-700 transition disabled:opacity-50"
              disabled={isLoading}
              onClick={handleNameSubmit}
            >
              {isLoading ? "Loading..." : "Continue"}
            </button>
          </div>
        ) : (
          <div className="bg-white rounded-lg shadow-xl p-8 space-y-6">
            <h2 className="text-2xl font-bold text-center text-gray-800">
              Join or Create Room
            </h2>
            <p className="text-center text-gray-600 text-sm">
              Playing as <span className="font-semibold">{user?.name}</span>
            </p>

            <div className="flex gap-4">
              <button
                className="flex-1 bg-purple-600 text-white px-4 py-2 rounded-md hover:bg-purple-700 transition"
                onClick={() => {
                  setShowJoinRoom(true);
                  setShowCreateRoom(false);
                }}
              >
                Join Room
              </button>
              <button
                className="flex-1 bg-pink-600 text-white px-4 py-2 rounded-md hover:bg-pink-700 transition"
                onClick={() => {
                  setShowCreateRoom(true);
                  setShowJoinRoom(false);
                }}
              >
                Create Room
              </button>
            </div>

            {showJoinRoom && (
              <div className="space-y-4">
                <div className="space-y-2">
                  <label
                    className="block text-sm font-medium text-gray-700"
                    htmlFor="roomId"
                  >
                    Enter Room ID
                  </label>
                  <input
                    className="w-full px-4 py-2 border rounded-md focus:ring-2 focus:ring-purple-500"
                    disabled={isLoading}
                    id="roomId"
                    placeholder="Room ID"
                    type="text"
                    value={roomId}
                    onChange={(e) => setRoomId(e.target.value)}
                  />
                </div>
                <button
                  className="w-full bg-purple-600 text-white px-4 py-2 rounded-md hover:bg-purple-700 transition disabled:opacity-50"
                  disabled={isLoading}
                  onClick={handleJoinRoom}
                >
                  {isLoading ? "Loading..." : "Join"}
                </button>
              </div>
            )}

            {showCreateRoom && (
              <div className="space-y-4">
                <div className="space-y-2">
                  <label
                    className="block text-sm font-medium text-gray-700"
                    htmlFor="roomName"
                  >
                    Enter Room Name
                  </label>
                  <input
                    className="w-full px-4 py-2 border rounded-md focus:ring-2 focus:ring-purple-500"
                    disabled={isLoading}
                    id="roomName"
                    placeholder="Room Name"
                    type="text"
                    value={roomName}
                    onChange={(e) => setRoomName(e.target.value)}
                  />
                </div>
                <button
                  className="w-full bg-pink-600 text-white px-4 py-2 rounded-md hover:bg-pink-700 transition disabled:opacity-50"
                  disabled={isLoading}
                  onClick={handleCreateRoom}
                >
                  {isLoading ? "Loading..." : "Create"}
                </button>
              </div>
            )}

            {error && <p className="text-red-500 text-sm">{error}</p>}
          </div>
        )}
      </div>
    </main>
  );
}
