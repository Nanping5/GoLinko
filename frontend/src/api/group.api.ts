import apiClient from './client';
import type { ApiResponse } from './types';

export interface GroupInfo {
  group_id: string;
  group_name: string;
  notice: string;
  created_at: string;
  owner_id: string;
  status: number;
}

export interface GroupMember {
  user_id: string;
  nickname: string;
  avatar?: string;
  role: number;
  join_time: string;
}

// GroupMemberInfo 是 GetGroupMembers 接口实际返回的成员信息
export interface GroupMemberInfo {
  user_id: string;
  nickname: string;
  avatar?: string;
  role?: number; // 1 群主, 0 普通成员
}

export const groupApi = {
  // 创建群组
  createGroup: (data: { group_name: string; notice?: string; add_mode: number; avatar?: string }) =>
    apiClient.post<any, ApiResponse<GroupInfo>>('/v1/create_group', data),

  // 加载我的群组
  loadMyGroups: () =>
    apiClient.get<any, ApiResponse<GroupInfo[]>>('/v1/load_my_groups'),

  // 加载我加入的群组（排除自己创建的）
  loadMyJoinedGroups: () =>
    apiClient.get<any, ApiResponse<GroupInfo[]>>('/v1/load_my_joined_groups'),

  // 获取群信息
  getGroupInfo: (groupId: string) =>
    apiClient.get<any, ApiResponse<GroupInfo>>(`/v1/get_group_info?group_id=${groupId}`),

  // 检查加入群组模式（通常用于申请入群前的校验）
  checkGroupAddMode: (groupId: string) =>
    apiClient.get<any, ApiResponse<{ add_mode: number }>>(`/v1/check_group_add_mode?group_id=${groupId}`),

  // 直接加入群组
  enterGroupDirectly: (groupId: string) =>
    apiClient.post<any, ApiResponse<null>>(`/v1/enter_group_directly?group_id=${groupId}`),

  // 更新群组资料
  updateGroupInfo: (data: { group_id: string; group_name?: string; group_notice?: string }) =>
    apiClient.put<any, ApiResponse<null>>('/v1/update_group_info', data),

  // 获取群成员列表
  getGroupMembers: (groupId: string) =>
    apiClient.get<any, ApiResponse<GroupMemberInfo[]>>(`/v1/get_group_members?group_id=${groupId}`),

  // 移除群成员（踢人）
  removeGroupMember: (groupId: string, userId: string) =>
    apiClient.delete<any, ApiResponse<null>>(`/v1/remove_group_member?group_id=${groupId}&user_id=${userId}`),

  // 解散群组
  dismissGroup: (groupId: string) =>
    apiClient.post<any, ApiResponse<null>>('/v1/dismiss_group', { group_id: groupId }),

  // 退出群组
  leaveGroup: (groupId: string) =>
    apiClient.post<any, ApiResponse<null>>('/v1/leave_group', { group_id: groupId }),
};
