export interface ApiResponse<T = any> {
  code: number;
  message: string;
  data: T;
}

export interface UserInfo {
  user_id: string;
  nickname: string;
  email: string;
  telephone?: string;
  avatar?: string;
  gender?: number;
  signature?: string;
  birthday?: string;
  is_admin?: number;
  status?: number;
  create_at?: string;
  token?: string;
}

export interface SessionInfo {
  session_id: string;
  receive_id: string;
  receive_name: string;
  avatar?: string;
}

export interface GroupSessionInfo {
  session_id: string;
  group_id: string;
  group_name: string;
  avatar?: string;
}

export interface MessageInfo {
  send_id: string;
  send_name?: string;
  send_avatar?: string;
  receive_id: string;
  receive_name?: string;
  receive_avatar?: string;
  content?: string;
  url?: string;
  type: number;
  file_type?: string;
  file_name?: string;
  file_size?: number;
  av_data?: string;
}

export type Gender = 0 | 1;
