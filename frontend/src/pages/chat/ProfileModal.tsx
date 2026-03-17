import React, { useState } from 'react';
import { useAuthStore } from '../../store/useAuthStore';
import { userApi } from '../../api/user.api';
import { getApiBaseUrl } from '../../utils/runtime';
import { X, User, Camera, Mail, Save, AlertCircle, CheckCircle2, Hash, Loader2, Copy } from 'lucide-react';

interface ProfileModalProps {
  onClose: () => void;
}

export const ProfileModal: React.FC<ProfileModalProps> = ({ onClose }) => {
  const apiBaseUrl = getApiBaseUrl();
  const { user, updateUser } = useAuthStore();
  const [nickName, setNickName] = useState(user?.nickName || '');
  const [signature, setSignature] = useState(user?.signature || '');
  const [gender, setGender] = useState<number | undefined>(user?.gender);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [avatarUrl, setAvatarUrl] = useState(user?.avatar || '');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!nickName.trim()) return;

    setLoading(true);
    setError(null);
    setSuccess(false);

    try {
      const resp = await userApi.updateInfo({
        nickname: nickName.trim(),
        avatar: avatarUrl || '',
        signature,
        gender,
      });

      if (resp.code === 200 || resp.code === 0) {
        // 更新本地 store
        updateUser({ ...user!, nickName: nickName.trim(), signature, gender, avatar: avatarUrl });
        setSuccess(true);
        setTimeout(() => setSuccess(false), 3000);
      } else {
        setError(resp.message || '更新失败');
      }
    } catch (err: any) {
      setError(typeof err === 'string' ? err : (err?.message || '网络错误'));
    } finally {
      setLoading(false);
    }
  };

  const fullAvatarUrl = React.useMemo(() => {
    if (!avatarUrl) return '';
    if (avatarUrl.startsWith('http')) return avatarUrl;
    const baseUrl = apiBaseUrl;
    return `${baseUrl}${avatarUrl.startsWith('/') ? '' : '/'}${avatarUrl}`;
  }, [avatarUrl, apiBaseUrl]);

  return (
    <div className="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-2xl w-full max-w-md shadow-2xl animate-in fade-in zoom-in duration-200">
        <div className="flex justify-between items-center p-6 border-b border-gray-100">
          <h2 className="text-xl font-bold text-gray-900">个人设置</h2>
          <button onClick={onClose} className="p-2 hover:bg-gray-100 rounded-full transition-colors text-gray-400">
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="p-6">
          <div className="flex flex-col items-center mb-8">
            <div className="relative group">
              {avatarUrl ? (
                <img
                  src={fullAvatarUrl}
                  alt="avatar"
                  className="w-24 h-24 rounded-3xl object-cover shadow-lg border border-white/60"
                />
              ) : (
                <div className="w-24 h-24 bg-primary text-white rounded-3xl flex items-center justify-center text-4xl font-bold shadow-lg">
                  {user?.nickName?.[0] || 'U'}
                </div>
              )}
              <label className="absolute -bottom-1 -right-1 p-2 bg-white rounded-xl shadow-md text-gray-400 hover:text-primary transition-colors border border-gray-100 cursor-pointer">
                {uploading ? <Loader2 className="w-4 h-4 animate-spin" /> : <Camera className="w-4 h-4" />}
                <input
                  type="file"
                  className="hidden"
                  accept="image/*"
                  onChange={async (e) => {
                    const file = e.target.files?.[0];
                    if (!file) return;
                    setUploading(true);
                    setError(null);
                    try {
                      const resp = await userApi.uploadAvatar(file);
                      if (resp.code === 200 || resp.code === 0) {
                        // 后端上传时已直接写入数据库，用后端返回的真实URL更新store
                        const newAvatar = resp.data?.url || '';
                        setAvatarUrl(newAvatar);
                        updateUser({ ...user!, avatar: newAvatar });
                        setSuccess(true);
                        setTimeout(() => setSuccess(false), 3000);
                      } else {
                        setError(resp.message || '上传失败');
                      }
                    } catch (err: any) {
                      setError(err?.message || '上传失败');
                    } finally {
                      setUploading(false);
                    }
                  }}
                />
              </label>
            </div>
            <p className="mt-3 text-xs text-gray-400">点击图标修改头像</p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-5">
            {error && (
              <div className="bg-red-50 text-red-600 p-3 rounded-xl flex items-center gap-2 text-sm border border-red-100">
                <AlertCircle className="w-4 h-4" />
                {error}
              </div>
            )}

            {success && (
              <div className="bg-green-50 text-green-600 p-3 rounded-xl flex items-center gap-2 text-sm border border-green-100">
                <CheckCircle2 className="w-4 h-4" />
                个人资料更新成功！
              </div>
            )}

            <div className="space-y-4">
              <div className="space-y-1.5">
                <label className="text-sm font-medium text-gray-700 flex items-center gap-2">
                  <User className="w-4 h-4 text-gray-400" />
                  昵称
                </label>
                <input 
                  value={nickName}
                  onChange={(e) => setNickName(e.target.value)}
                  className="w-full px-4 py-2 bg-gray-50 border border-gray-100 rounded-lg focus:ring-2 focus:ring-primary/20 focus:bg-white outline-none transition-all"
                  placeholder="给自己起个好听的名字"
                />
              </div>

              <div className="space-y-1.5">
                <label className="text-sm font-medium text-gray-700 flex items-center gap-2">
                  个性签名
                </label>
                <input
                  value={signature}
                  onChange={(e) => setSignature(e.target.value)}
                  className="w-full px-4 py-2 bg-gray-50 border border-gray-100 rounded-lg focus:ring-2 focus:ring-primary/20 focus:bg-white outline-none transition-all"
                  placeholder="写点什么吧"
                />
              </div>

              <div className="space-y-1.5">
                <label className="text-sm font-medium text-gray-700 flex items-center gap-2">
                  性别
                </label>
                <div className="flex gap-2">
                  {[{ v: 1, label: '男' }, { v: 0, label: '女' }].map((g) => (
                    <button
                      key={g.v}
                      type="button"
                      onClick={() => setGender(g.v)}
                      className={`px-3 py-1.5 rounded-lg border text-sm ${
                        gender === g.v ? 'border-primary text-primary bg-blue-50' : 'border-gray-200 text-gray-600'
                      }`}
                    >
                      {g.label}
                    </button>
                  ))}
                </div>
              </div>

              <div className="space-y-1.5">
                <label className="text-sm font-medium text-gray-700 flex items-center gap-2">
                  <Mail className="w-4 h-4 text-gray-400" />
                  电子邮箱
                </label>
                <div className="w-full px-4 py-2 bg-gray-100 border border-gray-100 rounded-lg text-gray-500 text-sm cursor-not-allowed">
                  {user?.email}
                </div>
              </div>

              <div className="space-y-1.5">
                <label className="text-sm font-medium text-gray-700 flex items-center gap-2">
                  <Hash className="w-4 h-4 text-gray-400" />
                  用户 ID
                </label>
                <div className="w-full px-4 py-2 bg-gray-100 border border-gray-100 rounded-lg text-gray-500 text-sm font-mono flex items-center justify-between">
                  <span className="truncate">{user?.userId}</span>
                  <button
                    type="button"
                    className="text-gray-400 hover:text-primary flex items-center gap-1 text-xs"
                    onClick={() => navigator.clipboard?.writeText(user?.userId || '')}
                  >
                    <Copy className="w-3 h-3" /> 复制
                  </button>
                </div>
              </div>
            </div>

            <div className="pt-4">
              <button 
                type="submit"
                disabled={loading || !nickName.trim() || (nickName === user?.nickName && avatarUrl === (user?.avatar || '') && signature === (user?.signature || '') && gender === user?.gender)}
                className="w-full py-3 bg-primary text-white rounded-xl hover:bg-blue-600 font-bold transition-all shadow-lg flex items-center justify-center gap-2 disabled:opacity-50 disabled:shadow-none"
              >
                {loading ? '正在保存...' : (
                  <>
                    <Save className="w-4 h-4" />
                    保存修改
                  </>
                )}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
};
