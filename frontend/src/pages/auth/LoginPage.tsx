import React, { useState, useRef, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { clsx } from 'clsx';
import { useAuthStore } from '../../store/useAuthStore';
import { Link, useNavigate } from 'react-router-dom';
import { authApi } from '../../api/auth.api';

const loginSchema = z.object({
  email: z.string().email('请输入有效的电子邮箱'),
  password: z.string().min(6, '密码长度不能少于6位'),
});

const codeLoginSchema = z.object({
  email: z.string().email('请输入有效的电子邮箱'),
  code: z.string().length(6, '验证码为6位'),
});

export type LoginForm = z.infer<typeof loginSchema>;

const CODE_COOLDOWN = 60;

export const LoginPage: React.FC = () => {
  const setAuth = useAuthStore((state) => state.setAuth);
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [codeInfo, setCodeInfo] = useState<string | null>(null);
  const [byCode, setByCode] = useState(false);
  const [sendingCode, setSendingCode] = useState(false);
  const [countdown, setCountdown] = useState(0);
  const countdownRef = useRef<ReturnType<typeof setInterval> | null>(null);

  useEffect(() => {
    return () => { if (countdownRef.current) clearInterval(countdownRef.current); };
  }, []);

  const startCountdown = () => {
    setCountdown(CODE_COOLDOWN);
    countdownRef.current = setInterval(() => {
      setCountdown((prev) => {
        if (prev <= 1) {
          clearInterval(countdownRef.current!);
          countdownRef.current = null;
          return 0;
        }
        return prev - 1;
      });
    }, 1000);
  };

  const {
    register,
    handleSubmit,
    getValues,
    formState: { errors },
  } = useForm<any>({
    resolver: zodResolver(byCode ? (codeLoginSchema as any) : loginSchema),
  });

  const onSubmit = async (data: any) => {
    setLoading(true);
    setError(null);
    setCodeInfo(null);
    try {
      const response = byCode
        ? await authApi.loginByCode({ email: data.email, code: data.code })
        : await authApi.login(data);
      if (response.code === 200 || response.code === 0) {
        const isAdminRaw = (response.data as any).is_admin ?? (response.data as any).isAdmin ?? 0;
        const userInfo = {
          userId: response.data.user_id,
          email: response.data.email,
          nickName: response.data.nickname,
          avatar: response.data.avatar,
          isAdmin: Number(isAdminRaw) === 1,
          status: Number(response.data.status || 0),
        };
        setAuth(userInfo, response.data.token || '');
        navigate('/');
      } else {
        setError(response.message || '登录失败');
      }
    } catch (e: any) {
      setError(e.message || '登录异常，请稍后重试');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 flex flex-col items-center justify-center bg-[#f0f2f5] px-4 py-8 overflow-y-auto">
      {/* Logo */}
      <div className="flex items-center gap-3 mb-8">
        <div className="w-12 h-12 bg-[#07c160] rounded-2xl flex items-center justify-center shadow-lg">
          <span className="text-white font-black text-xl tracking-tighter">CC</span>
        </div>
        <span className="text-2xl font-bold text-gray-900 tracking-tight">GoLinko</span>
      </div>

      <div className="w-full max-w-sm bg-white rounded-3xl shadow-xl border border-gray-100 p-8">
        <h2 className="text-xl font-bold text-gray-900 mb-5">登录账号</h2>

        {/* Mode tabs */}
        <div className="flex gap-4 mb-5 border-b border-gray-100">
          <button
            type="button"
            onClick={() => { setByCode(false); setError(null); setCodeInfo(null); }}
            className={`pb-3 text-sm font-medium border-b-2 -mb-px transition-all ${!byCode ? 'border-[#07c160] text-[#07c160]' : 'border-transparent text-gray-400 hover:text-gray-600'}`}
          >密码登录</button>
          <button
            type="button"
            onClick={() => { setByCode(true); setError(null); setCodeInfo(null); }}
            className={`pb-3 text-sm font-medium border-b-2 -mb-px transition-all ${byCode ? 'border-[#07c160] text-[#07c160]' : 'border-transparent text-gray-400 hover:text-gray-600'}`}
          >验证码登录</button>
        </div>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div>
            <label className="block text-xs font-semibold text-gray-500 uppercase tracking-wide mb-1.5">邮箱</label>
            <input
              {...register('email')}
              placeholder="your@email.com"
              className={clsx(
                'w-full px-4 py-3 bg-gray-50 border rounded-2xl text-sm outline-none focus:ring-2 transition-all',
                errors.email ? 'border-red-300 focus:ring-red-100' : 'border-gray-200 focus:ring-[#07c160]/20 focus:border-[#07c160]/40'
              )}
            />
            {errors.email && <p className="mt-1 text-xs text-red-500">{(errors.email as any)?.message}</p>}
          </div>

          {!byCode ? (
            <div>
              <label className="block text-xs font-semibold text-gray-500 uppercase tracking-wide mb-1.5">密码</label>
              <input
                type="password"
                {...register('password')}
                placeholder="••••••••"
                className={clsx(
                  'w-full px-4 py-3 bg-gray-50 border rounded-2xl text-sm outline-none focus:ring-2 transition-all',
                  errors.password ? 'border-red-300 focus:ring-red-100' : 'border-gray-200 focus:ring-[#07c160]/20 focus:border-[#07c160]/40'
                )}
              />
              {errors.password && <p className="mt-1 text-xs text-red-500">{(errors.password as any)?.message}</p>}
            </div>
          ) : (
            <div>
              <label className="block text-xs font-semibold text-gray-500 uppercase tracking-wide mb-1.5">验证码</label>
              <div className="flex gap-2">
                <input
                  {...register('code')}
                  placeholder="6位验证码"
                  className={clsx(
                    'flex-1 px-4 py-3 bg-gray-50 border rounded-2xl text-sm outline-none focus:ring-2 transition-all',
                    errors.code ? 'border-red-300 focus:ring-red-100' : 'border-gray-200 focus:ring-[#07c160]/20 focus:border-[#07c160]/40'
                  )}
                />
                <button
                  type="button"
                  disabled={sendingCode || countdown > 0}
                  onClick={async () => {
                    const email = getValues('email');
                    if (!email) { setError('请先填写邮箱'); return; }
                    if (countdown > 0) return;
                    setSendingCode(true);
                    setError(null);
                    setCodeInfo(null);
                    try {
                      const resp = await authApi.sendEmailCode(email);
                      if (resp.code === 200 || resp.code === 0) {
                        setCodeInfo('验证码已发送，请查收');
                        startCountdown();
                      } else {
                        setError(resp.message);
                      }
                    } catch (err: any) { setError(err?.message || '发送失败'); }
                    finally { setSendingCode(false); }
                  }}
                  className="px-4 py-3 bg-gray-100 text-gray-700 rounded-2xl text-sm font-medium hover:bg-gray-200 disabled:opacity-50 disabled:cursor-not-allowed transition-colors whitespace-nowrap"
                >
                  {sendingCode ? '发送中…' : countdown > 0 ? `${countdown}s 后重发` : '获取验证码'}
                </button>
              </div>
              {errors.code && <p className="mt-1 text-xs text-red-500">{(errors.code as any)?.message}</p>}
            </div>
          )}

          {codeInfo && (
            <div className="p-3 rounded-2xl text-sm bg-green-50 text-green-600">{codeInfo}</div>
          )}
          {error && (
            <div className="p-3 rounded-2xl text-sm bg-red-50 text-red-500">{error}</div>
          )}

          <button
            type="submit"
            disabled={loading}
            className="w-full py-3.5 bg-[#07c160] text-white rounded-2xl font-semibold text-sm hover:bg-[#06ad55] active:scale-[0.99] disabled:opacity-60 transition-all shadow-lg shadow-emerald-100 mt-2"
          >
            {loading ? '登录中...' : '立即登录'}
          </button>
        </form>

        <p className="text-center text-sm text-gray-400 mt-5">
          还没有账号？
          <Link to="/register" className="text-[#07c160] font-medium ml-1 hover:underline">免费注册</Link>
        </p>
      </div>
    </div>
  );
};
