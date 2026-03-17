import React, { useState } from 'react';
import { groupApi } from '../../api/group.api';
import { X, Save, AlertCircle } from 'lucide-react';

interface CreateGroupModalProps {
  onClose: () => void;
  onSuccess: () => void;
}

export const CreateGroupModal: React.FC<CreateGroupModalProps> = ({ onClose, onSuccess }) => {
  const [groupName, setGroupName] = useState('');
  const [groupNotice, setGroupNotice] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!groupName.trim()) return;

    setLoading(true);
    setError(null);
    try {
      const resp = await groupApi.createGroup({
        group_name: groupName,
        notice: groupNotice,
        add_mode: 1,
      });
      if (resp.code === 200 || resp.code === 0) {
        onSuccess();
        onClose();
      } else {
        setError(resp.message || '创建失败');
      }
    } catch (err: any) {
      setError(typeof err === 'string' ? err : (err?.message || '网络错误'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-2xl w-full max-w-md shadow-2xl animate-in fade-in zoom-in duration-200">
        <div className="flex justify-between items-center p-6 border-b border-gray-100">
          <h2 className="text-xl font-bold text-gray-900">创建新群组</h2>
          <button onClick={onClose} className="p-2 hover:bg-gray-100 rounded-full transition-colors text-gray-400">
            <X className="w-5 h-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          {error && (
            <div className="bg-red-50 text-red-600 p-3 rounded-lg flex items-center gap-2 text-sm border border-red-100">
              <AlertCircle className="w-4 h-4" />
              {error}
            </div>
          )}

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">群组名称 *</label>
            <input 
              required
              value={groupName}
              onChange={(e) => setGroupName(e.target.value)}
              placeholder="起一个响亮的名字"
              className="w-full px-4 py-2 border border-gray-200 rounded-lg focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none transition-all"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">群组公告</label>
            <textarea 
              value={groupNotice}
              onChange={(e) => setGroupNotice(e.target.value)}
              placeholder="简短描述群组用途"
              className="w-full px-4 py-2 border border-gray-200 rounded-lg focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none transition-all resize-none h-24"
            />
          </div>

          <div className="pt-4 flex gap-3">
            <button 
              type="button"
              onClick={onClose}
              className="flex-1 px-4 py-2 bg-gray-50 text-gray-600 rounded-lg hover:bg-gray-100 font-medium transition-colors"
            >
              取消
            </button>
            <button 
              type="submit"
              disabled={loading || !groupName.trim()}
              className="flex-1 px-4 py-2 bg-primary text-white rounded-lg hover:bg-blue-600 font-medium transition-all shadow-md flex items-center justify-center gap-2 disabled:opacity-50"
            >
              {loading ? '创建中...' : (
                <>
                  <Save className="w-4 h-4" />
                  创建
                </>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
