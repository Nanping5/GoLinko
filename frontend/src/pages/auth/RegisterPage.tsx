import React, { useState, useRef, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { clsx } from 'clsx';
import { Link, useNavigate } from 'react-router-dom';
import { authApi } from '../../api/auth.api';

const registerSchema = z.object({
  email: z.string().email('请输入有效的电子邮箱'),
  nickName: z.string().min(2, '昵称不能少于2位'),
  telephone: z.string().length(11, '手机号必须为11位'),
  code: z.string().length(6, '验证码为6位'),
  password: z.string().min(6, '密码长度不能少于6位'),
  confirmPassword: z.string().min(6, '确认密码长度不能少于6位'),
}).refine((data) => data.password === data.confirmPassword, {
  message: "密码不匹配",
  path: ["confirmPassword"],
});

type RegisterForm = z.infer<typeof registerSchema>;

const CODE_COOLDOWN = 60;

export const RegisterPage: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [sendingCode, setSendingCode] = useState(false);
  const [codeInfo, setCodeInfo] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
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
  } = useForm<RegisterForm>({
    resolver: zodResolver(registerSchema),
  });

  const onSendCode = async () => {
    const email = getValues('email');
    if (!email || !email.includes('@')) {
      setError('请输入有效的邮箱再发送验证码');
      return;
    }
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
    } catch (e: any) {
      setError(e.message || '验证码发送失败');
    } finally {
      setSendingCode(false);
    }
  };

  const onSubmit = async (data: RegisterForm) => {
    setLoading(true);
    setError(null);
    try {
      const resp = await authApi.register({
        nickname: data.nickName,
        email: data.email,
        telephone: data.telephone,
        password: data.password,
        code: data.code
      });

      if (resp.code === 200 || resp.code === 0) {
        navigate('/login');
      } else {
        setError(resp.message);
      }
    } catch (e: any) {
      setError(e.message || '注册失败');
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
        <h2 className="text-xl font-bold text-gray-900 mb-5">创建账号</h2>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-3.5">
          {[
            { label: '昵称', name: 'nickName', placeholder: '给自己起个好听的名字', type: 'text', error: errors.nickName },
            { label: '手机号', name: 'telephone', placeholder: '11位手机号', type: 'text', error: errors.telephone },
          ].map(({ label, name, placeholder, type, error: fieldError }) => (
            <div key={name}>
              <label className="block text-xs font-semibold text-gray-500 uppercase tracking-wide mb-1.5">{label}</label>
              <input
                {...register(name as any)}
                type={type}
                placeholder={placeholder}
                className={clsx(
                  'w-full px-4 py-3 bg-gray-50 border rounded-2xl text-sm outline-none focus:ring-2 transition-all',
                  fieldError ? 'border-red-300 focus:ring-red-100' : 'border-gray-200 focus:ring-[#07c160]/20 focus:border-[#07c160]/40'
                )}
              />
              {fieldError && <p className="mt-1 text-xs text-red-500">{(fieldError as any).message}</p>}
            </div>
          ))}

          <div>
            <label className="block text-xs font-semibold text-gray-500 uppercase tracking-wide mb-1.5">邮箱</label>
            <div className="flex gap-2">
              <input
                {...register('email')}
                placeholder="your@email.com"
                className={clsx(
                  'flex-1 px-4 py-3 bg-gray-50 border rounded-2xl text-sm outline-none focus:ring-2 transition-all',
                  errors.email ? 'border-red-300 focus:ring-red-100' : 'border-gray-200 focus:ring-[#07c160]/20 focus:border-[#07c160]/40'
                )}
              />
              <button
                type="button"
                onClick={onSendCode}
                disabled={sendingCode || countdown > 0}
                className="px-3 py-3 bg-gray-100 text-gray-700 rounded-2xl text-xs font-medium hover:bg-gray-200 disabled:opacity-50 disabled:cursor-not-allowed transition-colors whitespace-nowrap"
              >
                {sendingCode ? '发送中…' : countdown > 0 ? `${countdown}s 后重发` : '获取验证码'}
              </button>
            </div>
            {errors.email && <p className="mt-1 text-xs text-red-500">{errors.email.message}</p>}
          </div>

          <div>
            <label className="block text-xs font-semibold text-gray-500 uppercase tracking-wide mb-1.5">验证码</label>
            <input
              {...register('code')}
              placeholder="6位验证码"
              className={clsx(
                'w-full px-4 py-3 bg-gray-50 border rounded-2xl text-sm outline-none focus:ring-2 transition-all',
                errors.code ? 'border-red-300 focus:ring-red-100' : 'border-gray-200 focus:ring-[#07c160]/20 focus:border-[#07c160]/40'
              )}
            />
            {errors.code && <p className="mt-1 text-xs text-red-500">{errors.code.message}</p>}
          </div>

          <div className="grid grid-cols-2 gap-3">
            {[
              { label: '密码', name: 'password', error: errors.password },
              { label: '确认密码', name: 'confirmPassword', error: errors.confirmPassword },
            ].map(({ label, name, error: fieldError }) => (
              <div key={name}>
                <label className="block text-xs font-semibold text-gray-500 uppercase tracking-wide mb-1.5">{label}</label>
                <input
                  type="password"
                  {...register(name as any)}
                  placeholder="••••••••"
                  className={clsx(
                    'w-full px-4 py-3 bg-gray-50 border rounded-2xl text-sm outline-none focus:ring-2 transition-all',
                    fieldError ? 'border-red-300 focus:ring-red-100' : 'border-gray-200 focus:ring-[#07c160]/20 focus:border-[#07c160]/40'
                  )}
                />
                {fieldError && <p className="mt-1 text-xs text-red-500">{(fieldError as any).message}</p>}
              </div>
            ))}
          </div>

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
            {loading ? '提交中...' : '立即注册'}
          </button>
        </form>

        <p className="text-center text-sm text-gray-400 mt-5">
          已有账号？
          <Link to="/login" className="text-[#07c160] font-medium ml-1 hover:underline">立即登录</Link>
        </p>
      </div>
    </div>
  );
};
