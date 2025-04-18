import { User, CreateRoomResponse, ApiResponse } from "@/types/api";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export async function createUser(name: string): Promise<ApiResponse<User>> {
  try {
    const response = await fetch(`${API_BASE_URL}/api/users`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ name }),
    });

    if (!response.ok) {
      throw new Error("Failed to create user");
    }

    const data = await response.json();

    return { data };
  } catch (error) {
    return {
      data: null as any,
      error: error instanceof Error ? error.message : "Unknown error",
    };
  }
}

export async function createRoom(
  name: string,
  hostId: string,
): Promise<ApiResponse<CreateRoomResponse>> {
  try {
    const response = await fetch(`${API_BASE_URL}/api/rooms`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ name, hostId }),
    });

    if (!response.ok) {
      throw new Error("Failed to create room");
    }

    const data = await response.json();

    return { data };
  } catch (error) {
    return {
      data: null as any,
      error: error instanceof Error ? error.message : "Unknown error",
    };
  }
}
