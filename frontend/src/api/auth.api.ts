import apiClient from './client';
import type { ApiResponse, UserInfo } from './types';
import type { LoginForm } from '../pages/auth/LoginPage';

export interface RegisterRequest {
  nickname: string;
  telephone: string;
  email: string;
  password: string;
  code: string;
}

export const authApi = {
  // 邮箱/密码登录
  login: (data: LoginForm) => 
    apiClient.post<any, ApiResponse<UserInfo>>('/v1/login', data),

  // 验证码登录
  loginByCode: (data: { email: string; code: string }) =>
    apiClient.post<any, ApiResponse<UserInfo>>('/v1/login_by_code', data),

  // 注册
  register: (data: RegisterRequest) =>
    apiClient.post<any, ApiResponse<UserInfo>>('/v1/register', data),

  // 发送邮箱验证码
  sendEmailCode: (email: string) =>
    apiClient.post<any, ApiResponse<null>>('/v1/send_email_code', { email }),
};
