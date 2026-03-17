import apiClient from './client';
import type { ApiResponse } from './types';

export interface AdminUserItem {
  uuid: string;
  nickname: string;
  telephone: string;
  email: string;
  is_admin: boolean;
  status: number;
  avatar?: string;
  created_at?: string;
  birthday?: string;
  gender?: number;
  signature?: string;
}

export const adminApi = {
  getUserList: () =>
    apiClient.get<any, ApiResponse<AdminUserItem[]>>('/v1/admin/user_list'),

  ableUser: (targetUserId: string) =>
    apiClient.post<any, ApiResponse<null>>(`/v1/admin/able_user?target_user_id=${targetUserId}`),

  disableUser: (targetUserId: string) =>
    apiClient.post<any, ApiResponse<null>>(`/v1/admin/disable_user?target_user_id=${targetUserId}`),

  deleteUser: (targetUserId: string) =>
    apiClient.delete<any, ApiResponse<null>>(`/v1/admin/delete_user?target_user_id=${targetUserId}`),

  setAdmin: (targetUserId: string) =>
    apiClient.post<any, ApiResponse<null>>(`/v1/admin/set_admin?target_user_id=${targetUserId}`),
};
