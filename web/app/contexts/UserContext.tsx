"use client";

import React, {
  createContext,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";

import { User } from "@/types/api";

interface UserContextType {
  user: User | null;
  setUser: (user: User | null) => void;
  clearUser: () => void;
}

const UserContext = createContext<UserContextType | null>(null);

const STORAGE_KEY = "quizUser";

export function UserProvider({ children }: { children: ReactNode }) {
  const [user, setUserState] = useState<User | null>(null);

  useEffect(() => {
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      if (stored) {
        setUserState(JSON.parse(stored) as User);
      }
    } catch {
      localStorage.removeItem(STORAGE_KEY);
    }
  }, []);

  const setUser = (next: User | null) => {
    setUserState(next);
    if (next) {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
    } else {
      localStorage.removeItem(STORAGE_KEY);
    }
  };

  const clearUser = () => setUser(null);

  return (
    <UserContext.Provider value={{ user, setUser, clearUser }}>
      {children}
    </UserContext.Provider>
  );
}

export function useUser() {
  const ctx = useContext(UserContext);
  if (!ctx) {
    throw new Error("useUser must be used within a UserProvider");
  }
  return ctx;
}
