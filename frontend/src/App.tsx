import React, { Suspense } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import MainLayout from './components/layout/MainLayout';
import { useAuthStore } from './store/useAuthStore';
import { LoginPage, RegisterPage } from './pages/auth';
import { ChatPage } from './pages/chat';

const App: React.FC = () => {
  const token = useAuthStore((state) => state.token);

  return (
    <BrowserRouter>
      <MainLayout>
        <Suspense fallback={<div className="p-8">加载中...</div>}>
          <Routes>
            <Route path="/login" element={!token ? <LoginPage /> : <Navigate to="/" />} />
            <Route path="/register" element={!token ? <RegisterPage /> : <Navigate to="/" />} />
            <Route path="/" element={token ? <ChatPage /> : <Navigate to="/login" />} />
            <Route path="*" element={<Navigate to="/" />} />
          </Routes>
        </Suspense>
      </MainLayout>
    </BrowserRouter>
  );
};

export default App;

