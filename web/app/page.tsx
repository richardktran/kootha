"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";

import { useWebSocket } from "./contexts/WebSocketContext";

import { createUser, createRoom } from "@/services/api";
import { User } from "@/types/api";

export default function Home() {
  const [playerName, setPlayerName] = useState("");
  const [roomId, setRoomId] = useState("");
  const [roomName, setRoomName] = useState("");
  const [showJoinRoom, setShowJoinRoom] = useState(false);
  const [showCreateRoom, setShowCreateRoom] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [showNameForm, setShowNameForm] = useState(true);

  const router = useRouter();
  const { connect, sendMessage } = useWebSocket();

  useEffect(() => {
    // Check local storage for user data
    const storedUser = localStorage.getItem("quizUser");

    if (storedUser) {
      const user = JSON.parse(storedUser);

      setCurrentUser(user);
      setShowNameForm(false);
    }
  }, []);

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

    // Store user data in local storage
    localStorage.setItem("quizUser", JSON.stringify(response.data));
    setCurrentUser(response.data);
    setIsLoading(false);
    setShowNameForm(false);
  };

  const handleJoinRoom = () => {
    if (!currentUser || !roomId) return;
    connect();
    sendMessage({
      type: "JOIN_ROOM",
      payload: { roomId, participantName: currentUser.name },
    });
    router.push(`/quiz/${roomId}`);
  };

  const handleCreateRoom = async () => {
    if (!currentUser || !roomName) return;

    setIsLoading(true);
    setError(null);

    const response = await createRoom(roomName, currentUser.id);

    if (response.error || !response.data) {
      setError(response.error || "Failed to create room");
      setIsLoading(false);

      return;
    }

    connect();
    sendMessage({
      type: "JOIN_ROOM",
      payload: {
        roomId: response.data.roomId,
        participantName: currentUser.name,
      },
    });
    router.push(`/quiz/${response.data.roomId}`);
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
