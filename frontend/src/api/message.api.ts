import apiClient from './client';
import type { ApiResponse } from './types';

export type MessageType = 0 | 1 | 2 | 3;

// 消息信息（与后端字段对齐）
export interface Message {
  message_id?: string;
  send_id: string;
  send_name?: string;
  send_avatar?: string;
  receive_id: string;  // 单聊时为接收者user_id，群聊时为group_id
  group_id?: string;
  content?: string;
  url?: string;
  type: MessageType;
  file_type?: string;
  file_name?: string;
  file_size?: number;
  av_data?: string;
  created_at?: string;
}

// 会话信息（与后端 GetSessionList / GetGroupSessionList 返回对齐）
export interface Session {
  session_id: string;
  receive_id: string;    // 对方user_id或group_id
  receive_name: string;
  avatar?: string;
  group_id?: string;
  group_name?: string;
}

export const messageApi = {
  // ========== 会话管理 ==========

  // 检查是否可以发起会话
  checkOpenSessionAllowed: (receiveId: string) =>
    apiClient.get<any, ApiResponse<{ allowed: boolean }>>(`/v1/check_open_session_allowed?receive_id=${receiveId}`),

  // 打开/创建会话
  openSession: (receiveId: string) =>
    apiClient.post<any, ApiResponse<Session>>(`/v1/open_session?receive_id=${receiveId}`),

  // 获取单聊会话列表
  getSessionList: () =>
    apiClient.get<any, ApiResponse<Session[]>>('/v1/session_list'),

  // 获取群聊会话列表
  getGroupSessionList: () =>
    apiClient.get<any, ApiResponse<Session[]>>('/v1/group_session_list'),

  // 隐藏会话（仅对当前用户）
  hideSession: (sessionId: string) =>
    apiClient.delete<any, ApiResponse<null>>(`/v1/delete_session?session_id=${sessionId}`),

  // 兼容旧调用
  deleteSession: (sessionId: string) =>
    apiClient.delete<any, ApiResponse<null>>(`/v1/delete_session?session_id=${sessionId}`),

  // ========== 消息获取 ==========

  // 获取单聊消息列表
  getMessageList: (sessionId: string) =>
    apiClient.get<any, ApiResponse<Message[]>>(`/v1/message_list?session_id=${sessionId}`),

  // 获取群聊消息列表
  getGroupMessageList: (groupId: string) =>
    apiClient.get<any, ApiResponse<Message[]>>(`/v1/group_message_list?group_id=${groupId}`),

  // ========== 文件上传 ==========

  // 上传文件（用于发送文件消息）
  uploadFile: (file: File) => {
    const formData = new FormData();
    formData.append('file', file);
    return apiClient.post<any, ApiResponse<{ url: string; filename: string }>>(
      '/v1/upload_file',
      formData,
      {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      }
    );
  },
};
