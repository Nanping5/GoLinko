import React, { useEffect, useState } from 'react';
import { groupApi, type GroupMemberInfo } from '../../api/group.api';
import { X, User, Shield, LogOut, AlertCircle, Copy, Trash2 } from 'lucide-react';
import { ConfirmDialog } from '../../components/common/ConfirmDialog';

interface GroupDetailsModalProps {
  groupId: string;
  groupName: string;
  ownerId?: string;
  currentUserId: string;
  onClose: () => void;
  onChanged?: () => void;
}

export const GroupDetailsModal: React.FC<GroupDetailsModalProps> = ({ groupId, groupName, ownerId, currentUserId, onClose, onChanged }) => {
  const [members, setMembers] = useState<GroupMemberInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionLoading, setActionLoading] = useState(false);
  const [toast, setToast] = useState<string | null>(null);
  const [confirm, setConfirm] = useState<{
    title: string; message: string; confirmLabel?: string; danger?: boolean; onConfirm: () => void;
  } | null>(null);

  const showToast = (msg: string) => {
    setToast(msg);
    setTimeout(() => setToast(null), 2500);
  };

  useEffect(() => {
    const fetchMembers = async () => {
      try {
        setLoading(true);
        const resp = await groupApi.getGroupMembers(groupId);
        if (resp.code === 200 || resp.code === 0) {
          setMembers(resp.data || []);
        } else {
          setError(resp.message || '获取成员失败');
        }
      } catch (err: any) {
        setError(typeof err === 'string' ? err : (err?.message || '网络错误'));
      } finally {
        setLoading(false);
      }
    };
    fetchMembers();
  }, [groupId]);

  return (
    <div className="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex justify-end">
      <div className="bg-white w-full max-w-sm h-full shadow-2xl animate-in slide-in-from-right duration-300 flex flex-col">
        {/* Toast */}
        {toast && (
          <div className="fixed top-5 left-1/2 -translate-x-1/2 z-[400] px-5 py-2.5 rounded-2xl text-sm font-medium shadow-xl text-white bg-[#07c160] pointer-events-none">
            {toast}
          </div>
        )}
        {/* Confirm Dialog */}
        {confirm && (
          <ConfirmDialog
            title={confirm.title}
            message={confirm.message}
            confirmLabel={confirm.confirmLabel}
            danger={confirm.danger}
            onConfirm={confirm.onConfirm}
            onCancel={() => setConfirm(null)}
          />
        )}
        {/* Header */}
        <div className="p-6 border-b border-gray-100 flex items-center justify-between">
          <div>
            <h2 className="text-xl font-bold text-gray-900">群组信息</h2>
            <p className="text-xs text-gray-400 mt-1 flex items-center gap-1">
              ID: {groupId}
              <button
                className="text-gray-300 hover:text-primary"
                onClick={() => navigator.clipboard?.writeText(groupId)}
              >
                <Copy className="w-3 h-3" />
              </button>
            </p>
          </div>
          <button onClick={onClose} className="p-2 hover:bg-gray-100 rounded-full transition-colors text-gray-400">
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="flex-1 overflow-y-auto">
          {/* Group Profile Card */}
          <div className="p-6 text-center border-b border-gray-50 bg-gray-50/30">
            <div className="w-20 h-20 bg-primary text-white rounded-3xl flex items-center justify-center text-3xl font-bold shadow-lg mx-auto mb-4">
              {groupName[0] || 'G'}
            </div>
            <h3 className="text-lg font-bold text-gray-900">{groupName}</h3>
          </div>

          {/* Members List Section */}
          <div className="p-6">
          <div className="flex items-center justify-between mb-4">
            <h4 className="font-bold text-gray-900 flex items-center gap-2">
              群成员
              <span className="bg-gray-100 text-gray-500 text-[10px] px-2 py-0.5 rounded-full font-normal">
                {members.length}人
                </span>
              </h4>
            </div>

            {loading ? (
              <div className="space-y-3">
                {[1, 2, 3].map(i => (
                  <div key={i} className="flex gap-3 animate-pulse">
                    <div className="w-10 h-10 bg-gray-100 rounded-lg"></div>
                    <div className="flex-1 space-y-2 py-1">
                      <div className="h-2 bg-gray-100 rounded w-1/2"></div>
                      <div className="h-2 bg-gray-100 rounded w-1/4"></div>
                    </div>
                  </div>
                ))}
              </div>
            ) : error ? (
              <div className="text-center py-4 text-red-400 text-sm flex items-center justify-center gap-2">
                <AlertCircle className="w-4 h-4" />
                {error}
              </div>
            ) : (
              <div className="space-y-3">
                {members.map((member) => (
                  <div key={member.user_id} className="flex items-center gap-3 p-2 hover:bg-gray-50 rounded-xl transition-colors">
                    <div className="w-10 h-10 bg-blue-50 text-blue-600 rounded-lg flex items-center justify-center font-bold">
                      {member.nickname?.[0] || <User size={18} />}
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        <span className="text-sm font-medium text-gray-900 truncate">{member.nickname}</span>
                        {member.role === 1 && (
                          <Shield className="w-3 h-3 text-amber-500" />
                        )}
                      </div>
                      <p className="text-[10px] text-gray-400 flex items-center gap-1">
                        ID: {member.user_id}
                        <button
                          type="button"
                          className="text-gray-300 hover:text-primary"
                          onClick={() => navigator.clipboard?.writeText(member.user_id)}
                        >
                          <Copy className="w-3 h-3" />
                        </button>
                      </p>
                    </div>
                    {ownerId === currentUserId && member.user_id !== currentUserId && (
                      <button
                        className="text-xs text-red-500 px-2 py-1 bg-red-50 rounded-lg hover:bg-red-100"
                        onClick={() => {
                          setConfirm({
                            title: '移出成员',
                            message: `确定将 ${member.nickname} 移出该群？`,
                            confirmLabel: '移出',
                            danger: true,
                            onConfirm: async () => {
                              setConfirm(null);
                              const resp = await groupApi.removeGroupMember(groupId, member.user_id);
                              if (resp.code !== 200 && resp.code !== 0) {
                                showToast(resp.message || '移除失败');
                                return;
                              }
                              setMembers((prev) => prev.filter((m) => m.user_id !== member.user_id));
                              showToast(`已将 ${member.nickname} 移出群组`);
                              onChanged?.();
                            },
                          });
                        }}
                      >
                        移出
                      </button>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>

        {/* Footer Actions */}
        <div className="p-6 border-t border-gray-100 space-y-3">
          <button
            disabled={actionLoading}
            onClick={() => {
              setConfirm({
                title: '退出群组',
                message: `确定退出「${groupName}」？退出后需重新申请才能加入。`,
                confirmLabel: '退出',
                danger: true,
                onConfirm: async () => {
                  setConfirm(null);
                  setActionLoading(true);
                  try {
                    await groupApi.leaveGroup(groupId);
                    showToast('已退出群组');
                    onChanged?.();
                    setTimeout(() => onClose(), 800);
                  } finally {
                    setActionLoading(false);
                  }
                },
              });
            }}
            className="w-full py-3 px-4 bg-gray-50 text-gray-600 rounded-xl hover:bg-red-50 hover:text-red-500 font-medium transition-all flex items-center justify-center gap-2 disabled:opacity-50"
          >
            <LogOut className="w-4 h-4" />
            退出该群组
          </button>
          {ownerId === currentUserId && (
            <button
              disabled={actionLoading}
              onClick={() => {
                setConfirm({
                  title: '解散群组',
                  message: `确定解散「${groupName}」？此操作不可恢复，群内所有消息将被清除。`,
                  confirmLabel: '解散',
                  danger: true,
                  onConfirm: async () => {
                    setConfirm(null);
                    setActionLoading(true);
                    try {
                      await groupApi.dismissGroup(groupId);
                      showToast('群组已解散');
                      onChanged?.();
                      setTimeout(() => onClose(), 800);
                    } finally {
                      setActionLoading(false);
                    }
                  },
                });
              }}
              className="w-full py-3 px-4 bg-red-50 text-red-600 rounded-xl hover:bg-red-100 font-medium transition-all flex items-center justify-center gap-2 disabled:opacity-50"
            >
              <Trash2 className="w-4 h-4" />
              解散群组
            </button>
          )}
        </div>
      </div>
    </div>
  );
};
