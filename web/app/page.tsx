"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

import { useWebSocket } from "./contexts/WebSocketContext";

export default function Home() {
  const [playerName, setPlayerName] = useState("");
  const [roomId, setRoomId] = useState("");
  const [roomName, setRoomName] = useState("");
  const [showJoinRoom, setShowJoinRoom] = useState(false);
  const [showCreateRoom, setShowCreateRoom] = useState(false);

  const router = useRouter();
  const { connect, sendMessage } = useWebSocket();

  const handleJoinRoom = () => {
    if (!playerName || !roomId) return;
    connect();
    sendMessage({
      type: "JOIN_ROOM",
      payload: { roomId, participantName: playerName },
    });
    router.push(`/quiz/${roomId}`);
  };

  const handleCreateRoom = () => {
    if (!playerName || !roomName) return;
    connect();
    // In a real implementation, we would create the room first and get the ID
    const newRoomId = "room-" + Math.random().toString(36).substr(2, 9);

    sendMessage({
      type: "JOIN_ROOM",
      payload: { roomId: newRoomId, participantName: playerName },
    });
    router.push(`/quiz/${newRoomId}`);
  };

  return (
    <main className="min-h-screen flex items-center justify-center bg-gradient-to-br from-purple-500 to-pink-500 p-4">
      <div className="bg-white rounded-lg shadow-xl p-8 max-w-md w-full space-y-6">
        <h1 className="text-3xl font-bold text-center text-gray-800">
          Welcome to Kootha!
        </h1>

        {/* Player Name Input */}
        <div className="space-y-2">
          <label
            className="block text-sm font-medium text-gray-700"
            htmlFor="playerName"
          >
            Enter your name
          </label>
          <input
            className="w-full px-4 py-2 border rounded-md focus:ring-2 focus:ring-purple-500"
            id="playerName"
            placeholder="Your name"
            type="text"
            value={playerName}
            onChange={(e) => setPlayerName(e.target.value)}
          />
        </div>

        {/* Action Buttons */}
        <div className="flex gap-4">
          <button
            className="flex-1 bg-purple-600 text-white px-4 py-2 rounded-md hover:bg-purple-700 transition"
            onClick={() => setShowJoinRoom(true)}
          >
            Join Room
          </button>
          <button
            className="flex-1 bg-pink-600 text-white px-4 py-2 rounded-md hover:bg-pink-700 transition"
            onClick={() => setShowCreateRoom(true)}
          >
            Create Room
          </button>
        </div>

        {/* Join Room Form */}
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
                id="roomId"
                placeholder="Room ID"
                type="text"
                value={roomId}
                onChange={(e) => setRoomId(e.target.value)}
              />
            </div>
            <button
              className="w-full bg-purple-600 text-white px-4 py-2 rounded-md hover:bg-purple-700 transition"
              onClick={handleJoinRoom}
            >
              Join
            </button>
          </div>
        )}

        {/* Create Room Form */}
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
                id="roomName"
                placeholder="Room Name"
                type="text"
                value={roomName}
                onChange={(e) => setRoomName(e.target.value)}
              />
            </div>
            <button
              className="w-full bg-pink-600 text-white px-4 py-2 rounded-md hover:bg-pink-700 transition"
              onClick={handleCreateRoom}
            >
              Create
            </button>
          </div>
        )}
      </div>
    </main>
  );
}
