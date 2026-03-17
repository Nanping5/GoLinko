import apiClient from './client';
import type { ApiResponse } from './types';

// 仅类型别名（无运行时代码），以适配 isolatedModules/erasableSyntaxOnly
export type ContactType = 0 | 1;
export type ContactStatus = 0 | 2 | 3 | 4 | 5 | 6 | 7;
export type GroupAddMode = 0 | 1 | 2;

// 简单联系人信息（用于列表展示）
export interface ContactUserInfo {
  user_id: string;
  nickname: string;
  avatar?: string;
}

// 详细联系人信息
export interface ContactInfo {
  contact_id: string;
  contact_type: ContactType;
  nickname: string;
  avatar?: string;
  status: ContactStatus;
  // 群组特有字段
  group_notice?: string;
  group_owner_id?: string;
  add_mode?: GroupAddMode;
  create_at?: string;
}

// 好友申请信息
export interface ContactApply {
  apply_id: string;
  applicant_id: string;
  applicant: string;
  avatar?: string;
  message?: string;
  contact_type: ContactType;
  target_name?: string;
  created_at: string;
}

export const contactApi = {
  // ========== 好友列表与详情 ==========

  // 获取联系人列表（好友）
  getContactUserList: () =>
    apiClient.get<any, ApiResponse<ContactUserInfo[]>>('/v1/contact_user_list'),

  // 获取联系人详情（用户或群组）
  getContactInfo: (contactId: string) =>
    apiClient.get<any, ApiResponse<ContactInfo>>(`/v1/contact_info?contact_id=${contactId}`),

  // ========== 关系维护 ==========

  // 拉黑联系人
  blackContact: (contactId: string) =>
    apiClient.post<any, ApiResponse<null>>(`/v1/black_contact?contact_id=${contactId}`),

  // 解除拉黑
  unblackContact: (contactId: string) =>
    apiClient.post<any, ApiResponse<null>>(`/v1/unblack_contact?contact_id=${contactId}`),

  // 删除联系人
  deleteContact: (contactId: string) =>
    apiClient.delete<any, ApiResponse<null>>(`/v1/delete_contact?contact_id=${contactId}`),

  // ========== 申请管理 ==========

  // 申请添加联系人/加入群聊
  applyAddContact: (contactId: string, message?: string) =>
    apiClient.post<any, ApiResponse<null>>('/v1/apply_add_contact', {
      contact_id: contactId,
      message
    }),

  // 获取申请列表（待处理）
  getContactApplyList: () =>
    apiClient.get<any, ApiResponse<ContactApply[]>>('/v1/contact_apply_list'),

  // 同意申请
  acceptContactApply: (applyId: string) =>
    apiClient.post<any, ApiResponse<null>>(`/v1/accept_contact_apply?apply_id=${applyId}`),

  // 拒绝申请
  rejectContactApply: (applyId: string) =>
    apiClient.post<any, ApiResponse<null>>(`/v1/reject_contact_apply?apply_id=${applyId}`),
};
