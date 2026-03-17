import apiClient from './client';
import type { ApiResponse, UserInfo } from './types';

export const userApi = {
  // 获取当前用户信息
  getMyInfo: () =>
    apiClient.get<any, ApiResponse<UserInfo>>('/v1/get_user_info'),

  // 更新用户信息
  updateInfo: (data: Partial<UserInfo>) =>
    apiClient.put<any, ApiResponse<UserInfo>>('/v1/user_info', data),

  // 根据 ID 获取用户信息 (后端 get_user_info 通常从 context 取，如果查询他人通常需要 query 参数)
  getUserById: (userId: string) =>
    apiClient.get<any, ApiResponse<UserInfo>>(`/v1/get_user_info?user_id=${userId}`),

  // 上传头像
  uploadAvatar: (file: File) => {
    const formData = new FormData();
    formData.append('file', file);
    return apiClient.post<any, ApiResponse<{ url: string }>>(
      '/v1/upload_avatar',
      formData,
      {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      }
    );
  },
};
