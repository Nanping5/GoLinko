import React, { useState } from 'react';
import { groupApi, type GroupInfo } from '../../api/group.api';
import { X, Search, UserPlus, AlertCircle, CheckCircle2, Users } from 'lucide-react';

interface SearchGroupModalProps {
  onClose: () => void;
  onJoinSuccess: (groupId: string) => void;
}

export const SearchGroupModal: React.FC<SearchGroupModalProps> = ({ onClose, onJoinSuccess }) => {
  const [keyword, setKeyword] = useState('');
  const [result, setResult] = useState<GroupInfo | null>(null);
  const [loading, setLoading] = useState(false);
  const [joining, setJoining] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  const handleSearch = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!keyword.trim()) return;

    setLoading(true);
    setError(null);
    setResult(null);
    setSuccess(false);

    try {
      const resp = await groupApi.getGroupInfo(keyword.trim());
      if (resp.code === 200 || resp.code === 0) {
        if (resp.data) {
          setResult(resp.data);
        } else {
          setError('未找到该群组');
        }
      } else {
        setError(resp.message || '查询失败');
      }
    } catch (err: any) {
      setError(typeof err === 'string' ? err : (err?.message || '群组 ID 不存在'));
    } finally {
      setLoading(false);
    }
  };

  const handleJoin = async () => {
    if (!result || joining) return;

    setJoining(true);
    setError(null);
    try {
      const resp = await groupApi.enterGroupDirectly(result.group_id);
      if (resp.code === 200 || resp.code === 0) {
        setSuccess(true);
        setTimeout(() => {
          onJoinSuccess(result.group_id);
          onClose();
        }, 1500);
      } else {
        setError(resp.message || '加入失败');
      }
    } catch (err: any) {
      setError(typeof err === 'string' ? err : (err?.message || '请求加入失败'));
    } finally {
      setJoining(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-2xl w-full max-w-md shadow-2xl animate-in fade-in zoom-in duration-200">
        <div className="flex justify-between items-center p-6 border-b border-gray-100">
          <h2 className="text-xl font-bold text-gray-900">搜索并加入群组</h2>
          <button onClick={onClose} className="p-2 hover:bg-gray-100 rounded-full transition-colors text-gray-400">
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="p-6 space-y-4">
          <form onSubmit={handleSearch} className="relative">
            <Search className="absolute left-3 top-2.5 text-gray-400 w-4 h-4" />
            <input 
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              placeholder="输入群组 ID"
              className="w-full pl-10 pr-24 py-2 border border-gray-200 rounded-lg focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none transition-all"
            />
            <button 
              type="submit"
              disabled={loading || !keyword.trim()}
              className="absolute right-1.5 top-1.5 px-3 py-1 bg-primary text-white text-xs font-medium rounded-md hover:bg-blue-600 disabled:opacity-50 transition-all"
            >
              {loading ? '搜索中...' : '搜索'}
            </button>
          </form>

          {error && (
            <div className="bg-red-50 text-red-600 p-3 rounded-lg flex items-center gap-2 text-sm border border-red-100">
              <AlertCircle className="w-4 h-4" />
              {error}
            </div>
          )}

          {result && (
            <div className="border border-blue-100 bg-blue-50/30 rounded-xl p-4 space-y-3 animate-in fade-in slide-in-from-top-2">
              <div className="flex items-center gap-4">
                <div className="w-16 h-16 bg-primary text-white rounded-2xl flex items-center justify-center text-2xl font-bold shadow-sm">
                  {result.group_name?.[0] || 'G'}
                </div>
                <div className="flex-1 min-w-0">
                  <h3 className="font-bold text-gray-900 text-lg truncate">{result.group_name}</h3>
                  <p className="text-xs text-gray-500">ID: {result.group_id}</p>
                </div>
              </div>
              
              <div className="text-sm text-gray-600 space-y-1">
                <div className="flex items-start gap-2">
                  <span className="font-medium shrink-0">公告:</span>
                  <p className="line-clamp-2">{result.notice || '暂无公告'}</p>
                </div>
              </div>

              {success ? (
                <div className="flex items-center justify-center gap-2 text-green-600 font-medium py-2 bg-green-50 rounded-lg">
                  <CheckCircle2 className="w-5 h-5" />
                  加入成功！正在进入...
                </div>
              ) : (
                <button 
                  onClick={handleJoin}
                  disabled={joining}
                  className="w-full py-2.5 bg-primary text-white rounded-lg hover:bg-blue-600 font-bold transition-all shadow-md flex items-center justify-center gap-2 disabled:opacity-50"
                >
                  <UserPlus className="w-5 h-5" />
                  {joining ? '正在加入...' : '立即加入群组'}
                </button>
              )}
            </div>
          )}

          {!result && !loading && !error && (
            <div className="py-8 text-center text-gray-400">
              <Users className="w-12 h-12 mx-auto mb-2 opacity-20" />
              <p className="text-sm italic">输入 ID 搜索并加入感兴趣的群组</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
