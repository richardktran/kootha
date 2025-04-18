export interface User {
  id: string;
  name: string;
}

export interface CreateRoomResponse {
  roomId: string;
  name: string;
}

export interface ApiResponse<T> {
  data: T;
  error?: string;
} 