export interface User {
  userId: string;
  email: string;
  nickName: string;
  avatar?: string;
  isAdmin?: boolean;
  status?: number;
  telephone?: string;
  gender?: number;
  signature?: string;
  birthday?: string;
}

export interface AuthState {
  user: User | null;
  token: string | null;
  setAuth: (user: User, token: string) => void;
  updateUser: (user: User) => void;
  logout: () => void;
}

export interface ChatMessage {
  sessionId?: string;
  sendId: string;
  sendName?: string;
  sendAvatar?: string;
  receiveId: string;
  content?: string;
  url?: string;
  type?: number;
  fileType?: string;
  fileName?: string;
  fileSize?: number;
  avData?: string;
  createdAt?: string;
}

export type SessionType = 'private' | 'group';

export interface SessionItem {
  sessionId: string;
  receiveId: string;
  name: string;
  avatar?: string;
  type: SessionType;
  lastMessage?: string;
  lastTime?: string;
}
