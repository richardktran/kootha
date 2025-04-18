'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import { useWebSocket } from '../../contexts/WebSocketContext';
import { Question, Participant } from '@/types/quiz';

export default function QuizRoom() {
  const { roomId } = useParams();
  const { currentRoom, sendMessage } = useWebSocket();
  const [currentQuestion, setCurrentQuestion] = useState<Question | null>(null);
  const [timeLeft, setTimeLeft] = useState<number>(15);
  const [selectedAnswer, setSelectedAnswer] = useState<number | null>(null);
  const [showLeaderboard, setShowLeaderboard] = useState(false);

  useEffect(() => {
    let timer: NodeJS.Timeout;
    
    if (currentQuestion && timeLeft > 0) {
      timer = setInterval(() => {
        setTimeLeft((prev) => {
          if (prev <= 1) {
            clearInterval(timer);
            handleTimeUp();
            return 0;
          }
          return prev - 1;
        });
      }, 1000);
    }

    return () => {
      if (timer) clearInterval(timer);
    };
  }, [currentQuestion, timeLeft]);

  const handleTimeUp = () => {
    if (selectedAnswer !== null) {
      sendMessage({
        type: 'SUBMIT_ANSWER',
        payload: {
          participantId: 'current-user-id', // This should come from auth context
          questionId: currentQuestion!.id,
          selectedOption: selectedAnswer,
          timeToAnswer: 15 - timeLeft,
        },
      });
    }
    setShowLeaderboard(true);
  };

  const handleAnswerSelect = (optionIndex: number) => {
    setSelectedAnswer(optionIndex);
  };

  const handleStartQuiz = () => {
    sendMessage({
      type: 'START_QUIZ',
      payload: { roomId: roomId as string },
    });
  };

  const handleNextQuestion = () => {
    sendMessage({
      type: 'NEXT_QUESTION',
      payload: { roomId: roomId as string },
    });
  };

  if (!currentRoom) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-2xl">Loading...</div>
      </div>
    );
  }

  return (
    <main className="min-h-screen bg-gradient-to-br from-purple-500 to-pink-500 p-4">
      <div className="max-w-4xl mx-auto">
        {/* Room Info */}
        <div className="bg-white rounded-lg shadow-xl p-6 mb-6">
          <h1 className="text-2xl font-bold mb-2">Room: {currentRoom.name}</h1>
          <p className="text-gray-600">Room ID: {currentRoom.id}</p>
          <p className="text-gray-600">Players: {currentRoom.participants.length}</p>
        </div>

        {/* Quiz Content */}
        {currentRoom.status === 'waiting' && currentRoom.hostId === 'current-user-id' && (
          <div className="text-center">
            <button
              onClick={handleStartQuiz}
              className="bg-green-600 text-white px-8 py-4 rounded-lg text-xl font-bold hover:bg-green-700 transition"
            >
              Start Quiz
            </button>
          </div>
        )}

        {currentRoom.status === 'in-progress' && currentQuestion && !showLeaderboard && (
          <div className="bg-white rounded-lg shadow-xl p-6">
            {/* Timer */}
            <div className="text-center mb-6">
              <div className="text-4xl font-bold text-purple-600">{timeLeft}</div>
              <div className="text-gray-600">seconds remaining</div>
            </div>

            {/* Question */}
            <div className="mb-8">
              <h2 className="text-2xl font-bold mb-4">{currentQuestion.question}</h2>
              <div className="grid grid-cols-2 gap-4">
                {currentQuestion.options.map((option, index) => (
                  <button
                    key={index}
                    onClick={() => handleAnswerSelect(index)}
                    className={`p-4 rounded-lg text-lg font-semibold transition
                      ${selectedAnswer === index
                        ? 'bg-purple-600 text-white'
                        : 'bg-gray-100 hover:bg-gray-200 text-gray-800'
                      }`}
                  >
                    {option}
                  </button>
                ))}
              </div>
            </div>
          </div>
        )}

        {/* Leaderboard */}
        {showLeaderboard && (
          <div className="bg-white rounded-lg shadow-xl p-6">
            <h2 className="text-2xl font-bold mb-6">Leaderboard</h2>
            <div className="space-y-4">
              {currentRoom.participants
                .sort((a, b) => b.score - a.score)
                .map((participant, index) => (
                  <div
                    key={participant.id}
                    className="flex items-center justify-between bg-gray-50 p-4 rounded-lg"
                  >
                    <div className="flex items-center">
                      <div className="text-2xl font-bold text-gray-400 mr-4">#{index + 1}</div>
                      <div className="font-semibold">{participant.name}</div>
                    </div>
                    <div className="text-xl font-bold text-purple-600">{participant.score}</div>
                  </div>
                ))}
            </div>

            {currentRoom.hostId === 'current-user-id' && currentRoom.status !== 'finished' && (
              <button
                onClick={handleNextQuestion}
                className="mt-6 w-full bg-purple-600 text-white px-6 py-3 rounded-lg font-bold hover:bg-purple-700 transition"
              >
                Next Question
              </button>
            )}
          </div>
        )}
      </div>
    </main>
  );
} 