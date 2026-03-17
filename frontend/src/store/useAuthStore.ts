import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { User, AuthState } from '../types';

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: (() => {
        try {
          const cached = localStorage.getItem('user');
          return cached ? (JSON.parse(cached) as User) : null;
        } catch {
          return null;
        }
      })(),
      token: localStorage.getItem('token'),
      setAuth: (user, token) => {
        set({ user, token });
        localStorage.setItem('token', token);
        localStorage.setItem('user', JSON.stringify(user));
      },
      updateUser: (user) => {
        set({ user });
        localStorage.setItem('user', JSON.stringify(user));
      },
      logout: () => {
        set({ user: null, token: null });
        localStorage.removeItem('token');
        localStorage.removeItem('user');
      },
    }),
    {
      name: 'auth-storage',
    }
  )
);
